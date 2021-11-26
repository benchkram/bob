package server

import (
	"server-db/server/item"
)

type Application interface {
	item.Service
}

type App struct {
	err error

	items item.Repository
}

func (a *App) Start() error {
	return nil
}

func (a *App) Stop() error {
	return nil
}

func New(opts ...Option) *App {
	a := &App{}

	for _, opt := range opts {
		opt(a)
	}

	return a
}
