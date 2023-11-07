package container

import (
	authSrv "github.com/grulex/go-wishlist/pkg/auth/service"
	authInmemory "github.com/grulex/go-wishlist/pkg/auth/storage/inmemory"
	"github.com/grulex/go-wishlist/pkg/eventmanager/inmemory"
	fileSrv "github.com/grulex/go-wishlist/pkg/file/service"
	fileInmemory "github.com/grulex/go-wishlist/pkg/file/storage/inmemory"
	imageSrv "github.com/grulex/go-wishlist/pkg/image/service"
	imageInmemory "github.com/grulex/go-wishlist/pkg/image/storage/inmemory"
	productSrv "github.com/grulex/go-wishlist/pkg/product/service"
	productInmemory "github.com/grulex/go-wishlist/pkg/product/storage/inmemory"
	subscribeSrv "github.com/grulex/go-wishlist/pkg/subscribe/service"
	subscribeInmemory "github.com/grulex/go-wishlist/pkg/subscribe/storage/inmemory"
	userSrv "github.com/grulex/go-wishlist/pkg/user/service"
	userInmemory "github.com/grulex/go-wishlist/pkg/user/storage/inmemory"
	wishlistSrv "github.com/grulex/go-wishlist/pkg/wishlist/service"
	wishlistInmemory "github.com/grulex/go-wishlist/pkg/wishlist/storage/inmemory"
)

func NewInMemoryServiceContainer() *ServiceContainer {
	eventManager := inmemory.NewEventManager(1000)

	authStorage := authInmemory.NewAuthInMemory()
	authService := authSrv.NewAuthService(authStorage)

	fileStorages := make([]fileSrv.FileStorage, 1)
	fileStorages[0] = fileInmemory.NewFileInMemory()
	fileService := fileSrv.NewFileService(fileStorages)

	imageStorage := imageInmemory.NewImageInMemory()
	imageService := imageSrv.NewImageService(imageStorage)

	productStorage := productInmemory.NewProductInMemory()
	productService := productSrv.NewProductService(productStorage)

	subscribeStorage := subscribeInmemory.NewSubscribeInMemory()
	subscribeService := subscribeSrv.NewSubscribeService(subscribeStorage)

	userStorage := userInmemory.NewUserInMemory()
	userService := userSrv.NewUserService(userStorage)

	wishlistStorage := wishlistInmemory.NewWishlistInMemory()
	wishlistService := wishlistSrv.NewWishlistService(wishlistStorage, eventManager)

	return &ServiceContainer{
		Auth:         authService,
		File:         fileService,
		Image:        imageService,
		Product:      productService,
		Subscribe:    subscribeService,
		User:         userService,
		Wishlist:     wishlistService,
		EventManager: eventManager,
	}
}
