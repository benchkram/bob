package bob

import (
	"errors"
	"fmt"

	"github.com/benchkram/errz"

	"github.com/benchkram/bob/pkg/auth"
	"github.com/benchkram/bob/pkg/usererror"
)

func (b *B) CreateAuthContext(name, token string) (err error) {
	defer errz.Recover(&err)

	err = b.authStore.CreateContext(name, token)
	if errors.Is(err, auth.ErrAlreadyExists) {
		return usererror.Wrapm(err, fmt.Sprintf("failed to create authentication context [%s]", name))
	}

	return nil
}

func (b *B) DeleteAuthContext(name string) (err error) {
	defer errz.Recover(&err)

	err = b.authStore.DeleteContext(name)
	if errors.Is(err, auth.ErrNotFound) {
		return usererror.Wrapm(err, fmt.Sprintf("failed to delete authentication context [%s]", name))
	}

	return nil
}

func (b *B) AuthContexts() ([]auth.Context, error) {
	// no usererror needed
	return b.authStore.Contexts()
}

func (b *B) AuthContext(name string) (authCtx auth.Context, err error) {
	defer errz.Recover(&err)

	authCtx, err = b.authStore.Context(name)
	if errors.Is(err, auth.ErrNotFound) {
		return auth.Context{}, usererror.Wrapm(err, fmt.Sprintf("failed to retrieve authentication context [%s]", name))
	}
	errz.Fatal(err)

	return authCtx, nil
}

func (b *B) CurrentAuthContext() (curr auth.Context, err error) {
	defer errz.Recover(&err)

	curr, err = b.authStore.CurrentContext()
	if errors.Is(err, auth.ErrNotFound) {
		return auth.Context{}, usererror.Wrapm(err, "failed to retrieve current authentication context")
	}
	errz.Fatal(err)

	return curr, nil
}

func (b *B) SetCurrentAuthContext(name string) (err error) {
	defer errz.Recover(&err)

	err = b.authStore.SetCurrentContext(name)
	if errors.Is(err, auth.ErrNotFound) {
		return usererror.Wrapm(err, fmt.Sprintf("failed to set current authentication context [%s]", name))
	}

	return nil
}

func (b *B) UpdateAuthContext(name string, token string) (err error) {
	defer errz.Recover(&err)

	err = b.authStore.UpdateContext(name, token)
	if errors.Is(err, auth.ErrNotFound) {
		return usererror.Wrapm(err, fmt.Sprintf("failed to update authentication context [%s]", name))
	}
	errz.Fatal(err)

	return nil
}
