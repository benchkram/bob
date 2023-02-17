package envutil

type Store map[Hash][]string

func NewStore() Store {
	return make(Store)
}
