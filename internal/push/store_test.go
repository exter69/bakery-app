package push

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_SaveAndGetByUser(t *testing.T) {
	store := NewStore()

	sub := Subscription{
		ID:       "sub-1",
		UserID:   "user-1",
		Endpoint: "https://push.example.com/sub1",
		P256dh:   "key1",
		Auth:     "auth1",
	}

	store.Save(sub)
	subs := store.GetByUser("user-1")

	require.Len(t, subs, 1)
	assert.Equal(t, sub, subs[0])
}

func TestStore_GetByUser_returnsNilForUnknownUser(t *testing.T) {
	store := NewStore()
	subs := store.GetByUser("nonexistent")
	assert.Nil(t, subs)
}

func TestStore_Save_replacesExistingEndpoint(t *testing.T) {
	store := NewStore()

	sub1 := Subscription{
		ID:       "sub-1",
		UserID:   "user-1",
		Endpoint: "https://push.example.com/sub1",
		P256dh:   "key1",
		Auth:     "auth1",
	}
	sub2 := Subscription{
		ID:       "sub-1-updated",
		UserID:   "user-1",
		Endpoint: "https://push.example.com/sub1", // same endpoint
		P256dh:   "key2-rotated",
		Auth:     "auth2-rotated",
	}

	store.Save(sub1)
	store.Save(sub2)

	subs := store.GetByUser("user-1")
	require.Len(t, subs, 1)
	assert.Equal(t, "key2-rotated", subs[0].P256dh)
	assert.Equal(t, "sub-1-updated", subs[0].ID)
}

func TestStore_Save_multipleDevicesPerUser(t *testing.T) {
	store := NewStore()

	store.Save(Subscription{ID: "sub-1", UserID: "user-1", Endpoint: "https://push.example.com/device1"})
	store.Save(Subscription{ID: "sub-2", UserID: "user-1", Endpoint: "https://push.example.com/device2"})

	subs := store.GetByUser("user-1")
	require.Len(t, subs, 2)
}

func TestStore_Delete_removesSubscription(t *testing.T) {
	store := NewStore()

	store.Save(Subscription{ID: "sub-1", UserID: "user-1", Endpoint: "https://push.example.com/sub1"})
	store.Save(Subscription{ID: "sub-2", UserID: "user-1", Endpoint: "https://push.example.com/sub2"})

	store.Delete("sub-1")

	subs := store.GetByUser("user-1")
	require.Len(t, subs, 1)
	assert.Equal(t, "sub-2", subs[0].ID)
}

func TestStore_Delete_cleansUpEmptyUserSlice(t *testing.T) {
	store := NewStore()

	store.Save(Subscription{ID: "sub-1", UserID: "user-1", Endpoint: "https://push.example.com/sub1"})
	store.Delete("sub-1")

	subs := store.GetByUser("user-1")
	assert.Nil(t, subs)
}

func TestStore_Delete_noopForNonexistentID(t *testing.T) {
	store := NewStore()
	store.Save(Subscription{ID: "sub-1", UserID: "user-1", Endpoint: "https://push.example.com/sub1"})

	store.Delete("nonexistent")

	subs := store.GetByUser("user-1")
	require.Len(t, subs, 1)
}

func TestStore_DeleteByEndpoint(t *testing.T) {
	store := NewStore()

	store.Save(Subscription{ID: "sub-1", UserID: "user-1", Endpoint: "https://push.example.com/sub1"})
	store.Save(Subscription{ID: "sub-2", UserID: "user-1", Endpoint: "https://push.example.com/sub2"})

	store.DeleteByEndpoint("user-1", "https://push.example.com/sub1")

	subs := store.GetByUser("user-1")
	require.Len(t, subs, 1)
	assert.Equal(t, "sub-2", subs[0].ID)
}

func TestStore_ConcurrentAccess(t *testing.T) {
	store := NewStore()
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			store.Save(Subscription{
				ID:       "sub-" + string(rune('a'+idx%26)),
				UserID:   "user-1",
				Endpoint: "https://push.example.com/" + string(rune('a'+idx%26)),
			})
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			store.GetByUser("user-1")
		}()
	}

	wg.Wait()

	// Just verify no panic occurred and store is in a valid state
	subs := store.GetByUser("user-1")
	assert.NotNil(t, subs)
}
