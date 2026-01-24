package logic

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"nurture/internal/config"
	"nurture/internal/global"
	"path/filepath"

	"github.com/minio/minio-go/v7"
)

type ICommonLogic interface {
	UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error)
}

type CommonLogic struct{}

func NewCommonLogic() *CommonLogic {
	return &CommonLogic{}
}

var _ ICommonLogic = (*CommonLogic)(nil)

func (l *CommonLogic) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error) {
	// 1. Calculate MD5
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		global.Log.Error(err)
		return "", ErrFileRead
	}
	fileHash := hex.EncodeToString(hash.Sum(nil))

	// Reset file pointer
	if _, err := file.Seek(0, 0); err != nil {
		global.Log.Error(err)
		return "", ErrFileRead
	}

	// 2. Generate filename
	ext := filepath.Ext(header.Filename)
	objectName := fileHash + ext

	// 3. Construct URL
	protocol := "http"
	if config.Conf.Minio.UseSSL {
		protocol = "https"
	}
	url := fmt.Sprintf("%s://%s/%s/%s", protocol, config.Conf.Minio.Endpoint, config.Conf.Minio.Bucket, objectName)

	// 4. Check if file exists
	_, err := global.MIO.StatObject(ctx, config.Conf.Minio.Bucket, objectName, minio.StatObjectOptions{})
	if err == nil {
		return url, nil
	}

	// 5. Upload to MinIO
	_, err = global.MIO.PutObject(ctx, config.Conf.Minio.Bucket, objectName, file, header.Size, minio.PutObjectOptions{
		ContentType: header.Header.Get("Content-Type"),
	})
	if err != nil {
		global.Log.Error(err)
		return "", ErrFileUpload
	}

	return url, nil
}
