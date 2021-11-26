package server

import (
	"fmt"

	"server-db/server/item"
)

var (
	ErrItemNotFound = fmt.Errorf("item not found")
)

func (a *App) CreateItem(opts ...item.Option) (*item.Item, error) {
	i := item.New(opts...)

	err := a.items.CreateItem(i)
	if err != nil {
		return nil, err
	}

	return &i, nil
}

func (a *App) Item(id string) (*item.Item, error) {
	return a.items.Item(id)
}
