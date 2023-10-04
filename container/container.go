package container

type ServiceContainer struct {
	Auth      authService
	File      fileService
	Image     imageService
	Product   productService
	Subscribe subscribeService
	User      userService
	Wishlist  wishlistService
}
