package gae

import (
	"time"

	"appengine"
	"appengine/datastore"
)

type UpdateEntry struct {
	Updated time.Time
}

func GetUpdateKey(ctx appengine.Context) *datastore.Key {
	return datastore.NewKey(ctx, "UpdateEntry", "lastupdate", 0, nil)
}

func SetLastUpdateTime(ctx appengine.Context, updated time.Time) error {
	update := &UpdateEntry{Updated: updated}
	key := GetUpdateKey(ctx)
	_, err := datastore.Put(ctx, key, update)
	return err
}

func GetLastUpdateTime(ctx appengine.Context) (time.Time, error) {
	update := new(UpdateEntry)
	key := GetUpdateKey(ctx)
	err := datastore.Get(ctx, key, update)
	if err != nil {
		return time.Time{}, err
	}
	return update.Updated, nil
}

func DeleteLastUpdateTime(ctx appengine.Context) error {
	key := GetUpdateKey(ctx)
	return datastore.Delete(ctx, key)
}
