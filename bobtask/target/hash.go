package target

import (
	"errors"
	"fmt"

	"github.com/benchkram/bob/pkg/dockermobyutil"
	"github.com/benchkram/bob/pkg/usererror"
)

func (t *T) dockerImageHash(image string) (string, error) {
	hash, err := t.dockerRegistryClient.ImageHash(image)
	if err != nil {
		if errors.Is(err, dockermobyutil.ErrImageNotFound) {
			return "", usererror.Wrapm(err, "failed to fetch docker image hash")
		} else {
			return "", fmt.Errorf("failed to get docker image hash info %q: %w", image, err)
		}
	}
	return hash, nil
}
