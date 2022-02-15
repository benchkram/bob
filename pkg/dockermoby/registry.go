package dockermoby

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Benchkram/errz"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var (
	ErrImageNotFoundByTag = fmt.Errorf("Image not found by tag")
)

const DefaultArchiveDir = "/tmp/docker-archives"

type RegistryHandler interface {
	GetArchiveDir() string
	ImageExists(repotag string) (bool, error)
	FetchImageHash(repotag string) (string, error)
	SaveImageToPath(imageID string, savedir string, imgtag string) (string, error)
	SaveImage(imageID string, imgtag string) (string, error)
	DeleteImage(imageID string) error
	LoadImage(imgpath string) error
	TagImage(source, target string) error
}

type R struct {
	client     *client.Client
	archiveDir string
}

func New() RegistryHandler {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		errz.Fatal(err)
	}

	r := &R{
		client:     cli,
		archiveDir: DefaultArchiveDir,
	}

	err = os.MkdirAll(r.archiveDir, os.ModePerm)
	if err != nil {
		errz.Fatal(err)
	}

	return r
}

func (r *R) GetArchiveDir() string {
	return r.archiveDir
}

func (r *R) ImageExists(repotag string) (bool, error) {
	_, err := r.FetchImageHash(repotag)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *R) FetchImageHash(repotag string) (string, error) {
	summaries, err := r.client.ImageList(context.Background(), types.ImageListOptions{All: false})
	if err != nil {
		return "", err
	}

	var selected types.ImageSummary
	for _, s := range summaries {
		for _, rtag := range s.RepoTags {
			if rtag == repotag {
				selected = s
				break
			}
		}
	}

	if selected.ID == "" {
		return "", ErrImageNotFoundByTag
	}

	return selected.ID, nil
}

func (r *R) SaveImageToPath(imageID string, savedir string, imgtag string) (string, error) {
	reader, err := r.client.ImageSave(context.Background(), []string{imageID})
	if err != nil {
		return "", err
	}
	defer reader.Close()

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}

	filename := imgtag
	if imgtag == "" {
		filename = getImageIdFromHash(imageID)
	}

	imagePath := filepath.Join(savedir, filename+".tar")
	err = ioutil.WriteFile(imagePath, body, 0644)
	if err != nil {
		return "", err
	}

	return imagePath, nil
}

func (r *R) SaveImage(imageID string, imgtag string) (string, error) {
	return r.SaveImageToPath(imageID, r.archiveDir, imgtag)
}

func (r *R) DeleteImage(imageID string) error {
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

func (r *R) LoadImage(imgpath string) error {
	filename := filepath.Base(imgpath)
	nameparts := strings.Split(filename, ".")
	imgtag := nameparts[0]

	f, err := os.Open(imgpath)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := io.Reader(f)
	resp, err := r.client.ImageLoad(context.Background(), reader, true)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	imageID, err := getImageIDFromResponse(resp.Body)
	if err != nil {
		return err
	}

	// set the image tag for the newly loaded images
	err = r.TagImage(imageID, imgtag)
	if err != nil {
		return err
	}

	return nil
}

func (r *R) TagImage(source, target string) error {
	err := r.client.ImageTag(context.Background(), source, target)
	if err != nil {
		return err
	}

	return nil
}

func getImageIdFromHash(hashID string) string {
	output := hashID
	if strings.HasPrefix(hashID, "sha256") {
		output = hashID[7:]
	}

	return output[:12]
}

// getImageIDFromResponse parse imageID from strings like below,
//
// {"stream":"Loaded image ID: sha256:7a2e4090ba3ffb8278f358c82c2015179ab37da9fb6358a3cbfa71e1cf9901a5\n"}
func getImageIDFromResponse(body io.Reader) (string, error) {
	rstart := regexp.MustCompile(`sha256:(.+)\\n`)

	stream, err := ioutil.ReadAll(body)
	if err != nil {
		return "", err
	}

	matched := rstart.FindString(string(stream))

	// do the necessary trimming, does includes the ending newline and
	//  brackets on the regex matched string
	trimmed := strings.Replace(matched, "\\n", "", -1)
	trimmed = strings.Replace(trimmed, "\"", "", -1)
	trimmed = strings.Replace(trimmed, "}", "", -1)

	trimmed = strings.TrimSpace(trimmed)

	return trimmed, nil
}
