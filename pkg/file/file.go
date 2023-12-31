package file

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

var ErrStorageNotDefined = errors.New("storage not defined")
var ErrNotFound = errors.New("file not found")

type ID string
type StorageType string

const (
	StorageTypeInMemory    StorageType = "in_memory"
	StorageTypePostgres    StorageType = "postgres"
	StorageTypeTelegramBot StorageType = "telegram_bot"
	StorageTypeRemoteLink  StorageType = "remote_link"
)

const linkBase64Delimiter = ":"

type Link struct {
	StorageType StorageType
	ID          ID
}

func NewLinkFromBase64(base64Str string) (Link, error) {
	decoded, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return Link{}, err
	}

	var storageType StorageType
	var id ID

	decodedParts := strings.Split(string(decoded), linkBase64Delimiter)
	if len(decodedParts) != 2 {
		return Link{}, errors.New("incorrect base64 string")
	}
	storageType = StorageType(decodedParts[0])
	id = ID(decodedParts[1])

	return Link{
		StorageType: storageType,
		ID:          id,
	}, nil
}

func (l *Link) Base64() string {
	str := fmt.Sprintf("%s"+linkBase64Delimiter+"%s", l.StorageType, l.ID)
	return base64.StdEncoding.EncodeToString([]byte(str))
}

type ImageSize struct {
	Width  uint
	Height uint
	Link   Link
}
