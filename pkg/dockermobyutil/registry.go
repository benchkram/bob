package dockermobyutil

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/Benchkram/errz"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var (
	ErrImageNotFound = fmt.Errorf("image not found")
)

type RegistryClient interface {
	ImageExists(image string) (bool, error)
	ImageHash(image string) (string, error)
	ImageSave(image string) (pathToArchive string, _ error)
	ImageRemove(image string) error

	ImageLoad(pathToArchive string) error
}

type R struct {
	client     *client.Client
	archiveDir string
}

func NewRegistryClient() RegistryClient {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		errz.Fatal(err)
	}

	r := &R{
		client:     cli,
		archiveDir: os.TempDir(),
	}

	return r
}

func (r *R) ImageExists(image string) (bool, error) {
	_, err := r.ImageHash(image)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *R) ImageHash(image string) (string, error) {
	summaries, err := r.client.ImageList(context.Background(), types.ImageListOptions{All: false})
	if err != nil {
		return "", err
	}

	var selected types.ImageSummary
	for _, s := range summaries {
		for _, rtag := range s.RepoTags {
			if rtag == image {
				selected = s
				break
			}
		}
	}

	if selected.ID == "" {
		return "", fmt.Errorf("%s, %w", image, ErrImageNotFound)
	}

	return selected.ID, nil
}

func (r *R) imageSaveToPath(image string, savedir string) (pathToArchive string, _ error) {
	reader, err := r.client.ImageSave(context.Background(), []string{image})
	if err != nil {
		return "", err
	}
	defer reader.Close()

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}

	// rndExtension is added to the archive name. It prevents overwrite of images in tmp directory in case
	// of a image beeing used as target in multiple tasks (which should be avoided).
	rndExtension := randStringRunes(8)

	pathToArchive = filepath.Join(savedir, image+"-"+rndExtension+".tar")
	err = ioutil.WriteFile(pathToArchive, body, 0644)
	if err != nil {
		return "", err
	}

	return pathToArchive, nil
}

// ImageSave wraps for `docker save` with the addition to add a random string
// to archive name.
func (r *R) ImageSave(image string) (pathToArchive string, _ error) {
	return r.imageSaveToPath(image, r.archiveDir)
}

// ImageRemove from registry
func (r *R) ImageRemove(imageID string) error {
	options := types.ImageRemoveOptions{
		Force:         true,
		PruneChildren: true,
	}
	_, err := r.client.ImageRemove(context.Background(), imageID, options)
	if err != nil {
		return err
	}

	return nil
}

// ImageLoad from tar archive
func (r *R) ImageLoad(imgpath string) error {
	f, err := os.Open(imgpath)
	if err != nil {
		return err
	}
	defer f.Close()

	resp, err := r.client.ImageLoad(context.Background(), f, false)
	if err != nil {
		return err
	}

	return resp.Body.Close()
}

// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

func randStringRunes(n int) string {
	rand.Seed(time.Now().UnixNano())

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
