package bob

import "github.com/benchkram/bob/bobauth"

func (b *B) CreateAuthContext(name, token string) error {
	return b.authStore.CreateContext(name, token)
}

func (b *B) DeleteAuthContext(name string) error {
	return b.authStore.DeleteContext(name)
}

func (b *B) AuthContexts() ([]bobauth.AuthContext, error) {
	return b.authStore.Contexts()
}

func (b *B) AuthContext(name string) (bobauth.AuthContext, error) {
	return b.authStore.Context(name)
}

func (b *B) CurrentAuthContext() (bobauth.AuthContext, error) {
	return b.authStore.CurrentContext()
}

func (b *B) SetCurrentAuthContext(name string) error {
	return b.authStore.SetCurrentContext(name)
}

func (b *B) UpdateAuthContext(name string, token string) error {
	return b.authStore.UpdateContext(name, token)
}
