package container

import (
	"context"
	authPkg "github.com/grulex/go-wishlist/pkg/auth"
	filePkg "github.com/grulex/go-wishlist/pkg/file"
	imagePkg "github.com/grulex/go-wishlist/pkg/image"
	productPkg "github.com/grulex/go-wishlist/pkg/product"
	subscribePkg "github.com/grulex/go-wishlist/pkg/subscribe"
	userPkg "github.com/grulex/go-wishlist/pkg/user"
	wishlistPkg "github.com/grulex/go-wishlist/pkg/wishlist"
)

type authService interface {
	Get(ctx context.Context, method authPkg.Method, socialID authPkg.SocialID) (*authPkg.Auth, error)
	Create(ctx context.Context, auth *authPkg.Auth) error
}

type fileService interface {
	Upload(ctx context.Context, content []byte, storageType filePkg.StorageType) (*filePkg.Link, error)
	Download(ctx context.Context, link *filePkg.Link) ([]byte, error)
}

type imageService interface {
	Create(ctx context.Context, image *imagePkg.Image) error
	Get(ctx context.Context, id imagePkg.ID) (*imagePkg.Image, error)
}

type productService interface {
	Create(ctx context.Context, product *productPkg.Product) error
	Get(ctx context.Context, id productPkg.ID) (*productPkg.Product, error)
	GetMany(ctx context.Context, ids []productPkg.ID) ([]*productPkg.Product, error)
	Update(ctx context.Context, product *productPkg.Product) error
}

type subscribeService interface {
	Subscribe(ctx context.Context, userID userPkg.ID, wishlistID wishlistPkg.ID) error
	Get(ctx context.Context, userID userPkg.ID, wishlistID wishlistPkg.ID) (*subscribePkg.Subscribe, error)
	GetByUser(ctx context.Context, userID userPkg.ID) ([]*subscribePkg.Subscribe, error)
	Unsubscribe(ctx context.Context, userID userPkg.ID, wishlistID wishlistPkg.ID) error
}

type userService interface {
	Create(ctx context.Context, user *userPkg.User) error
	Get(ctx context.Context, userID userPkg.ID) (*userPkg.User, error)
}

type wishlistService interface {
	Create(ctx context.Context, wishlist *wishlistPkg.Wishlist) error
	Get(ctx context.Context, id wishlistPkg.ID) (*wishlistPkg.Wishlist, error)
	GetByUserID(ctx context.Context, userID userPkg.ID) ([]*wishlistPkg.Wishlist, error)
	Update(ctx context.Context, wishlist *wishlistPkg.Wishlist) error
	Archive(ctx context.Context, id wishlistPkg.ID) error
	Restore(ctx context.Context, id wishlistPkg.ID) error
	GetWishlistItem(ctx context.Context, itemID wishlistPkg.ItemID) (*wishlistPkg.Item, error)
	GetWishlistItems(ctx context.Context, wishlistID wishlistPkg.ID, limit, offset uint) ([]*wishlistPkg.Item, bool, error)
	AddWishlistItem(ctx context.Context, item *wishlistPkg.Item) error
	RemoveItem(ctx context.Context, item wishlistPkg.ItemID) error
	BookItem(ctx context.Context, itemID wishlistPkg.ItemID, userID userPkg.ID) error
	UnBookItem(ctx context.Context, itemID wishlistPkg.ItemID) error
}