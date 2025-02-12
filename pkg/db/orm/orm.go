// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package orm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"go.etcd.io/bbolt"

	"github.com/defended-net/malwatch/pkg/fsys"
)

// Pair represents a kv record.
type Pair struct {
	Key string
	Val []byte
}

// GetAll returns all of a given bucket's keys and objects.
func GetAll(db *bbolt.DB, bucket string) ([]Pair, error) {
	pairs := []Pair{}

	if err := db.View(func(tx *bbolt.Tx) error {
		bkt := tx.Bucket([]byte(bucket))
		if bkt == nil {
			return fmt.Errorf("%w, %v", ErrTxBktNotFound, bucket)
		}

		return bkt.ForEach(func(key, entry []byte) error {
			result := Pair{
				Key: string(key),
				Val: entry,
			}

			pairs = append(pairs, result)

			return nil
		})
	}); err != nil {
		return nil, fmt.Errorf("%w, %v", ErrBktIter, err)
	}

	return pairs, nil
}

// Get gets a given key's value from given bucket.
func Get(db *bbolt.DB, bucket string, key string) (*Pair, error) {
	pair := &Pair{
		Key: key,
		Val: []byte{},
	}

	if err := db.View(func(tx *bbolt.Tx) error {
		bkt := tx.Bucket([]byte(bucket))
		if bkt == nil {
			return fmt.Errorf("%w, %v", ErrTxBktNotFound, bucket)
		}

		pair.Val = bkt.Get([]byte(key))

		return nil
	}); err != nil {
		return pair, fmt.Errorf("%w, %v, %v", err, bucket, key)
	}

	return pair, nil
}

// Put puts a given key's value from given value.
func Put(db *bbolt.DB, bucket string, key string, obj any) error {
	val, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("%w, %v", ErrMarshal, err)
	}

	return db.Update(func(tx *bbolt.Tx) error {
		bkt := tx.Bucket([]byte(bucket))
		if bkt == nil {
			return fmt.Errorf("%w, %v", ErrTxBktNotFound, bucket)
		}

		if err := bkt.Put([]byte(key), val); err != nil {
			return fmt.Errorf("%w, %v, %v", ErrBktPut, err, key)
		}

		return nil
	})
}

// Mock mocks an orm.
func Mock(path string) (*bbolt.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return nil, fmt.Errorf("%w, %v, %v", fsys.ErrDirCreate, err, path)
	}

	db, err := bbolt.Open(path, 0660, nil)
	if err != nil {
		return nil, err
	}

	if err := db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("hits"))

		return err
	}); err != nil {
		return nil, err
	}

	return db, nil
}
