package bobsync

import (
	"errors"
)

var (
	ErrCollectionVersionExists  = errors.New("collection version exists on remote")
	ErrSyncPathTaken            = errors.New("sync collection path already in use")
	ErrInvalidCollectionName    = errors.New("invalid collection name")
	ErrInvalidCollectionVersion = errors.New("invalid collection version")
)
