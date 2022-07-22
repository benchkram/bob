package remotesyncstore

import (
	"context"
	"fmt"
	storeclient "github.com/benchkram/bob/pkg/store-client"
	"github.com/benchkram/bob/pkg/versionedsync/collection"
	"github.com/benchkram/errz"
	"io"
)

// S is a versioned sync store that handles the logic of accessing the bob-server to store collections and files
// it uses the store client interface to achieve that
type S struct {
	username string
	client   storeclient.I
	project  string
}

func New(username, project string, opts ...Option) *S {
	s := &S{
		username: username,
		project:  project,
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(s)
	}

	if s.client == nil {
		panic(fmt.Errorf("no client"))
	}

	return s
}

func (s *S) CollectionCreate(ctx context.Context, name, tag, path string) (cId string, err error) {
	defer errz.Recover(&err)
	genC, err := s.client.CollectionCreate(ctx, s.project, collection.JoinNameAndVersion(name, tag), path)
	errz.Fatal(err)
	c, err := collection.FromRestType(genC)
	errz.Fatal(err)
	return c.ID, nil
}
func (s *S) Collection(ctx context.Context, collectionId string) (c *collection.C, err error) {
	defer errz.Recover(&err)
	genC, err := s.client.Collection(ctx, s.project, collectionId)
	errz.Fatal(err)
	c, err = collection.FromRestType(genC)
	errz.Fatal(err)
	return c, nil
}
func (s *S) CollectionIdByName(ctx context.Context, name, tag string) (cId string, err error) {
	remoteName := collection.JoinNameAndVersion(name, tag)
	defer errz.Recover(&err)
	collections, err := s.client.Collections(ctx, s.project)
	errz.Fatal(err)
	for _, c := range collections {
		if c.Name == remoteName {
			return c.Id, nil
		}
	}
	return "", ErrCollectionNotFound
}

func (s *S) Collections(ctx context.Context) (collections []*collection.C, err error) {
	defer errz.Recover(&err)
	genCs, err := s.client.Collections(ctx, s.project)
	errz.Fatal(err)
	for _, gc := range genCs {
		c, err := collection.FromRestType(&gc)
		errz.Fatal(err)
		collections = append(collections, c)
	}

	return collections, nil
}
func (s *S) FileUpload(ctx context.Context, collectionId, localPath string, srcReader io.Reader) (err error) {
	defer errz.Recover(&err)
	_, err = s.client.FileCreate(ctx, s.project, collectionId, localPath, srcReader)
	errz.Fatal(err)
	return nil

}
func (s *S) File(ctx context.Context, collectionId, fileId string) (r io.ReadCloser, err error) {
	defer errz.Recover(&err)

	_, r, err = s.client.File(ctx, s.project, collectionId, fileId)
	errz.Fatal(err)
	return r, nil
}
func (s *S) FileUpdate(ctx context.Context, collectionId, fileId string, srcReader io.Reader) (err error) {
	defer errz.Recover(&err)
	_, err = s.client.FileUpdate(ctx, s.project, collectionId, fileId, "", &srcReader)
	errz.Fatal(err)
	return nil
}
func (s *S) FileDelete(ctx context.Context, collectionId, fileId string) (err error) {
	defer errz.Recover(&err)
	err = s.client.FileDelete(ctx, s.project, collectionId, fileId)
	errz.Fatal(err)
	return nil
}
