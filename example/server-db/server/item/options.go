package item

import (
	"time"
)

type Option func(*Item)

func WithId(id string) Option {
	return func(i *Item) {
		i.id = id
	}
}

func WithName(name string) Option {
	return func(i *Item) {
		i.name = name
	}
}

func WithCreatedAt(createdAt time.Time) Option {
	return func(i *Item) {
		i.createdAt = createdAt
	}
}
