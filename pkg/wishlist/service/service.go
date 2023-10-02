package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/grulex/go-wishlist/pkg/user"
	wishlistPkg "github.com/grulex/go-wishlist/pkg/wishlist"
)

type storage interface {
	Upsert(ctx context.Context, wishlist wishlistPkg.Wishlist) error
	Get(ctx context.Context, id wishlistPkg.ID) (wishlistPkg.Wishlist, error)
	GetByUserID(ctx context.Context, userID user.ID) ([]wishlistPkg.Wishlist, error)
	GetWishlistItems(ctx context.Context, wishlistID wishlistPkg.ID, limit, offset uint) (items []wishlistPkg.Item, haveMore bool, err error)
	GetWishlistItemByID(ctx context.Context, itemID wishlistPkg.ItemID) (wishlistPkg.Item, error)
	UpsertWishlistItem(ctx context.Context, item wishlistPkg.Item) error
	DeleteWishlistItem(ctx context.Context, item wishlistPkg.Item) error
}

type Service struct {
	storage storage
}

func NewWishlistService(storage storage) *Service {
	return &Service{
		storage: storage,
	}
}

func (s *Service) Create(ctx context.Context, wishlist wishlistPkg.Wishlist) error {
	wishlist.ID = wishlistPkg.ID(uuid.NewString())
	return s.storage.Upsert(ctx, wishlist)
}

func (s *Service) Get(ctx context.Context, id wishlistPkg.ID) (wishlistPkg.Wishlist, error) {
	return s.storage.Get(ctx, id)
}

func (s *Service) GetByUserID(ctx context.Context, userID user.ID) ([]wishlistPkg.Wishlist, error) {
	return s.storage.GetByUserID(ctx, userID)
}

func (s *Service) Update(ctx context.Context, wishlist wishlistPkg.Wishlist) error {
	return s.storage.Upsert(ctx, wishlist)
}

func (s *Service) Archive(ctx context.Context, id wishlistPkg.ID) error {
	wishlist, err := s.storage.Get(ctx, id)
	if err != nil {
		return err
	}
	wishlist.IsArchived = true
	return s.storage.Upsert(ctx, wishlist)
}

func (s *Service) Restore(ctx context.Context, id wishlistPkg.ID) error {
	wishlist, err := s.storage.Get(ctx, id)
	if err != nil {
		return err
	}
	wishlist.IsArchived = false
	return s.storage.Upsert(ctx, wishlist)
}

func (s *Service) GetWishlistItems(ctx context.Context, wishlistID wishlistPkg.ID, limit, offset uint) ([]wishlistPkg.Item, bool, error) {
	return s.storage.GetWishlistItems(ctx, wishlistID, limit, offset)
}

func (s *Service) AddWishlistItem(ctx context.Context, item wishlistPkg.Item) error {
	return s.storage.UpsertWishlistItem(ctx, item)
}

func (s *Service) RemoveItem(ctx context.Context, item wishlistPkg.Item) error {
	return s.storage.DeleteWishlistItem(ctx, item)
}

func (s *Service) BookItem(ctx context.Context, itemID wishlistPkg.ItemID, userID user.ID) error {
	item, err := s.storage.GetWishlistItemByID(ctx, itemID)
	if err != nil {
		return err
	}
	if !item.IsBookingAvailable {
		return wishlistPkg.ErrBookingNotAvailable
	}
	if item.IsBookedBy == &userID {
		return nil
	}
	if item.IsBookedBy != nil {
		return wishlistPkg.ErrItemAlreadyBooked
	}
	item.IsBookedBy = &userID
	return s.storage.UpsertWishlistItem(ctx, item)
}

func (s *Service) UnBookItem(ctx context.Context, itemID wishlistPkg.ItemID) error {
	item, err := s.storage.GetWishlistItemByID(ctx, itemID)
	if err != nil {
		return err
	}
	if item.IsBookedBy == nil {
		return nil
	}
	item.IsBookedBy = nil
	return s.storage.UpsertWishlistItem(ctx, item)
}
