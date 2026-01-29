package storage

import (
	"log/slog"

	"github.com/dgraph-io/badger/v4"
)

type Store struct {
	db     *badger.DB
	logger *slog.Logger
}

func NewStore(storageLocation string, logger *slog.Logger) (*Store, error) {
	opts := badger.DefaultOptions(storageLocation).WithLogger(nil)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return &Store{
		db:     db,
		logger: logger,
	}, nil
}

func (s *Store) CloseDB() {
	if s.db != nil {
		s.db.Close()
	}
}

func (s *Store) Store(key string, value []byte) error {
	if key == "" {
		return nil
	}

	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), value)
	})
}

func (s *Store) Get(key string) ([]byte, error) {
	var val []byte
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		val, err = item.ValueCopy(nil)
		return err
	})

	return val, err
}
