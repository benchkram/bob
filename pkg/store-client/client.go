package storeclient

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"

	"github.com/benchkram/errz"
	"github.com/pkg/errors"
	"github.com/schollz/progressbar/v3"

	"github.com/benchkram/bob/pkg/usererror"
)

var ErrProjectNotFound = errors.New("project not found")

func (c *c) UploadArtifact(
	ctx context.Context,
	projectName string,
	artifactID string,
	src io.Reader,
	size int64,
) (err error) {
	defer errz.Recover(&err)

	r, w := io.Pipe()
	mpw := multipart.NewWriter(w)

	bar := progressbar.DefaultBytes(
		size,
		fmt.Sprintf("Upload %s", artifactID),
	)

	go func() {
		err0 := attachMimeHeader(mpw, "id", artifactID)
		if err0 != nil {
			_ = w.CloseWithError(err)
			return
		}

		pw, err0 := mpw.CreateFormFile("file", artifactID+".bin")
		if err0 != nil {
			_ = w.CloseWithError(err)
			return
		}

		tr := io.TeeReader(src, pw)
		buf := make([]byte, 8192)

		for {
			n, err0 := tr.Read(buf)
			if err0 == io.EOF {
				_ = mpw.Close()
				_ = w.Close()
				break
			}
			if err0 != nil {
				_ = w.CloseWithError(err)
			}
			_ = bar.Add(n)
		}
	}()

	resp, err := c.clientWithResponses.UploadArtifactWithBodyWithResponse(
		ctx, projectName, mpw.FormDataContentType(), r)
	errz.Fatal(err)

	err = bar.Finish()
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

func (c *c) ListArtifacts(ctx context.Context, project string) (ids []string, err error) {
	defer errz.Recover(&err)

	res, err := c.clientWithResponses.GetProjectArtifactsWithResponse(
		ctx, project)
	errz.Fatal(err)

	if res.StatusCode() == http.StatusNotFound {
		errz.Fatal(usererror.Wrapm(ErrProjectNotFound, "upload to remote repository failed"))
	} else if res.StatusCode() != http.StatusOK {
		err = errors.Errorf("request failed [status: %d, msg: %q]", res.StatusCode(), res.Body)
		errz.Fatal(err)
	}

	if res.JSON200 == nil {
		errz.Fatal(errors.New("invalid response"))
	}

	return *res.JSON200, nil
}

func (c *c) GetArtifact(ctx context.Context, projectId string, artifactId string) (rc io.ReadCloser, size int64, err error) {
	defer errz.Recover(&err)

	res, err := c.clientWithResponses.GetProjectArtifactWithResponse(
		ctx, projectId, artifactId)
	errz.Fatal(err)

	if res.StatusCode() == http.StatusNotFound {
		errz.Fatal(usererror.Wrapm(ErrProjectNotFound, "upload to remote repository failed"))
	} else if res.StatusCode() != http.StatusOK {
		err = errors.Errorf("request failed [status: %d, msg: %q]", res.StatusCode(), res.Body)
		errz.Fatal(err)
	}

	if res.JSON200 == nil {
		errz.Fatal(errors.New("invalid response"))
	}

	res2, err := http.Get(*res.JSON200.Location)
	errz.Fatal(err)

	if res2.StatusCode != http.StatusOK {
		errz.Fatal(fmt.Errorf("invalid response"))
	}

	bar := progressbar.DefaultBytes(
		res2.ContentLength,
		fmt.Sprintf("Download %s", artifactId),
	)

	rb := progressbar.NewReader(res2.Body, bar)

	return &rb, res2.ContentLength, nil
}
