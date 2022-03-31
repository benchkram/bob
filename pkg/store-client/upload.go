package storeclient

import (
	"context"
	"fmt"
	"github.com/benchkram/errz"
	"github.com/pkg/errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
)

func (c *c) Upload(
	ctx context.Context,
	projectID string,
	artifactID string,
	src io.Reader,
) (err error) {
	defer errz.Recover(&err)

	r, w := io.Pipe()
	mpw := multipart.NewWriter(w)

	go func() {
		err := attachMimeHeader(mpw, "id", artifactID)
		if err != nil {
			_ = w.CloseWithError(err)
		}

		pw, err0 := mpw.CreateFormFile("file", artifactID+".bin")
		if err0 != nil {
			_ = w.CloseWithError(err)
		}

		tr := io.TeeReader(src, pw)
		buf := make([]byte, 256)
		for {
			_, err := tr.Read(buf)
			if err == io.EOF {
				_ = mpw.Close()
				_ = w.Close()
				break
			}
			if err != nil {
				_ = w.CloseWithError(err)
			}
		}
	}()

	resp, err := c.clientWithResponses.CreateProjectArtifactWithBodyWithResponse(
		ctx, projectID, mpw.FormDataContentType(), r)
	errz.Fatal(err)

	if resp.StatusCode() != http.StatusOK {
		err = errors.Errorf("request failed [status: %d, msg: %q]", resp.StatusCode(), resp.Body)
		errz.Fatal(err)
	}

	return nil
}

func attachMimeHeader(w *multipart.Writer, key, value string) error {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"`, key))
	h.Set("Content-Type", "text/plain")

	p, err := w.CreatePart(h)
	if err != nil {
		return errors.Wrapf(err, "failed to create form field [%s]", key)
	}

	if _, err := p.Write([]byte(value)); err != nil {
		return errors.Wrapf(err, "failed to write form field [%s]", key)
	}

	return nil
}
