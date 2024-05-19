package usecase

import (
	"fmt"
	"github.com/grulex/go-wishlist/pkg/file"
	"net/http"
	"strings"
)

func GetFileUrl(r *http.Request, link file.Link) string {
	if link.StorageType == file.StorageTypeRemoteLink {
		return string(link.ID)
	}
	host := r.Host
	mask := "https://%s/api/images/%s"

	// for local env
	hostPort := strings.Split(r.Host, ":")
	if hostPort[0] == "localhost" || hostPort[0] == "127.0.0.1" {
		mask = "http://%s/api/images/%s"
	}

	linkBase64 := link.Base64()
	return fmt.Sprintf(mask, host, linkBase64)
}
