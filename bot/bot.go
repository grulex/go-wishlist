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
		if update.InlineQuery == nil {
			continue
		}

		inlineMessage := update.InlineQuery.Query
		inlineMessage = strings.Trim(inlineMessage, " ")
		if inlineMessage == "" {
			continue
		}
		wishlist, err := s.wishlistService.Get(context.Background(), wishlistPkg.ID(inlineMessage))
		if err != nil {
			continue
		}

		article := tgbotapi.NewInlineQueryResultArticle(update.InlineQuery.ID, wishlist.Title, wishlist.Description)

		button := getButton("Open Wishlist", s.miniAppUrl+"?startapp="+string(wishlist.ID))
		article.ReplyMarkup = &button
		answer := tgbotapi.InlineConfig{
			InlineQueryID: update.InlineQuery.ID,
			Results:       []interface{}{article},
		}
		_, err = s.telegramBot.Send(answer)
		if err != nil {
			log.Println(err)
		}
	}
	return nil
}
