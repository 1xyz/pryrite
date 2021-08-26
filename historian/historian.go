package historian

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"time"

	"go.etcd.io/bbolt"
)

type Historian struct {
	Path string
	TTL  time.Duration

	db    *bbolt.DB
	state state
}

type Filter struct {
	Start   *time.Time
	Stop    *time.Time
	Reverse bool
}

const DefaultTTL = 30 * 24 * time.Hour

// Indiate when Each should stop executing and return cleanly.
var ErrStop = errors.New("stop")

//------------------------------------------------------------------------

func Open(path string, readOnly bool) (*Historian, error) {
	opts := &bbolt.Options{
		ReadOnly: readOnly,
	}

	db, err := bbolt.Open(path, 0600, opts)
	if err != nil {
		return nil, err
	}

	h := &Historian{
		Path: path,
		TTL:  DefaultTTL, // FIXME: add TTL to config
		db:   db,
	}

	if readOnly {
		err = h.db.View(func(tx *bbolt.Tx) error {
			return h.loadState(tx)
		})
	} else {
		err = h.db.Update(func(tx *bbolt.Tx) error {
			for _, key := range [][]byte{stateBucketKey, itemsBucketKey} {
				bucket := tx.Bucket(key)
				if bucket == nil {
					_, err = tx.CreateBucket(key)
					if err != nil {
						return wrap(err, "create a bucket")
					}
				}
			}

			err = h.loadState(tx)
			if err != nil {
				return err
			}

			return h.removeOldItems(tx)
		})
	}

	return h, err
}

func (h *Historian) Put(item *Item) error {
	if item.RecordedAt.IsZero() {
		item.RecordedAt = time.Now()
	}

	return h.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(itemsBucketKey)

		if item.ID == 0 {
			var err error
			item.ID, err = bucket.NextSequence()
			if err != nil {
				return wrap(err, "auto-increment an ID")
			}
		}

		mItem, err := json.Marshal(item)
		if err != nil {
			return wrap(err, "marshal an item")
		}

		err = bucket.Put(dateBytes(item.RecordedAt), mItem)
		if err != nil {
			return wrap(err, "save item to bucket")
		}

		return nil
	})
}

func (h *Historian) Get(id uint64) (*Item, error) {
	var found *Item
	err := h.Each(&Filter{Reverse: true}, func(item *Item) error {
		if item.ID == id {
			found = item
			return ErrStop
		}

		return nil
	})

	return found, err
}

func (h *Historian) Each(filter *Filter, cb func(item *Item) error) error {
	return h.EachItem(filter, &Item{}, func(itemI interface{}) error {
		item, ok := itemI.(*Item)
		if !ok {
			return wrap(nil, "cast an item")
		}

		return cb(item)
	})
}

func (h *Historian) EachItem(filter *Filter, exampleItem interface{}, cb func(item interface{}) error) error {
	itemType := reflect.TypeOf(exampleItem).Elem()
	return h.eachEntry(filter, func(line string) error {
		item := reflect.New(itemType).Interface()
		err := json.Unmarshal([]byte(line), item)
		if err != nil {
			return wrap(err, "unmarshal an item")
		}

		return cb(item)
	})
}

func (h *Historian) Len() uint {
	len := 0

	h.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(itemsBucketKey)
		if bucket != nil {
			len = bucket.Stats().KeyN
		}

		return nil
	})

	if len < 0 {
		return 0
	}

	return uint(len)
}

func (h *Historian) Close() error {
	return h.db.Close()
}

//------------------------------------------------------------------------

type state struct {
	RemovedOldItemsAt time.Time
}

var stateBucketKey = []byte("state")
var stateKey = []byte("state")

var itemsBucketKey = []byte("items")

func (h *Historian) loadState(tx *bbolt.Tx) error {
	bucket := tx.Bucket(stateBucketKey)
	data := bucket.Get(stateKey)
	if data != nil {
		err := json.Unmarshal(data, &h.state)
		if err != nil {
			return wrap(err, "unmarshal state")
		}
	}

	return nil
}

func (h *Historian) saveState(tx *bbolt.Tx) error {
	bucket := tx.Bucket(stateBucketKey)
	data, err := json.Marshal(h.state)
	if err != nil {
		return wrap(err, "marshal state")
	}

	return bucket.Put(stateKey, data)
}

func (h *Historian) eachEntry(filter *Filter, cb func(line string) error) error {
	if filter == nil {
		filter = &Filter{}
	}

	return h.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(itemsBucketKey)
		if bucket == nil {
			return nil
		}

		cursor := bucket.Cursor()

		start := cursor.First
		if filter.Start != nil {
			start = func() ([]byte, []byte) {
				return cursor.Seek(dateBytes(*filter.Start))
			}
		} else if filter.Reverse {
			start = cursor.Last
		}

		cont := func([]byte) bool { return true }
		if filter.Stop != nil {
			stop := dateBytes(*filter.Stop)
			if filter.Reverse {
				cont = func(k []byte) bool {
					// continue while the key date is more recent than the stop date
					// (i.e. walking backward in history until the stop)
					return bytes.Compare(k, stop) > 0
				}
			} else {
				cont = func(k []byte) bool {
					// continue while the key date is older than the stop date
					// (i.e. walking forward in history until the stop)
					return bytes.Compare(k, stop) < 0
				}
			}
		}

		next := cursor.Next
		if filter.Reverse {
			next = cursor.Prev
		}

		for k, v := start(); k != nil && cont(k); k, v = next() {
			err := cb(string(v))
			if err == ErrStop {
				break
			}

			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (h *Historian) removeOldItems(tx *bbolt.Tx) error {
	if time.Since(h.state.RemovedOldItemsAt) < 24*time.Hour {
		// only bother cleaning up once a day
		return nil
	}

	bucket := tx.Bucket(itemsBucketKey)
	if bucket == nil {
		return nil
	}

	now := time.Now()
	h.state.RemovedOldItemsAt = now
	oldest := dateBytes(now.Add(-h.TTL))

	cursor := bucket.Cursor()
	for k, _ := cursor.First(); k != nil; k, _ = cursor.Next() {
		if bytes.Compare(k, oldest) < 0 {
			err := bucket.Delete(k)
			if err != nil {
				return err
			}
		} else {
			break
		}
	}

	return h.saveState(tx)
}
