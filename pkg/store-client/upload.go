package storeclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/Benchkram/errz"
)

func (c *c) Upload(
	ctx context.Context,
	projectID string,
	artifactID string,
	contentType string,
	filename string,
	src io.Reader,
) (err error) {

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	err = w.WriteField("id", artifactID)
	errz.Fatal(err)

	fieldWriter, err := w.CreateFormFile("file", filename)
	errz.Fatal(err)

	go func() {
		_, err = io.Copy(fieldWriter, src)
		errz.Log(err)
		w.Close()
	}()

	response, err := c.clientWithResponses.CreateProjectArtifactWithBodyWithResponse(ctx,
		projectID,
		contentType,
		&b,
		func(ctx context.Context, req *http.Request) (err error) {
			req.ContentLength = -1
			return nil
		},
	)
	errz.Fatal(err)

	if response.StatusCode() != 200 {
		return fmt.Errorf("failed upload request [code: %d] [msg: %s]", response.StatusCode(), string(response.Body))
	}

	return nil
}
