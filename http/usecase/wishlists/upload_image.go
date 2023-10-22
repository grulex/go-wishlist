package wishlists

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/corona10/goimagehash"
	"github.com/grulex/go-wishlist/http/httputil"
	filePkg "github.com/grulex/go-wishlist/pkg/file"
	imagePkg "github.com/grulex/go-wishlist/pkg/image"
	"image"
	_ "image/png"
	"io"
)

type fileService interface {
	UploadPhoto(ctx context.Context, reader io.Reader) (filePkg.Link, error)
}

type imageService interface {
	Create(ctx context.Context, image *imagePkg.Image) error
}

func UploadBase64Image(ctx context.Context, fService fileService, iService imageService, base64image string) (*imagePkg.Image, httputil.HandleResult) {
	decodedSrc, err := base64.StdEncoding.DecodeString(base64image)
	if err != nil {
		return nil, httputil.HandleResult{
			Error: &httputil.HandleError{
				Type:    httputil.ErrorBadData,
				Message: "invalid base64 image",
				Err:     err,
			},
		}
	}
	invalidJpegResult := httputil.HandleResult{
		Error: &httputil.HandleError{
			Type:    httputil.ErrorBadData,
			Message: "invalid jpeg image",
			Err:     err,
		},
	}
	imageObject, format, err := image.Decode(bytes.NewReader(decodedSrc))
	if err != nil {
		fmt.Println(format)
		return nil, invalidJpegResult
	}
	aHash, err := goimagehash.AverageHash(imageObject)
	if err != nil {
		return nil, invalidJpegResult
	}
	pHash, err := goimagehash.PerceptionHash(imageObject)
	if err != nil {
		return nil, invalidJpegResult
	}
	dHash, err := goimagehash.DifferenceHash(imageObject)
	if err != nil {
		return nil, invalidJpegResult
	}

	avatarLink, err := fService.UploadPhoto(ctx, bytes.NewReader(decodedSrc))
	if err != nil {
		return nil, httputil.HandleResult{
			Error: &httputil.HandleError{
				Type:    httputil.ErrorInternal,
				Message: "Error uploading avatar",
				Err:     err,
			},
		}
	}
	newImage := &imagePkg.Image{
		FileLink: avatarLink,
		Width:    uint(imageObject.Bounds().Dx()),
		Height:   uint(imageObject.Bounds().Dy()),
		Hash: imagePkg.Hash{
			AHash: aHash.ToString(),
			DHash: dHash.ToString(),
			PHash: pHash.ToString(),
		},
	}
	err = iService.Create(ctx, newImage)
	if err != nil {
		return nil, httputil.HandleResult{
			Error: &httputil.HandleError{
				Type:    httputil.ErrorInternal,
				Message: "Error creating image",
				Err:     err,
			},
		}
	}

	return newImage, httputil.HandleResult{}
}
