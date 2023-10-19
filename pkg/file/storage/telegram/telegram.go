package telegram

import (
	"context"
	"errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/grulex/go-wishlist/pkg/file"
	"io"
	"net/http"
)

type Storage struct {
	tgBot  *tgbotapi.BotAPI
	chatID int64
}

func NewTelegramStorage(token string, chatID int64) *Storage {
	tgBot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		panic(err)
	}
	return &Storage{
		tgBot:  tgBot,
		chatID: chatID,
	}
}

func (s Storage) GetPhotoReader(ctx context.Context, fileID file.ID) (io.ReadCloser, error) {
	url, err := s.tgBot.GetFileDirectURL(string(fileID))
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	httpClient := http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func (s Storage) UploadPhoto(_ context.Context, reader io.Reader) (file.ID, error) {
	photo := tgbotapi.NewInputMediaPhoto(tgbotapi.FileReader{
		Name:   "wishes.jpg",
		Reader: reader,
	})

	media := tgbotapi.NewMediaGroup(s.chatID, []interface{}{photo})
	media.DisableNotification = true
	mediaMsg, err := s.tgBot.SendMediaGroup(media)
	if err != nil {
		return "", err
	}
	if len(mediaMsg) == 0 {
		return "", errors.New("expected more than one result message, got 0")
	}

	sizes := mediaMsg[0].Photo

	middleSizeFile := sizes[0]
	if len(sizes) >= 3 {
		middleSizeFile = sizes[len(sizes)-2:][0]
	}

	return file.ID(middleSizeFile.FileID), nil
}

func (s Storage) GetStorageType() file.StorageType {
	return file.StorageTypeTelegramBot
}
