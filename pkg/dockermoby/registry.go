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

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var (
	ErrImageNotFoundByTag = fmt.Errorf("Image not found by tag")
)

type DockerRegistry interface {
	FetchImageHash(repotag string) (string, error)
	SaveImage(imageID string, savedir string, imgtag string) (string, error)
	DeleteImage(imageID string) error
	LoadImage(dir string, imgtag string) error
	TagImage(source, target string) error
}

type R struct {
	client *client.Client
}

func New() (DockerRegistry, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}

	r := &R{
		client: cli,
	}

	return r, nil
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

func (r *R) SaveImage(imageID string, savedir string, imgtag string) (string, error) {
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

func (r *R) LoadImage(dir string, imgtag string) error {
	filename := filepath.Join(dir, imgtag+".tar")

	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := io.Reader(f)
	resp, err := r.client.ImageLoad(context.Background(), reader, true)
	if err != nil {
		return err
	}

	imageID, err := getImageIDFromResponse(resp.Body)
	defer resp.Body.Close()

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
