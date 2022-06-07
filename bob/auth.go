package bob

import (
	"errors"

	"github.com/benchkram/errz"

	"github.com/benchkram/bob/bobauth"
	"github.com/benchkram/bob/pkg/authstore"
	"github.com/benchkram/bob/pkg/usererror"
)

func (b *B) CreateAuthContext(name, token string) (err error) {
	defer errz.Recover(&err)

	err = b.authStore.CreateContext(name, token)
	if errors.Is(err, authstore.ErrAlreadyExists) {
		return usererror.Wrapm(err, "failed to create authentication context")
	}

	return nil
}

func (b *B) DeleteAuthContext(name string) (err error) {
	defer errz.Recover(&err)

	err = b.authStore.DeleteContext(name)
	if errors.Is(err, authstore.ErrNotFound) {
		return usererror.Wrapm(err, "failed to delete authentication context")
	}

	return nil
}

func (b *B) AuthContexts() ([]bobauth.AuthContext, error) {
	// no usererror needed
	return b.authStore.Contexts()
}

func (b *B) AuthContext(name string) (authCtx bobauth.AuthContext, err error) {
	defer errz.Recover(&err)

	authCtx, err = b.authStore.Context(name)
	if errors.Is(err, authstore.ErrNotFound) {
		return bobauth.AuthContext{}, usererror.Wrapm(err, "failed to retrieve authentication context")
	}
	errz.Fatal(err)

	return authCtx, nil
}

func (b *B) CurrentAuthContext() (curr bobauth.AuthContext, err error) {
	defer errz.Recover(&err)

	curr, err = b.authStore.CurrentContext()
	if errors.Is(err, authstore.ErrNotFound) {
		return bobauth.AuthContext{}, usererror.Wrapm(err, "failed to retrieve current authentication context")
	}
	errz.Fatal(err)

	return curr, nil
}

func (b *B) SetCurrentAuthContext(name string) (err error) {
	defer errz.Recover(&err)

	err = b.authStore.SetCurrentContext(name)
	if errors.Is(err, authstore.ErrNotFound) {
		return usererror.Wrapm(err, "failed to set the current authentication context")
	}

	return nil
}

func (b *B) UpdateAuthContext(name string, token string) (err error) {
	defer errz.Recover(&err)

	err = b.authStore.UpdateContext(name, token)
	if errors.Is(err, authstore.ErrNotFound) {
		return usererror.Wrapm(err, "failed to update authentication context")
	}
	errz.Fatal(err)

	return nil
}
