package bot

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"github.com/corona10/goimagehash"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/grulex/go-wishlist/container"
	authPkg "github.com/grulex/go-wishlist/pkg/auth"
	"github.com/grulex/go-wishlist/pkg/file"
	imagePkg "github.com/grulex/go-wishlist/pkg/image"
	productPkg "github.com/grulex/go-wishlist/pkg/product"
	userPkg "github.com/grulex/go-wishlist/pkg/user"
	wishlistPkg "github.com/grulex/go-wishlist/pkg/wishlist"
	"github.com/grulex/go-wishlist/scrapper"
	"github.com/jmoiron/sqlx"
	"github.com/mvdan/xurls"
	"gopkg.in/guregu/null.v4"
	"image/jpeg"
	"io"
	"log"
	"net/http"
	urlPkg "net/url"
	"strconv"
	"strings"
	"time"
)

const defaultAvatarImageID = imagePkg.ID("0fc13627-7e95-4bde-ac63-e962969b921a")

type TelegramBot struct {
	telegramBot *tgbotapi.BotAPI
	miniAppUrl  string
	container   *container.ServiceContainer
}

func NewTelegramBot(token, miniAppUrl string, container *container.ServiceContainer) *TelegramBot {
	telegramBot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		if err.Error() == "Not Found" {
			panic("Bot not found. Token is invalid")
		}
		panic(err)
	}
	return &TelegramBot{
		telegramBot: telegramBot,
		miniAppUrl:  miniAppUrl,
		container:   container,
	}
}

func getButton(text, link string) tgbotapi.InlineKeyboardMarkup {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(text, link),
		),
	)
	return kb
}

func (s TelegramBot) Start() error {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	updates := s.telegramBot.GetUpdatesChan(updateConfig)
	for update := range updates {
		if update.InlineQuery != nil {
			inlineMessage := update.InlineQuery.Query
			inlineMessage = strings.Trim(inlineMessage, " @")
			if inlineMessage == "" {
				continue
			}
			wishlist, err := s.container.Wishlist.Get(context.Background(), wishlistPkg.ID(inlineMessage))
			if err != nil {
				continue
			}

			postText := wishlist.Title + "\n\n" + wishlist.Description
			article := tgbotapi.NewInlineQueryResultArticle(update.InlineQuery.ID, wishlist.Title, postText)
			button := getButton("Open Wishlist", s.miniAppUrl+"?startapp="+string(wishlist.ID))
			article.ReplyMarkup = &button
			article.Description = wishlist.Description
			article.ThumbURL = "https://png.pngtree.com/png-vector/20221121/ourmid/pngtree-comicstyle-wishlist-icon-with-splash-effect-health-sign-add-vector-png-image_41870708.jpg"

			answer := tgbotapi.InlineConfig{
				InlineQueryID: update.InlineQuery.ID,
				IsPersonal:    true,
				Results:       []interface{}{article},
			}
			_, err = s.telegramBot.Send(answer)
			if err != nil && err.Error() != "json: cannot unmarshal bool into Go value of type tgbotapi.Message" {
				log.Println(err)
			}
		}

		if update.MyChatMember != nil {
			if update.MyChatMember.NewChatMember.Status == "member" {
				err := s.checkAndRegisterUser(update.MyChatMember.From)
				if err != nil {
					continue
				}

				go func() {
					err := s.checkAvatar(update.MyChatMember.From.ID)
					if err != nil {
						log.Println(err)
					}
				}()

				// waiting for render "/start" message
				time.Sleep(time.Millisecond * 100)

				button := getButton(" 🎁Create Wishlist!", s.miniAppUrl)
				msg := tgbotapi.NewMessage(update.MyChatMember.Chat.ID,
					"Hello, I'm Wishlist Bot!\n\nI can help you to manage your wishlist.\n\n"+
						"Press \"Create Wishlist!\" button or \"My Wishlist\" menu.\n\n"+
						"Also, you can type @"+s.telegramBot.Self.UserName+" and your username in "+
						"any chat and I'll share your wishlist.")
				msg.ReplyMarkup = &button
				msg.DisableNotification = true
				_, err = s.telegramBot.Send(msg)
				if err != nil {
					log.Println(err)
				}
			}
		}

		if update.Message != nil {
			if update.Message.Text == "/start" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID,
					"☝️One more tip!\n I can add a Wish to your List by external link!\n"+
						"Just *share* the link with me, and I'll try to create a wish from it.")
				msg.ParseMode = tgbotapi.ModeMarkdown
				msg.DisableNotification = true
				_, err := s.telegramBot.Send(msg)
				if err != nil {
					log.Println(err)
				}
				continue
			}
			urlsParser := xurls.Relaxed
			urls := urlsParser.FindAllString(update.Message.Text, -1)
			if len(urls) == 0 {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID,
					" I can add a Wish to your List by external link!\n"+
						"Just *share* the link with me, and I'll try to create a wish from it.")
				msg.ParseMode = tgbotapi.ModeMarkdown
				msg.DisableNotification = true
				_, err := s.telegramBot.Send(msg)
				if err != nil {
					log.Println(err)
				}
				continue
			}

			go s.createWishItemsFromUrls(context.Background(), urls, update.Message.From.ID, update.Message.Chat.ID)
		}

	}
	return nil
}

func (s TelegramBot) checkAndRegisterUser(tgUser tgbotapi.User) error {
	ctx := context.Background()
	userSocialID := authPkg.SocialID(null.NewString(strconv.Itoa(int(tgUser.ID)), true))

	auth, err := s.container.Auth.Get(ctx, authPkg.MethodTelegram, userSocialID)
	if err != nil {
		if !errors.Is(err, authPkg.ErrNotFound) {
			return err
		}
	}
	if auth == nil {
		err = s.register(ctx, tgUser)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s TelegramBot) checkAvatar(tgUserID int64) error {
	ctx := context.Background()
	userSocialID := authPkg.SocialID(null.NewString(strconv.Itoa(int(tgUserID)), true))
	auth, err := s.container.Auth.Get(ctx, authPkg.MethodTelegram, userSocialID)
	if err != nil {
		return err
	}

	avatar, _ := s.createAvatarImage(ctx, tgUserID)
	if avatar == nil {
		return nil
	}

	user, err := s.container.User.Get(ctx, auth.UserID)
	if err != nil {
		return err
	}
	wishlists, err := s.container.Wishlist.GetByUserID(ctx, user.ID)
	if err != nil {
		return err
	}
	wishlist := wishlists.GetDefault()
	if wishlist.Avatar == nil {
		wishlist.Avatar = &avatar.ID
		err = s.container.Wishlist.Update(ctx, wishlist)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s TelegramBot) register(ctx context.Context, tgUser tgbotapi.User) error {
	createAuthTransaction, err := s.container.Auth.MakeCreateTransaction(ctx)
	if err != nil {
		return err
	}

	defer func(createAuthTransaction *sqlx.Tx) {
		if createAuthTransaction != nil {
			_ = createAuthTransaction.Rollback()
		}
	}(createAuthTransaction)

	user := &userPkg.User{
		FullName: tgUser.FirstName + " " + tgUser.LastName,
		Language: userPkg.Language(tgUser.LanguageCode),
	}

	err = s.container.User.Create(ctx, user)
	if err != nil {
		return err
	}

	wishlistId := strconv.Itoa(int(tgUser.ID))
	if tgUser.UserName != "" {
		wishlistId = tgUser.UserName
	}

	avatarID := defaultAvatarImageID
	newWishlist := &wishlistPkg.Wishlist{
		ID:          wishlistPkg.ID(wishlistId),
		UserID:      user.ID,
		IsDefault:   true,
		Title:       user.FullName + "'s Wishlist",
		Description: "I will be happy to receive any of these gifts!",
		IsArchived:  false,
		Avatar:      &avatarID,
	}

	err = s.container.Wishlist.Create(ctx, newWishlist)
	if err != nil {
		return err
	}

	userSocialID := authPkg.SocialID(null.NewString(strconv.Itoa(int(tgUser.ID)), true))
	auth := &authPkg.Auth{
		Method:   authPkg.MethodTelegram,
		SocialID: userSocialID,
		UserID:   user.ID,
	}
	err = s.container.Auth.CreateByTransaction(ctx, createAuthTransaction, auth)
	if err != nil {
		return err
	}
	if createAuthTransaction != nil {
		err = createAuthTransaction.Commit()
	}

	return err
}

func (s TelegramBot) createAvatarImage(ctx context.Context, tgUserId int64) (*imagePkg.Image, error) {
	resp, err := s.telegramBot.GetUserProfilePhotos(tgbotapi.UserProfilePhotosConfig{
		UserID: tgUserId,
		Offset: 0,
		Limit:  1,
	})
	if err != nil {
		return nil, err
	}
	if resp.TotalCount <= 0 {
		return nil, nil
	}

	maxSizeFile := resp.Photos[0][len(resp.Photos[0])-1]
	fileUrl, err := s.telegramBot.GetFileDirectURL(maxSizeFile.FileID)
	if err != nil {
		return nil, err
	}

	httpClient := http.Client{}
	httpResp, err := httpClient.Get(fileUrl)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = httpResp.Body
	}(httpResp.Body)

	var body bytes.Buffer
	copyBody := io.TeeReader(httpResp.Body, &body)
	httpImage, err := jpeg.Decode(copyBody)
	if err != nil {
		return nil, err
	}
	fileLink, err := s.container.File.UploadPhoto(ctx, &body)
	if err != nil {
		return nil, err
	}

	aHash, err := goimagehash.AverageHash(httpImage)
	if err != nil {
		return nil, err
	}
	dHash, err := goimagehash.DifferenceHash(httpImage)
	if err != nil {
		return nil, err
	}
	pHash, err := goimagehash.PerceptionHash(httpImage)
	if err != nil {
		return nil, err
	}

	image := &imagePkg.Image{
		FileLink: fileLink,
		Width:    uint(httpImage.Bounds().Dx()),
		Height:   uint(httpImage.Bounds().Dy()),
		Hash: imagePkg.Hash{
			AHash: aHash.ToString(),
			DHash: dHash.ToString(),
			PHash: pHash.ToString(),
		},
	}
	err = s.container.Image.Create(ctx, image)
	if err != nil {
		return nil, err
	}
	return image, nil
}

func (s TelegramBot) sendErrorToChat(chatID int64) {
	msg := tgbotapi.NewMessage(chatID,
		"Sorry, I can't do that now. Please, try again later")
	_, err := s.telegramBot.Send(msg)
	if err != nil {
		log.Println(err)
	}
}

func (s TelegramBot) createWishItemsFromUrls(ctx context.Context, urls []string, tgUserID, chatID int64) {
	userSocialID := authPkg.SocialID(null.NewString(strconv.Itoa(int(tgUserID)), true))
	auth, err := s.container.Auth.Get(ctx, authPkg.MethodTelegram, userSocialID)
	if err != nil {
		s.sendErrorToChat(chatID)
		return
	}
	var wID *wishlistPkg.ID
	if auth != nil {
		user, err := s.container.User.Get(ctx, auth.UserID)
		if err != nil {
			s.sendErrorToChat(chatID)
			return
		}
		wishlists, err := s.container.Wishlist.GetByUserID(ctx, user.ID)
		if err != nil {
			s.sendErrorToChat(chatID)
			return
		}
		wishlist := wishlists.GetDefault()
		if err != nil {
			s.sendErrorToChat(chatID)
			return
		}
		wID = &wishlist.ID
	}

	if wID == nil {
		s.sendErrorToChat(chatID)
		return
	}

	resultProductByUrl := make(map[string]*productPkg.Product)
	for _, url := range urls {
		urlObj, err := urlPkg.Parse(url)
		if urlObj.Scheme == "" {
			urlObj.Scheme = "https"
		}

		linkResult, _ := scrapper.Scrape(urlObj.String(), 5)

		title := ""
		description := ""
		var imageID *imagePkg.ID
		if linkResult != nil {
			title = linkResult.Preview.Title
			description = linkResult.Preview.Description
			if len(linkResult.Preview.Images) != 0 {
				link := file.Link{
					StorageType: file.StorageTypeRemoteLink,
					ID:          file.ID(linkResult.Preview.Images[0]),
				}
				image := &imagePkg.Image{
					FileLink: link,
					Width:    0,
					Height:   0,
					Hash:     imagePkg.Hash{},
				}
				err := s.container.Image.Create(ctx, image)
				if err != nil {
					resultProductByUrl[url] = nil
					continue
				}
				imageID = &image.ID
			}
		}

		if title == "" {
			title = "Wish by attached link"
		}

		titleRunes := []rune(title)
		if len(titleRunes) > productPkg.MaxTitleLength {
			title = string(titleRunes[:productPkg.MaxTitleLength-3]) + "..."
		}

		product := &productPkg.Product{
			Title:       title,
			Description: null.NewString(description, true),
			Url:         null.NewString(urlObj.String(), true),
			ImageID:     imageID,
		}

		err = s.container.Product.Create(ctx, product)
		if err != nil {
			resultProductByUrl[url] = nil
			continue
		}
		item := &wishlistPkg.Item{
			ID: wishlistPkg.ItemID{
				WishlistID: *wID,
				ProductID:  product.ID,
			},
			IsBookingAvailable: true,
		}
		err = s.container.Wishlist.AddWishlistItem(ctx, item)
		if err != nil {
			resultProductByUrl[url] = nil
			continue
		}
		resultProductByUrl[url] = product
	}

	for _, prod := range resultProductByUrl {
		if prod != nil {
			description := "_ <empty> _"
			if prod.Description.String != "" {
				description = prod.Description.String
			}
			msg := tgbotapi.NewMessage(chatID,
				"Wish added to your List!\n\n"+
					"*Title:*\n"+prod.Title+"\n\n"+
					"*Description:*\n"+description+"\n\n"+
					"Open your new [Wish]("+s.makeLinkToItem(*wID, prod.ID)+") to see it.",
			)
			msg.DisableNotification = true
			msg.ParseMode = tgbotapi.ModeMarkdown
			msg.DisableWebPagePreview = true
			_, err := s.telegramBot.Send(msg)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func (s TelegramBot) makeLinkToItem(wishlistID wishlistPkg.ID, productID productPkg.ID) string {
	miniAppInternalRoute := "/wishlists/" + string(wishlistID) + "/items/" + string(productID)

	queryBase64 := base64.StdEncoding.EncodeToString([]byte(miniAppInternalRoute))
	return s.miniAppUrl + "?startapp=-" + queryBase64
}
