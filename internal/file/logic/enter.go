package logic

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"nurture/internal/config"
	"nurture/internal/file/constant"
	"nurture/internal/pkg/zapx"
	"path/filepath"

	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
)

type IFileLogic interface {
	Upload(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error)
}

type ObjectStorage interface {
	StatObject(ctx context.Context, bucketName, objectName string, opts minio.StatObjectOptions) (minio.ObjectInfo, error)
	PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error)
}

type FileLogic struct {
	storage ObjectStorage
	config  config.Minio
	log     *zap.SugaredLogger
}

func NewFileLogic(storage ObjectStorage, cfg config.Minio, log *zap.SugaredLogger) *FileLogic {
	return &FileLogic{
		storage: storage,
		config:  cfg,
		log:     zapx.OrNop(log),
	}
}

var _ IFileLogic = (*FileLogic)(nil)

func (l *FileLogic) Upload(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error) {
	if !l.config.Enable || l.storage == nil {
		return "", ErrFileUpload
	}
	if header == nil {
		return "", ErrFileRead
	}
	if header.Size > constant.FileMaxSize {
		return "", ErrFileOverSize
	}

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		l.log.Error(err)
		return "", ErrFileRead
	}
	fileHash := hex.EncodeToString(hash.Sum(nil))

	if _, err := file.Seek(0, 0); err != nil {
		l.log.Error(err)
		return "", ErrFileRead
	}

	objectName := fileHash + filepath.Ext(header.Filename)
	url := l.objectURL(objectName)

	if _, err := l.storage.StatObject(ctx, l.config.Bucket, objectName, minio.StatObjectOptions{}); err == nil {
		return url, nil
	}

	_, err := l.storage.PutObject(ctx, l.config.Bucket, objectName, file, header.Size, minio.PutObjectOptions{
		ContentType: header.Header.Get("Content-Type"),
	})
	if err != nil {
		l.log.Error(err)
		return "", ErrFileUpload
	}
	return url, nil
}

func (l *FileLogic) objectURL(objectName string) string {
	protocol := "http"
	if l.config.UseSSL {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", protocol, l.config.Endpoint, l.config.Bucket, objectName)
}
