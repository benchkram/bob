package file

import (
	"github.com/benchkram/bob/pkg/store-client/generated"
)

// F represents a synced file
type F struct {
	// ID is the optional identifier on the server
	ID *string

	// LocalPath is the path of the file on the client relative to the collection root
	LocalPath string

	// Hash is the hash generated on the server or client over the file content
	// it is used to detect changes in the content when comparing local and remote
	Hash string
}

func FileFromRestType(f generated.SyncFile) *F {
	return &F{
		ID:        &f.Id,
		LocalPath: f.LocalPath,
		Hash:      *f.EncryptedHash,
	}
}

func FileFromRestStubType(f generated.SyncFileStub) *F {
	return &F{
		ID:        &f.Id,
		LocalPath: f.LocalPath,
		Hash:      *f.EncryptedHash,
	}
}
