package server

import (
	"server-db/server/item"
)

type Option func(a *App)

func WithItemRepository(repo item.Repository) Option {
	return func(a *App) {
		a.items = repo
	}
}
