package bot

import (
	"context"
	"errors"
	"fmt"
	"github.com/corona10/goimagehash"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/grulex/go-wishlist/container"
	authPkg "github.com/grulex/go-wishlist/pkg/auth"
	filePkg "github.com/grulex/go-wishlist/pkg/file"
	imagePkg "github.com/grulex/go-wishlist/pkg/image"
	userPkg "github.com/grulex/go-wishlist/pkg/user"
	wishlistPkg "github.com/grulex/go-wishlist/pkg/wishlist"
	"github.com/jmoiron/sqlx"
	"gopkg.in/guregu/null.v4"
	"image/jpeg"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type wishlistService interface {
	Get(ctx context.Context, id wishlistPkg.ID) (*wishlistPkg.Wishlist, error)
}

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

				button := getButton(" 🎁Create Wishlist!", s.miniAppUrl)
				msg := tgbotapi.NewMessage(update.MyChatMember.Chat.ID,
					"Hello, I'm Wishlist Bot!\n\nI can help you to manage your wishlist.\n\n"+
						"Press \"Create Wishlist!\" button or \"My Wishlist\" menu.\n\n"+
						"Also, you can type @"+s.telegramBot.Self.UserName+" and your username in "+
						"any chat and I'll share your wishlist.")
				msg.ReplyMarkup = &button
				_, err = s.telegramBot.Send(msg)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}
	return nil
}

func (s TelegramBot) checkAndRegisterUser(tgUser tgbotapi.User) error {
	ctx := context.Background()
	userSocialID := authPkg.SocialID(null.NewString(strconv.Itoa(int(tgUser.ID)), true))

	var avatarId *imagePkg.ID
	avatar, _ := s.createAvatarImage(ctx, tgUser.ID)
	if avatar != nil {
		avatarId = &avatar.ID
	}

	auth, err := s.container.Auth.Get(ctx, authPkg.MethodTelegram, userSocialID)
	if err != nil {
		if !errors.Is(err, authPkg.ErrNotFound) {
			return err
		}
	}
	if auth != nil {
		user, err := s.container.User.Get(ctx, auth.UserID)
		if err != nil {
			return err
		}
		wishlists, err := s.container.Wishlist.GetByUserID(ctx, user.ID)
		if err != nil {
			return err
		}
		wishlist := wishlists.GetDefault()
		wishlist.Avatar = avatarId
		err = s.container.Wishlist.Update(ctx, wishlist)
		if err != nil {
			return err
		}

		return nil
	}

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

	newWishlist := &wishlistPkg.Wishlist{
		ID:          wishlistPkg.ID(wishlistId),
		UserID:      user.ID,
		IsDefault:   true,
		Title:       user.FullName + "'s Wishlist",
		Description: "I will be happy to receive any of these gifts!",
		IsArchived:  false,
		Avatar:      avatarId,
	}

	err = s.container.Wishlist.Create(ctx, newWishlist)
	if err != nil {
		return err
	}

	auth = &authPkg.Auth{
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

	middleSizeFile := resp.Photos[0][len(resp.Photos[0])-2:][0]

	fileResp, err := s.telegramBot.GetFile(tgbotapi.FileConfig{
		FileID: middleSizeFile.FileID,
	})
	if resp.TotalCount <= 0 {
		return nil, err
	}

	url := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", s.telegramBot.Token, fileResp.FilePath)
	httpClient := http.Client{}
	httpResp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	// get image from response
	httpImage, err := jpeg.Decode(httpResp.Body)
	if err != nil {
		return nil, err
	}
	_ = httpResp.Body.Close()

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
		FileLink: filePkg.Link{
			StorageType: filePkg.StorageTypeTelegramBot,
			ID:          filePkg.ID(middleSizeFile.FileID),
		},
		Width:  uint(httpImage.Bounds().Dx()),
		Height: uint(httpImage.Bounds().Dy()),
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
