package storeclient

import (
	"context"
	"io"
	"net/http"

	"github.com/benchkram/bob/pkg/store-client/generated"
)

type I interface {
	UploadArtifact(ctx context.Context, projectName string, artifactID string, src io.Reader, size int64) (err error)
	ListArtifacts(ctx context.Context, projectName string) (artifactIds []string, err error)
	GetArtifact(ctx context.Context, projectName string, artifactId string) (rc io.ReadCloser, size int64, err error)

	CollectionCreate(ctx context.Context, projectName, name, localPath string) (*generated.SyncCollection, error)
	Collection(ctx context.Context, projectName, collectionId string) (*generated.SyncCollection, error)
	Collections(ctx context.Context, projectName string) ([]generated.SyncCollection, error)

	FileCreate(ctx context.Context, projectName, collectionId, localPath string, isDir bool, src *io.Reader) (*generated.SyncFile, error)
	File(ctx context.Context, projectName, collectionId, fileId string) (*generated.SyncFile, *io.ReadCloser, error)
	Files(ctx context.Context, projectName, collectionId string, withLocation bool) ([]generated.SyncFile, error)
	FileUpdate(ctx context.Context, projectName, collectionId, fileId, localPath string, isDir bool, src *io.Reader) (*generated.SyncFile, error)
	FileDelete(ctx context.Context, projectName, collectionId, fileId string) error
}

type c struct {
	endpoint            string
	client              *generated.Client
	clientWithResponses *generated.ClientWithResponses
}

func New(endpoint, token string) I {
	c := &c{
		endpoint:            endpoint,
		client:              createClientMust(endpoint, token),
		clientWithResponses: createClientWithResponsesMust(endpoint, token),
	}

	return c
}

func createClientWithResponsesMust(endpoint, token string) *generated.ClientWithResponses {
	client, err := createClientWithResponses(endpoint, token)
	if err != nil {
		panic(err)
	}
	return client
}

func createClientWithResponses(endpoint, token string) (*generated.ClientWithResponses, error) {
	return generated.NewClientWithResponses(endpoint, generated.WithRequestEditorFn(
		func(ctx context.Context, req *http.Request) (err error) {
			if token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}

			return nil
		},
	))
}

func createClientMust(endpoint, token string) *generated.Client {
	client, err := createClient(endpoint, token)
	if err != nil {
		panic(err)
	}
	return client
}

func createClient(endpoint, token string) (*generated.Client, error) {
	return generated.NewClient(endpoint, generated.WithRequestEditorFn(
		func(ctx context.Context, req *http.Request) (err error) {
			if token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}

			return nil
		},
	))
}
