package container

import (
	authSrv "github.com/grulex/go-wishlist/pkg/auth/service"
	authStore "github.com/grulex/go-wishlist/pkg/auth/storage/postgres"
	fileSrv "github.com/grulex/go-wishlist/pkg/file/service"
	fileStore "github.com/grulex/go-wishlist/pkg/file/storage/inmemory"
	imageSrv "github.com/grulex/go-wishlist/pkg/image/service"
	imageStore "github.com/grulex/go-wishlist/pkg/image/storage/postgres"
	productSrv "github.com/grulex/go-wishlist/pkg/product/service"
	productStore "github.com/grulex/go-wishlist/pkg/product/storage/postgres"
	subscribeSrv "github.com/grulex/go-wishlist/pkg/subscribe/service"
	subscribeStore "github.com/grulex/go-wishlist/pkg/subscribe/storage/postgres"
	userSrv "github.com/grulex/go-wishlist/pkg/user/service"
	userStore "github.com/grulex/go-wishlist/pkg/user/storage/postgres"
	wishlistSrv "github.com/grulex/go-wishlist/pkg/wishlist/service"
	wishlistStore "github.com/grulex/go-wishlist/pkg/wishlist/storage/postgres"
	"github.com/jmoiron/sqlx"
)

type ServiceContainer struct {
	Auth      authService
	File      fileService
	Image     imageService
	Product   productService
	Subscribe subscribeService
	User      userService
	Wishlist  wishlistService
}

func NewServiceContainer(db *sqlx.DB) *ServiceContainer {
	authStorage := authStore.NewAuthStorage(db)
	authService := authSrv.NewAuthService(authStorage)

	fileStorages := make([]fileSrv.FileStorage, 1)
	fileStorages[0] = fileStore.NewFileInMemory()
	fileService := fileSrv.NewFileService(fileStorages)

	imageStorage := imageStore.NewImageStorage(db)
	imageService := imageSrv.NewImageService(imageStorage)

	productStorage := productStore.NewProductStorage(db)
	productService := productSrv.NewProductService(productStorage)

	subscribeStorage := subscribeStore.NewSubscribeStorage(db)
	subscribeService := subscribeSrv.NewSubscribeService(subscribeStorage)

	userStorage := userStore.NewUserStorage(db)
	userService := userSrv.NewUserService(userStorage)

	wishlistStorage := wishlistStore.NewImageStorage(db)
	wishlistService := wishlistSrv.NewWishlistService(wishlistStorage)

	return &ServiceContainer{
		Auth:      authService,
		File:      fileService,
		Image:     imageService,
		Product:   productService,
		Subscribe: subscribeService,
		User:      userService,
		Wishlist:  wishlistService,
	}
}
