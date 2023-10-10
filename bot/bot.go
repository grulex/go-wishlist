package bot

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	wishlistPkg "github.com/grulex/go-wishlist/pkg/wishlist"
	"log"
	"strings"
)

type wishlistService interface {
	Get(ctx context.Context, id wishlistPkg.ID) (*wishlistPkg.Wishlist, error)
}

type TelegramBot struct {
	telegramBot     *tgbotapi.BotAPI
	miniAppUrl      string
	wishlistService wishlistService
}

func NewTelegramBot(token, miniAppUrl string, wService wishlistService) *TelegramBot {
	telegramBot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		panic(err)
	}
	return &TelegramBot{
		telegramBot:     telegramBot,
		miniAppUrl:      miniAppUrl,
		wishlistService: wService,
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
			wishlist, err := s.wishlistService.Get(context.Background(), wishlistPkg.ID(inlineMessage))
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
				button := getButton(" 🎁Create Wishlist!", s.miniAppUrl)
				msg := tgbotapi.NewMessage(update.MyChatMember.Chat.ID,
					"Hello, I'm Wishlist Bot!\n\nI can help you to manage your wishlist.\n\n"+
						"Press \"Create Wishlist!\" or \"My Wishlist\" menu.\n\n"+
						"Also, you can type @"+s.telegramBot.Self.UserName+" and your username in any chat and I'll share your wishlist.")
				msg.ReplyMarkup = &button
				_, err := s.telegramBot.Send(msg)
				if err != nil {
					log.Println(err)
				}
			}
		}

	}
	return nil
}
