package usecase

import (
	"fmt"
	"github.com/grulex/go-wishlist/pkg/file"
	"net/http"
)

const mask = "https://%s/api/images/%s"

func GetFileUrl(r *http.Request, link file.Link) string {
	host := r.Host
	linkBase64 := link.Base64()
	return fmt.Sprintf(mask, host, linkBase64)
}
