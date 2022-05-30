package authstore

import "github.com/benchkram/bob/bobauth"

type Store interface {
	Contexts() ([]bobauth.AuthContext, error)
	Context(name string) (bobauth.AuthContext, error)
	CreateContext(name, token string) error
	DeleteContext(name string) error
	CurrentContext() (bobauth.AuthContext, error)
	SetCurrentContext(name string) error
	UpdateContext(name, token string) error
}
