package storeclient

import (
	"context"
	"fmt"
	"github.com/benchkram/bob/pkg/store-client/generated"
	"github.com/benchkram/errz"
	"github.com/pkg/errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/benchkram/bob/pkg/usererror"
)

var (
	ErrProjectNotFound     = errors.New("project not found")
	ErrCollectionNotFound  = errors.New("collection not found")
	ErrFileNotFound        = errors.New("file not found")
	ErrAuthorizationFailed = errors.New("authorization failed")
	ErrResourceForbidden   = errors.New("accessed to resource forbidden")
	ErrEmptyResponse       = errors.New("empty response")
	ErrDownloadFailed      = errors.New("binary download failed")
	ErrConnectionRefused   = errors.New("connection to server failed (connection refused)")
)

func (c *c) UploadArtifact(
	ctx context.Context,
	projectName string,
	artifactID string,
	src io.Reader,
) (err error) {
	defer errz.Recover(&err)

	r, w := io.Pipe()
	mpw := multipart.NewWriter(w)

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
			_, err0 := tr.Read(buf)
			if err0 == io.EOF {
				_ = mpw.Close()
				_ = w.Close()
				break
			}
			if err0 != nil {
				_ = w.CloseWithError(err)
			}
		}
	}()

	resp, err := c.clientWithResponses.UploadArtifactWithBodyWithResponse(
		ctx, projectName, mpw.FormDataContentType(), r)
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
		errz.Fatal(ErrEmptyResponse)
	}

	return *res.JSON200, nil
}

func (c *c) GetArtifact(ctx context.Context, projectId string, artifactId string) (rc io.ReadCloser, err error) {
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
		errz.Fatal(ErrEmptyResponse)
	}

	res2, err := http.Get(*res.JSON200.Location)
	errz.Fatal(err)

	if res2.StatusCode != http.StatusOK {
		errz.Fatal(fmt.Errorf("invalid response"))
	}

	return res2.Body, nil
}

func (c *c) CollectionCreate(ctx context.Context, projectName, name, localPath string) (collection *generated.SyncCollection, err error) {
	defer errz.Recover(&err)
	body := generated.CreateSyncCollectionJSONRequestBody{
		LocalPath: localPath,
		Name:      name,
	}

	res, err := c.clientWithResponses.CreateSyncCollectionWithResponse(
		ctx,
		projectName,
		body,
	)
	errMsg := "creation of sync collection on remote failed"
	if errors.Is(err, syscall.ECONNREFUSED) {
		errz.Fatal(usererror.Wrapm(ErrConnectionRefused, errMsg))
	}
	errz.Fatal(err)

	switch res.StatusCode() {
	case http.StatusOK:
	case http.StatusNotFound:
		errz.Fatal(usererror.Wrapm(ErrProjectNotFound, errMsg))
	case http.StatusUnauthorized:
		err = usererror.Wrapm(ErrAuthorizationFailed, errMsg)
		errz.Fatal(err)
	case http.StatusForbidden:
		err = usererror.Wrapm(ErrResourceForbidden, errMsg)
		errz.Fatal(err)
	default:
		err = errors.Errorf("request failed [status: %d, msg: %q]", res.StatusCode(), res.Body)
		errz.Fatal(err)
	}

	if res.JSON200 == nil {
		errz.Fatal(ErrEmptyResponse)
	}

	return res.JSON200, nil

}

func (c *c) Collection(ctx context.Context, projectName, collectionId string) (collection *generated.SyncCollection, err error) {
	defer errz.Recover(&err)

	res, err := c.clientWithResponses.GetSyncCollectionWithResponse(
		ctx,
		projectName,
		collectionId,
	)
	errMsg := "reading of sync collection from remote failed"
	if errors.Is(err, syscall.ECONNREFUSED) {
		errz.Fatal(usererror.Wrapm(ErrConnectionRefused, errMsg))
	}
	errz.Fatal(err)

	switch res.StatusCode() {
	case http.StatusOK:
	case http.StatusUnauthorized:
		errz.Fatal(usererror.Wrapm(ErrAuthorizationFailed, errMsg))
	case http.StatusForbidden:
		errz.Fatal(usererror.Wrapm(ErrResourceForbidden, errMsg))
	case http.StatusNotFound:
		errz.Fatal(usererror.Wrapm(ErrCollectionNotFound, errMsg))
	default:
		err = errors.Errorf("request failed [status: %d, msg: %q]", res.StatusCode(), res.Body)
		errz.Fatal(err)
	}

	if res.JSON200 == nil {
		errz.Fatal(ErrEmptyResponse)
	}

	return res.JSON200, nil
}

func (c *c) Collections(ctx context.Context, projectName string) (collections []generated.SyncCollection, err error) {
	defer errz.Recover(&err)

	res, err := c.clientWithResponses.GetSyncCollectionsWithResponse(
		ctx,
		projectName,
	)
	errMsg := "reading of sync collections from remote failed"
	if errors.Is(err, syscall.ECONNREFUSED) {
		errz.Fatal(usererror.Wrapm(ErrConnectionRefused, errMsg))
	}

	errz.Fatal(err)

	switch res.StatusCode() {
	case http.StatusOK:
	case http.StatusUnauthorized:
		err = usererror.Wrapm(ErrAuthorizationFailed, errMsg)
		errz.Fatal(err)
	case http.StatusForbidden:
		errz.Fatal(usererror.Wrapm(ErrResourceForbidden, errMsg))
	case http.StatusNotFound:
		err = usererror.Wrapm(ErrProjectNotFound, errMsg)
		errz.Fatal(err)
	default:
		err = errors.Errorf("request failed [status: %d, msg: %q]", res.StatusCode(), res.Body)
		errz.Fatal(err)
	}
	if res.JSON200 == nil {
		errz.Fatal(ErrEmptyResponse)
	}

	return *res.JSON200, nil
}

func (c *c) FileCreate(ctx context.Context, projectName, collectionId, localPath string, isDir bool, src *io.Reader) (f *generated.SyncFile, err error) {
	r, w := io.Pipe()
	mpw := multipart.NewWriter(w)

	go func() {
		err0 := mpw.WriteField("local_path", localPath)
		if err0 != nil {
			_ = w.CloseWithError(err0)
		}
		err0 = mpw.WriteField("is_directory", strconv.FormatBool(isDir))
		if err0 != nil {
			_ = w.CloseWithError(err0)
		}

		if src != nil {
			fieldWriter, err0 := mpw.CreateFormFile("file", filepath.Base(localPath))
			if err0 != nil {
				_ = w.CloseWithError(err0)
			}
			_, err0 = io.Copy(fieldWriter, *src)
			if err0 != nil {
				_ = w.CloseWithError(err0)
			}
		}
		err0 = mpw.Close()
		if err0 != nil {
			_ = w.CloseWithError(err0)
		}
		_ = w.Close()
	}()

	resp, err := c.clientWithResponses.CreateSyncFileWithBodyWithResponse(
		ctx,
		projectName,
		collectionId,
		mpw.FormDataContentType(),
		r,
	)
	errMsg := "creation of file on remote failed"
	if errors.Is(err, syscall.ECONNREFUSED) {
		errz.Fatal(usererror.Wrapm(ErrConnectionRefused, errMsg))
	}
	errz.Fatal(err)

	switch resp.StatusCode() {
	case http.StatusOK:
	case http.StatusUnauthorized:
		err = usererror.Wrapm(ErrAuthorizationFailed, errMsg)
		errz.Fatal(err)
	case http.StatusForbidden:
		err = usererror.Wrapm(ErrResourceForbidden, errMsg)
		errz.Fatal(err)
	case http.StatusNotFound:
		err = usererror.Wrapm(ErrCollectionNotFound, errMsg)
		errz.Fatal(err)
	default:
		//TODO: add specific error handling for http.StatusConflict and http.StatusBadRequest
		err = errors.Errorf("request failed [status: %d, msg: %q]", resp.StatusCode(), resp.Body)
		errz.Fatal(err)
	}
	if resp.JSON200 == nil {
		errz.Fatal(ErrEmptyResponse)
	}

	return resp.JSON200, nil
}

func (c *c) File(ctx context.Context, projectName, collectionId, fileId string) (f *generated.SyncFile, rc *io.ReadCloser, err error) {
	defer errz.Recover(&err)

	res, err := c.clientWithResponses.GetSyncFileWithResponse(
		ctx,
		projectName,
		collectionId,
		fileId,
	)
	errMsg := "reading of sync file from remote failed"
	if errors.Is(err, syscall.ECONNREFUSED) {
		errz.Fatal(usererror.Wrapm(ErrConnectionRefused, errMsg))
	}
	errz.Fatal(err)

	switch res.StatusCode() {
	case http.StatusOK:
	case http.StatusUnauthorized:
		err = usererror.Wrapm(ErrAuthorizationFailed, errMsg)
		errz.Fatal(err)
	case http.StatusForbidden:
		err = usererror.Wrapm(ErrResourceForbidden, errMsg)
		errz.Fatal(err)
	case http.StatusNotFound:
		errz.Fatal(usererror.Wrapm(ErrFileNotFound, errMsg))
	default:
		err = errors.Errorf("request failed [status: %d, msg: %q]", res.StatusCode(), res.Body)
		errz.Fatal(err)
	}
	if res.JSON200 == nil {
		errz.Fatal(ErrEmptyResponse)
	}
	if !res.JSON200.IsDirectory {

		res2, err := http.Get(*res.JSON200.Location)
		errz.Fatal(err)

		if res2.StatusCode != http.StatusOK {
			errz.Fatal(usererror.Wrapm(ErrDownloadFailed, fmt.Sprintf("reading from storage failed (Status %d)", res2.StatusCode)))
		}
		if res2.Body == nil {
			errz.Fatal(ErrEmptyResponse)
		}
		return res.JSON200, &res2.Body, nil
	} else {
		return res.JSON200, nil, nil
	}

}

func (c *c) Files(ctx context.Context, projectName, collectionId string, withLocation bool) (files []generated.SyncFile, err error) {
	defer errz.Recover(&err)

	params := generated.GetSyncFilesParams{
		WithLocation: &withLocation,
	}
	res, err := c.clientWithResponses.GetSyncFilesWithResponse(
		ctx,
		projectName,
		collectionId,
		&params,
	)
	errMsg := "reading of sync files from remote failed"
	if errors.Is(err, syscall.ECONNREFUSED) {
		errz.Fatal(usererror.Wrapm(ErrConnectionRefused, errMsg))
	}
	errz.Fatal(err)

	switch res.StatusCode() {
	case http.StatusOK:
	case http.StatusUnauthorized:
		err = usererror.Wrapm(ErrAuthorizationFailed, errMsg)
		errz.Fatal(err)
	case http.StatusForbidden:
		err = usererror.Wrapm(ErrResourceForbidden, errMsg)
		errz.Fatal(err)
	case http.StatusNotFound:
		errz.Fatal(usererror.Wrapm(ErrCollectionNotFound, errMsg))
	default:
		err = errors.Errorf("request failed [status: %d, msg: %q]", res.StatusCode(), res.Body)
		errz.Fatal(err)
	}
	if res.JSON200 == nil {
		errz.Fatal(ErrEmptyResponse)
	}

	return *res.JSON200, nil
}

func (c *c) FileUpdate(ctx context.Context, projectName, collectionId, fileId, localPath string, isDir bool, src *io.Reader) (file *generated.SyncFile, err error) {
	r, w := io.Pipe()
	mpw := multipart.NewWriter(w)

	go func() {
		if localPath != "" {
			err0 := mpw.WriteField("local_path", localPath)
			if err0 != nil {
				_ = w.CloseWithError(err0)
			}
		}

		err0 := mpw.WriteField("is_directory", strconv.FormatBool(isDir))
		if err0 != nil {
			_ = w.CloseWithError(err0)
		}

		if src != nil {
			fieldWriter, err0 := mpw.CreateFormFile("file", filepath.Base(localPath))
			if err0 != nil {
				_ = w.CloseWithError(err0)
			}
			_, err0 = io.Copy(fieldWriter, *src)
			if err0 != nil {
				_ = w.CloseWithError(err0)
			}
		}
		err0 = mpw.Close()
		if err0 != nil {
			_ = w.CloseWithError(err0)
		}
		_ = w.Close()
	}()

	resp, err := c.clientWithResponses.PutSyncFileWithBodyWithResponse(
		ctx,
		projectName,
		collectionId,
		fileId,
		mpw.FormDataContentType(),
		r,
	)
	errMsg := "update of sync file on remote failed"
	if errors.Is(err, syscall.ECONNREFUSED) {
		errz.Fatal(usererror.Wrapm(ErrConnectionRefused, errMsg))
	}
	errz.Fatal(err)

	switch resp.StatusCode() {
	case http.StatusOK:
	case http.StatusUnauthorized:
		err = usererror.Wrapm(ErrAuthorizationFailed, errMsg)
		errz.Fatal(err)
	case http.StatusForbidden:
		err = usererror.Wrapm(ErrResourceForbidden, errMsg)
		errz.Fatal(err)
	case http.StatusNotFound:
		err = usererror.Wrapm(ErrFileNotFound, errMsg)
		errz.Fatal(err)
	default:
		//TODO: add specific error handling for http.StatusConflict and http.StatusBadRequest
		err = errors.Errorf("request failed [status: %d, msg: %q]", resp.StatusCode(), resp.Body)
		errz.Fatal(err)
	}

	if resp.JSON200 == nil {
		errz.Fatal(ErrEmptyResponse)
	}

	return resp.JSON200, nil
}

func (c *c) FileDelete(ctx context.Context, projectName, collectionId, fileId string) (err error) {
	defer errz.Recover(&err)

	res, err := c.client.DeleteSyncFile(
		ctx,
		projectName,
		collectionId,
		fileId,
	)
	errMsg := "delete from remote failed"
	if errors.Is(err, syscall.ECONNREFUSED) {
		errz.Fatal(usererror.Wrapm(ErrConnectionRefused, errMsg))
	}
	errz.Fatal(err)

	switch res.StatusCode {
	case http.StatusOK:
	case http.StatusUnauthorized:
		err = usererror.Wrapm(ErrAuthorizationFailed, errMsg)
		errz.Fatal(err)
	case http.StatusForbidden:
		err = usererror.Wrapm(ErrResourceForbidden, errMsg)
		errz.Fatal(err)
	case http.StatusNotFound:
		errz.Fatal(usererror.Wrapm(ErrFileNotFound, errMsg))
	default:
		err = errors.Errorf("request failed [status: %d, msg: %q]", res.StatusCode, res.Body)
		errz.Fatal(err)
	}

	return nil
}
