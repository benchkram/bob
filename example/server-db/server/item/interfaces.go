package item

type Repository interface {
	CreateItem(item Item) error
	Item(id string) (*Item, error)
}

type Service interface {
	CreateItem(opts ...Option) (*Item, error)
	Item(id string) (*Item, error)
}
