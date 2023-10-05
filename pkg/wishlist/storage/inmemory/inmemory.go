package inmemory

import (
	"context"
	"github.com/grulex/go-wishlist/pkg/user"
	"github.com/grulex/go-wishlist/pkg/wishlist"
	"sync"
)

type Storage struct {
	Wishlists    map[wishlist.ID]*wishlist.Wishlist
	Items        map[wishlist.ID][]*wishlist.Item
	WishlistLock *sync.RWMutex
	ItemsLock    *sync.RWMutex
}

func NewWishlistInMemory() *Storage {
	return &Storage{
		Wishlists:    map[wishlist.ID]*wishlist.Wishlist{},
		Items:        map[wishlist.ID][]*wishlist.Item{},
		WishlistLock: &sync.RWMutex{},
		ItemsLock:    &sync.RWMutex{},
	}
}

func (s *Storage) Upsert(_ context.Context, w *wishlist.Wishlist) error {
	s.WishlistLock.Lock()
	s.Wishlists[w.ID] = w
	s.WishlistLock.Unlock()

	s.ItemsLock.Lock()
	s.Items[w.ID] = []*wishlist.Item{}
	s.ItemsLock.Unlock()

	return nil
}

func (s *Storage) Get(_ context.Context, id wishlist.ID) (*wishlist.Wishlist, error) {
	w, ok := s.Wishlists[id]
	if !ok {
		return nil, wishlist.ErrNotFound
	}
	return w, nil
}

func (s *Storage) GetByUserID(_ context.Context, userID user.ID) ([]*wishlist.Wishlist, error) {
	s.WishlistLock.RLock()
	var wishlists []*wishlist.Wishlist
	for _, w := range s.Wishlists {
		if w.UserID == userID {
			wishlists = append(wishlists, w)
		}
	}
	s.WishlistLock.RUnlock()
	return wishlists, nil
}

func (s *Storage) GetWishlistItems(_ context.Context, wishlistID wishlist.ID, limit, offset uint) (items []*wishlist.Item, haveMore bool, err error) {
	s.ItemsLock.RLock()

	items = s.Items[wishlistID]
	if offset > uint(len(items)) {
		return nil, false, nil
	}
	if offset+limit > uint(len(items)) {
		return items[offset:], false, nil
	}
	s.ItemsLock.RUnlock()
	return items[offset : offset+limit], true, nil
}

func (s *Storage) UpsertWishlistItem(_ context.Context, item *wishlist.Item) error {
	s.ItemsLock.Lock()
	s.Items[item.ID.WishlistID] = append(s.Items[item.ID.WishlistID], item)
	s.ItemsLock.Unlock()
	return nil
}

func (s *Storage) DeleteWishlistItem(_ context.Context, itemID wishlist.ItemID) error {
	s.ItemsLock.Lock()
	var newItems []*wishlist.Item
	for _, i := range s.Items[itemID.WishlistID] {
		if i.ID.ProductID != itemID.ProductID {
			newItems = append(newItems, i)
		}
	}
	s.Items[itemID.WishlistID] = newItems
	s.ItemsLock.Unlock()
	return nil
}

func (s *Storage) GetWishlistItemByID(_ context.Context, itemID wishlist.ItemID) (*wishlist.Item, error) {
	s.ItemsLock.RLock()
	for _, i := range s.Items[itemID.WishlistID] {
		if i.ID.ProductID == itemID.ProductID {
			s.ItemsLock.RUnlock()
			return i, nil
		}
	}
	s.ItemsLock.RUnlock()
	return nil, wishlist.ErrItemNotFound
}
