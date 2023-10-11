package images

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gorilla/mux"
	"github.com/grulex/go-wishlist/http/httputil"
	"github.com/grulex/go-wishlist/pkg/file"
	"net/http"
)

func MakeGetImageFileHandler(telegramBotToken string) httputil.HttpUseCase {
	return func(r *http.Request) httputil.HandleResult {
		vars := mux.Vars(r)
		linkBase64, ok := vars["link_base64"]
		if !ok {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Type:     httputil.ErrorNotFound,
					ErrorKey: "not_found",
					Message:  "incorrect path parameter",
					Err:      nil,
				},
			}
		}
		link, err := file.NewLinkFromBase64(linkBase64)
		if err != nil {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Type:     httputil.ErrorNotFound,
					ErrorKey: "not_found",
					Message:  "incorrect path parameter",
					Err:      nil,
				},
			}
		}
		if link.StorageType != file.StorageTypeTelegramBot {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Type:     httputil.ErrorNotFound,
					ErrorKey: "not_found",
					Message:  "incorrect path parameter",
					Err:      nil,
				},
			}
		}

		telegramBot, err := tgbotapi.NewBotAPI(telegramBotToken)
		if err != nil {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Type:    httputil.ErrorInternal,
					Message: "Error connecting to telegram bot",
				},
			}
		}

		telegramUrl, err := getTelegramFileUrl(telegramBot, string(link.ID))
		if err != nil {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Type:    httputil.ErrorInternal,
					Message: "Error getting file url",
				},
			}
		}

		req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, telegramUrl, nil)
		if err != nil {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Type:    httputil.ErrorInternal,
					Message: "Error creating request",
				},
			}
		}
		httpClient := http.Client{}
		resp, err := httpClient.Do(req)
		if err != nil {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Type:    httputil.ErrorInternal,
					Message: "Error getting file",
				},
			}
		}
		// read body
		//body, err := io.ReadAll(resp.Body)
		//if err != nil {
		//	return httputil.HandleResult{
		//		Error: &httputil.HandleError{
		//			Type:    httputil.ErrorInternal,
		//			Message: "Error reading file",
		//		},
		//	}
		//}
		//_ = resp.Body.Close()

		return httputil.HandleResult{
			Payload: resp.Body,
			Type:    httputil.ResponseTypeJpeg,
		}
	}
}

func getTelegramFileUrl(telegramBot *tgbotapi.BotAPI, fileID string) (string, error) {
	fileResp, err := telegramBot.GetFile(tgbotapi.FileConfig{
		FileID: fileID,
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", telegramBot.Token, fileResp.FilePath), nil
}
