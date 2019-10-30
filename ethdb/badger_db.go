// Copyright 2019 The turbo-geth authors
// This file is part of the turbo-geth library.
//
// The turbo-geth library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The turbo-geth library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the turbo-geth library. If not, see <http://www.gnu.org/licenses/>.

package ethdb

import (
	"github.com/dgraph-io/badger"

	"github.com/ledgerwatch/turbo-geth/log"
)

// BadgerDatabase is a wrapper over BadgerDb,
// compatible with the Database interface.
type BadgerDatabase struct {
	db *badger.DB // BadgerDB instance

	log log.Logger // Contextual logger tracking the database path
}

// NewBadgerDatabase returns a BadgerDB wrapper.
func NewBadgerDatabase(dir string) (*BadgerDatabase, error) {
	logger := log.New("database", dir)

	db, err := badger.Open(badger.DefaultOptions(dir))
	if err != nil {
		return nil, err
	}

	return &BadgerDatabase{
		db:  db,
		log: logger,
	}, nil
}

// Close closes the database.
func (db *BadgerDatabase) Close() {
	if err := db.db.Close(); err == nil {
		db.log.Info("Database closed")
	} else {
		db.log.Error("Failed to close database", "err", err)
	}
}

const bucketSeparator = byte(0xA6) // broken bar '¦'

func bucketKey(bucket, key []byte) []byte {
	var composite []byte
	composite = append(composite, bucket...)
	composite = append(composite, bucketSeparator)
	composite = append(composite, key...)
	return composite
}

// Delete removes a single entry.
func (db *BadgerDatabase) Delete(bucket, key []byte) error {
	err := db.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(bucketKey(bucket, key))
	})
	return err
}

// Put inserts or updates a single entry.
func (db *BadgerDatabase) Put(bucket, key []byte, value []byte) error {
	err := db.db.Update(func(txn *badger.Txn) error {
		return txn.Set(bucketKey(bucket, key), value)
	})
	return err
}

// Get returns a single value.
func (db *BadgerDatabase) Get(bucket, key []byte) ([]byte, error) {
	var val []byte
	err := db.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(bucketKey(bucket, key))
		if err != nil {
			return err
		}
		val, err = item.ValueCopy(nil)
		return err
	})
	return val, err
}

// TODO [Andrew] implement the full Database interface
