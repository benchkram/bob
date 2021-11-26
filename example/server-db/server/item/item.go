package item

import (
	"fmt"
	"github.com/lithammer/shortuuid/v3"
	"time"
)

type Item struct {
	id        string
	name      string
	createdAt time.Time
}

func (i *Item) Id() string {
	return i.id
}

func (i *Item) Name() string {
	return i.name
}

func (i *Item) CreatedAt() time.Time {
	return i.createdAt
}

func (i *Item) String() string {
	return fmt.Sprintf("{id: %q, name: %q, createdAt: %d}", i.id, i.name, i.createdAt.Unix())
}

func New(opts ...Option) Item {
	id := shortuuid.New()

	i := Item{
		id:        id,
		name:      id,
		createdAt: time.Now(),
	}

	for _, opt := range opts {
		opt(&i)
	}

	return i
}
