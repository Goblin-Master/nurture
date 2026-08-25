package test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/textproto"
	"nurture/internal/config"
	"nurture/internal/file/constant"
	"nurture/internal/file/logic"
	"strings"
	"testing"

	"github.com/minio/minio-go/v7"
)

type storageFake struct {
	statErr error
	putErr  error
	putData string
}

func (f *storageFake) StatObject(context.Context, string, string, minio.StatObjectOptions) (minio.ObjectInfo, error) {
	return minio.ObjectInfo{}, f.statErr
}

func (f *storageFake) PutObject(_ context.Context, _ string, _ string, reader io.Reader, _ int64, _ minio.PutObjectOptions) (minio.UploadInfo, error) {
	data, _ := io.ReadAll(reader)
	f.putData = string(data)
	return minio.UploadInfo{}, f.putErr
}

func TestFileLogicUploadReturnsExistingObjectURL(t *testing.T) {
	l := logic.NewFileLogic(&storageFake{}, minioConfig(), nil)
	url, err := l.Upload(t.Context(), multipartFile("hello"), fileHeader("avatar.png", int64(len("hello"))))

	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	want := "http://minio.local/nurture/5d41402abc4b2a76b9719d911017c592.png"
	if url != want {
		t.Fatalf("Upload() url = %q, want %q", url, want)
	}
}

func TestFileLogicUploadStoresMissingObject(t *testing.T) {
	storage := &storageFake{statErr: errors.New("not found")}
	l := logic.NewFileLogic(storage, minioConfig(), nil)

	_, err := l.Upload(t.Context(), multipartFile("hello"), fileHeader("avatar.png", int64(len("hello"))))

	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	if storage.putData != "hello" {
		t.Fatalf("put data = %q, want hello", storage.putData)
	}
}

func TestFileLogicUploadRejectsDisabledStorage(t *testing.T) {
	cfg := minioConfig()
	cfg.Enable = false
	l := logic.NewFileLogic(&storageFake{}, cfg, nil)

	_, err := l.Upload(t.Context(), multipartFile("hello"), fileHeader("avatar.png", int64(len("hello"))))

	if !errors.Is(err, logic.ErrFileUpload) {
		t.Fatalf("Upload() error = %v, want %v", err, logic.ErrFileUpload)
	}
}

func TestFileLogicUploadRejectsOversizeFile(t *testing.T) {
	l := logic.NewFileLogic(&storageFake{}, minioConfig(), nil)

	_, err := l.Upload(t.Context(), multipartFile("hello"), fileHeader("avatar.png", constant.FileMaxSize+1))

	if !errors.Is(err, logic.ErrFileOverSize) {
		t.Fatalf("Upload() error = %v, want %v", err, logic.ErrFileOverSize)
	}
}

func minioConfig() config.Minio {
	return config.Minio{
		Enable:   true,
		Endpoint: "minio.local",
		Bucket:   "nurture",
	}
}

func multipartFile(content string) multipart.File {
	return multipartFileReader{Reader: bytes.NewReader([]byte(content))}
}

func fileHeader(name string, size int64) *multipart.FileHeader {
	return &multipart.FileHeader{
		Filename: name,
		Size:     size,
		Header: textproto.MIMEHeader{
			"Content-Type": []string{"image/png"},
		},
	}
}

type multipartFileReader struct {
	*bytes.Reader
}

func (r multipartFileReader) Close() error {
	return nil
}

func TestMultipartFileReaderCanSeek(t *testing.T) {
	file := multipartFile("hello")
	if _, err := file.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	buf := new(strings.Builder)
	if _, err := io.Copy(buf, file); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "hello" {
		t.Fatalf("read = %q, want hello", buf.String())
	}
}
