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
	imagePkg "github.com/grulex/go-wishlist/pkg/image"
	"github.com/grulex/go-wishlist/pkg/notify"
	productPkg "github.com/grulex/go-wishlist/pkg/product"
	userPkg "github.com/grulex/go-wishlist/pkg/user"
	wishlistPkg "github.com/grulex/go-wishlist/pkg/wishlist"
	"github.com/grulex/go-wishlist/scrapper"
	"github.com/grulex/go-wishlist/translate"
	"github.com/jmoiron/sqlx"
	"github.com/mvdan/xurls"
	_ "golang.org/x/image/webp"
	"gopkg.in/guregu/null.v4"
	"image"
	_ "image/jpeg"
	_ "image/png"
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
	translator  *translate.Translator
}

func NewTelegramBot(token, miniAppUrl string, container *container.ServiceContainer) *TelegramBot {
	telegramBot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		if err.Error() == "Not Found" {
			panic("Bot not found. Token is invalid")
		}
		panic(err)
	}
	translator := translate.NewTranslator("en")

	return &TelegramBot{
		telegramBot: telegramBot,
		miniAppUrl:  miniAppUrl,
		container:   container,
		translator:  translator,
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
	ctx := context.Background()
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	updates := s.telegramBot.GetUpdatesChan(updateConfig)
	for update := range updates {
		if update.InlineQuery != nil {
			lang := update.InlineQuery.From.LanguageCode

			inlineMessage := update.InlineQuery.Query
			inlineMessage = strings.Trim(inlineMessage, " @")
			if inlineMessage == "" {
				continue
			}
			wishlist, err := s.container.Wishlist.Get(ctx, wishlistPkg.ID(inlineMessage))
			if err != nil {
				continue
			}

			postText := wishlist.Title + "\n\n" + wishlist.Description
			article := tgbotapi.NewInlineQueryResultArticle(update.InlineQuery.ID, wishlist.Title, postText)
			button := getButton(s.translator.Translate(lang, "open_wishlist")+"!", s.miniAppUrl+"?startapp="+string(wishlist.ID))
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
				// waiting for render "/start" message
				time.Sleep(time.Millisecond * 200)

				err := s.checkAndRegisterUser(ctx, update.MyChatMember.From, update.MyChatMember.Chat)
				if err != nil {
					continue
				}

				tgUserID := update.MyChatMember.From.ID
				tgChatID := update.MyChatMember.Chat.ID
				go func() {
					err := s.checkUpdates(ctx, tgUserID, tgChatID)
					if err != nil {
						log.Println(err)
					}
				}()
			} else if update.MyChatMember.NewChatMember.Status == "kicked" {
				// TODO: remove notify channel
			}
		}

		if update.Message != nil {
			if update.Message.Text == "/stats_week" && update.Message.Chat.ID == 39439763 {
				stats, err := s.container.User.GetDailyStats(ctx, time.Hour*24*7)
				if err != nil {
					log.Println(err)
				}
				msgStr := ""
				for _, stat := range stats {
					msgStr += stat.String() + "\n"
				}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "```\n"+msgStr+"\n```")
				msg.ParseMode = tgbotapi.ModeMarkdown
				_, err = s.telegramBot.Send(msg)
				if err != nil {
					log.Println(err)
				}
				continue
			}
			lang := update.Message.From.LanguageCode
			err := s.checkAndRegisterUser(ctx, *update.Message.From, *update.Message.Chat)
			if err != nil {
				log.Println(err)
				continue
			}
			if update.Message.Text == "/start" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, s.translator.Translate(lang, "tip_1"))
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
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, s.translator.Translate(lang, "tip_2"))
				msg.ParseMode = tgbotapi.ModeMarkdown
				msg.DisableNotification = true
				_, err := s.telegramBot.Send(msg)
				if err != nil {
					log.Println(err)
				}
				continue
			}

			go s.createWishItemsFromUrls(ctx, urls, update.Message.From.ID, update.Message.Chat.ID)
		}

	}
	return nil
}

func (s TelegramBot) checkAndRegisterUser(ctx context.Context, tgUser tgbotapi.User, tgChat tgbotapi.Chat) error {
	userSocialID := authPkg.SocialID(null.NewString(strconv.Itoa(int(tgUser.ID)), true))

	auth, err := s.container.Auth.Get(ctx, authPkg.MethodTelegram, userSocialID)
	if err != nil {
		if !errors.Is(err, authPkg.ErrNotFound) {
			return err
		}
	}
	if auth == nil {
		err = s.register(ctx, tgUser, tgChat)
		if err != nil {
			return err
		}

		button := getButton(" 🎁"+s.translator.Translate(tgUser.LanguageCode, "open_wishlist")+"!", s.miniAppUrl)
		msg := tgbotapi.NewMessage(tgChat.ID, s.translator.Translate(tgUser.LanguageCode, "welcome_message"))
		msg.ReplyMarkup = &button
		msg.DisableNotification = true
		_, err = s.telegramBot.Send(msg)
		if err != nil {
			log.Println(err)
		}

	}

	return nil
}

func (s TelegramBot) checkUpdates(ctx context.Context, tgUserID, tgChatID int64) error {
	userSocialID := authPkg.SocialID(null.NewString(strconv.Itoa(int(tgUserID)), true))
	auth, err := s.container.Auth.Get(ctx, authPkg.MethodTelegram, userSocialID)
	if err != nil {
		return err
	}

	avatar, err := s.createAvatarImage(ctx, tgUserID)
	if err != nil {
		return err
	}
	if avatar == nil {
		return nil
	}

	user, err := s.container.User.Get(ctx, auth.UserID)
	if err != nil {
		return err
	}
	if user.NotifyType == nil {
		tgType := notify.TypeTelegram
		channelID := strconv.Itoa(int(tgChatID))
		user.NotifyType = &tgType
		user.NotifyChannelID = &channelID
		err = s.container.User.Update(ctx, user)
		if err != nil {
			return err
		}
	}
	wishlists, err := s.container.Wishlist.GetByUserID(ctx, user.ID)
	if err != nil {
		return err
	}
	wishlist := wishlists.GetDefault()
	if wishlist.Avatar == nil || (wishlist.Avatar != nil && *wishlist.Avatar == defaultAvatarImageID) {
		wishlist.Avatar = &avatar.ID
		err = s.container.Wishlist.Update(ctx, wishlist)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s TelegramBot) register(ctx context.Context, tgUser tgbotapi.User, tgChat tgbotapi.Chat) error {
	createAuthTransaction, err := s.container.Auth.MakeCreateTransaction(ctx)
	if err != nil {
		return err
	}

	defer func(createAuthTransaction *sqlx.Tx) {
		if createAuthTransaction != nil {
			_ = createAuthTransaction.Rollback()
		}
	}(createAuthTransaction)

	notifyTg := notify.TypeTelegram
	notifyChannel := strconv.Itoa(int(tgChat.ID))
	user := &userPkg.User{
		FullName:        tgUser.FirstName + " " + tgUser.LastName,
		Language:        userPkg.Language(tgUser.LanguageCode),
		NotifyType:      &notifyTg,
		NotifyChannelID: &notifyChannel,
	}

	err = s.container.User.Create(ctx, user)
	if err != nil {
		return err
	}

	wishlistId := strconv.Itoa(int(tgUser.ID))
	if tgUser.UserName != "" {
		wishlistId = tgUser.UserName
	}

	name := tgUser.UserName
	if name == "" {
		name = tgUser.FirstName
	}
	avatarID := defaultAvatarImageID
	newWishlist := &wishlistPkg.Wishlist{
		ID:          wishlistPkg.ID(wishlistId),
		UserID:      user.ID,
		IsDefault:   true,
		Title:       s.translator.Translate(string(user.Language), "wishlist_title") + " — " + name,
		Description: s.translator.Translate(string(user.Language), "init_description"),
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

	return s.createImageFromUrl(ctx, fileUrl)
}

func (s TelegramBot) createImageFromUrl(ctx context.Context, fileUrl string) (*imagePkg.Image, error) {
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
	httpImage, _, err := image.Decode(copyBody)
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

	imgSizes, err := s.container.File.UploadPhoto(ctx, &body)
	if err != nil {
		return nil, err
	}
	sizes := make([]imagePkg.Size, len(imgSizes))
	for i, imgSize := range imgSizes {
		sizes[i] = imagePkg.Size{
			Width:    imgSize.Width,
			Height:   imgSize.Height,
			FileLink: imgSize.Link,
		}
	}

	image := &imagePkg.Image{
		FileLink: imgSizes[len(imgSizes)-1].Link,
		Width:    uint(httpImage.Bounds().Dx()),
		Height:   uint(httpImage.Bounds().Dy()),
		Hash: imagePkg.Hash{
			AHash: aHash.ToString(),
			DHash: dHash.ToString(),
			PHash: pHash.ToString(),
		},
		Sizes: sizes,
	}
	err = s.container.Image.Create(ctx, image)
	if err != nil {
		return nil, err
	}
	return image, nil
}

func (s TelegramBot) sendErrorToChat(err error, chatID int64) {
	log.Println(err)
	msg := tgbotapi.NewMessage(chatID,
		"Sorry, I can't do that now. Please, try again later")
	_, err = s.telegramBot.Send(msg)
	if err != nil {
		log.Println(err)
	}
}

func (s TelegramBot) createWishItemsFromUrls(ctx context.Context, urls []string, tgUserID, chatID int64) {
	userSocialID := authPkg.SocialID(null.NewString(strconv.Itoa(int(tgUserID)), true))
	auth, err := s.container.Auth.Get(ctx, authPkg.MethodTelegram, userSocialID)
	if err != nil {
		s.sendErrorToChat(err, chatID)
		return
	}
	var wID *wishlistPkg.ID
	if auth != nil {
		user, err := s.container.User.Get(ctx, auth.UserID)
		if err != nil {
			s.sendErrorToChat(err, chatID)
			return
		}
		wishlists, err := s.container.Wishlist.GetByUserID(ctx, user.ID)
		if err != nil {
			s.sendErrorToChat(err, chatID)
			return
		}
		wishlist := wishlists.GetDefault()
		if err != nil {
			s.sendErrorToChat(err, chatID)
			return
		}
		wID = &wishlist.ID
	}

	if wID == nil {
		s.sendErrorToChat(errors.New("wishlist not found"), chatID)
		return
	}

	user, err := s.container.User.Get(ctx, auth.UserID)
	if err != nil {
		s.sendErrorToChat(err, chatID)
		return
	}
	lang := string(user.Language)

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
				image, err := s.createImageFromUrl(ctx, linkResult.Preview.Images[0])
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
			description := "_ <" + s.translator.Translate(lang, "empty") + "> _"
			if prod.Description.String != "" {
				description = prod.Description.String
			}
			msg := tgbotapi.NewMessage(chatID,
				s.translator.Translate(lang, "wish_added_pattern", prod.Title, description, s.makeLinkToItem(*wID, prod.ID)))

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
