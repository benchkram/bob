package dockermobyutil

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/benchkram/errz"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var (
	ErrImageNotFound    = fmt.Errorf("image not found")
	ErrConnectionFailed = errors.New("connection to docker daemon failed")
)

type RegistryClient interface {
	ImageExists(image string) (bool, error)
	ImageHash(image string) (string, error)
	ImageSave(image string) (pathToArchive string, _ error)
	ImageRemove(image string) error
	ImageTag(src string, target string) error

	ImageLoad(pathToArchive string) error
}

type R struct {
	client     *client.Client
	archiveDir string

	// mutex assure to only sequentially access a local docker registry.
	// Some storage driver might not allow for parallel image extraction,
	// @rdnt realised this on his ubuntu22.04 using a zfsfilesystem.
	// Some context https://github.com/moby/moby/issues/21814
	//
	// explicitly using a pointer here to beeing able to detect
	// weather a mutex is required.
	mutex *sync.Mutex
}

func NewRegistryClient() (RegistryClient, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		errz.Fatal(err)
	}

	r := &R{
		client:     cli,
		archiveDir: os.TempDir(),
	}

	// Use a lock to suppress parallel image reads on zfs.
	info, err := r.client.Info(context.Background())
	if client.IsErrConnectionFailed(err) {
		return nil, ErrConnectionFailed
	} else if err != nil {
		return nil, err
	}

	if info.Driver == "zfs" {
		r.mutex = &sync.Mutex{}
	}

	return r, nil
}

func (r *R) ImageExists(image string) (bool, error) {
	_, err := r.ImageHash(image)
	if err != nil {
		if errors.Is(err, ErrImageNotFound) {
			return false, nil
		}
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
	if r.mutex != nil {
		r.mutex.Lock()
		defer r.mutex.Unlock()
	}
	reader, err := r.client.ImageSave(context.Background(), []string{image})
	if err != nil {
		return "", err
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	// rndExtension is added to the archive name. It prevents overwrite of images in tmp directory in case
	// of a image beeing used as target in multiple tasks (which should be avoided).
	rndExtension := randStringRunes(8)

	image = strings.ReplaceAll(image, "/", "-")

	pathToArchive = filepath.Join(savedir, image+"-"+rndExtension+".tar")
	err = os.WriteFile(pathToArchive, body, 0644)
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

// ImageLoad from tar archive
func (r *R) ImageTag(src string, target string) error {
	return r.client.ImageTag(context.Background(), src, target)
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
