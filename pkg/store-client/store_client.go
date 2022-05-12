package storeclient

import (
	"context"
	"io"
	"net/http"

	"github.com/benchkram/bob/pkg/store-client/generated"
)

type I interface {
	UploadArtifact(ctx context.Context, projectName string, artifactID string, src io.Reader) (err error)
	ListArtifacts(ctx context.Context, projectName string) (artifactIds []string, err error)
	GetArtifact(ctx context.Context, projectName string, artifactId string) (rc io.ReadCloser, err error)
}

type c struct {
	endpoint            string
	client              *generated.Client
	clientWithResponses *generated.ClientWithResponses
}

func New(endpoint string, opts ...Option) I {
	c := &c{
		endpoint:            endpoint,
		client:              createClientMust(endpoint),
		clientWithResponses: createClientWithResponsesMust(endpoint),
	}

	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}

	return c
}

func createClientWithResponsesMust(endpoint string) *generated.ClientWithResponses {
	client, err := createClientWithResponses(endpoint)
	if err != nil {
		panic(err)
	}
	return client
}

func createClientWithResponses(endpoint string) (*generated.ClientWithResponses, error) {
	return generated.NewClientWithResponses(endpoint, generated.WithRequestEditorFn(
		func(ctx context.Context, req *http.Request) (err error) {
			return nil
		},
	))
}

func createClientMust(endpoint string) *generated.Client {
	client, err := createClient(endpoint)
	if err != nil {
		panic(err)
	}
	return client
}

func createClient(endpoint string) (*generated.Client, error) {
	return generated.NewClient(endpoint, generated.WithRequestEditorFn(
		func(ctx context.Context, req *http.Request) (err error) {
			return nil
		},
	))
}
