package storeclient

import (
	"context"
	"net/http"

	"github.com/Benchkram/bob/pkg/store-client/generated"
)

type I interface {
}

type c struct {
	endpoint string
	client   *generated.ClientWithResponses
}

func New(endpoint string, opts ...Option) I {
	c := &c{
		endpoint: endpoint,
		client:   createClientMust(endpoint),
	}

	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}

	return c
}

func createClientMust(endpoint string) *generated.ClientWithResponses {
	client, err := createClient(endpoint)
	if err != nil {
		panic(err)
	}
	return client
}

func createClient(endpoint string) (*generated.ClientWithResponses, error) {
	return generated.NewClientWithResponses(endpoint, generated.WithRequestEditorFn(
		func(ctx context.Context, req *http.Request) (err error) {
			return nil
		},
	))
}
