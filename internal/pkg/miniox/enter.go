package miniox

import (
	"fmt"
	"nurture/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func InitMinio() *minio.Client {
	client, err := minio.New(config.Conf.Minio.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.Conf.Minio.AccessKey, config.Conf.Minio.SecretKey, ""),
		Secure: config.Conf.Minio.UseSSL,
	})
	if err != nil {
		panic(fmt.Sprintf("minio init error: %v", err))
	}
	return client
}
