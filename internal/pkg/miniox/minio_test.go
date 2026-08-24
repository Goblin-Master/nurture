package miniox

import (
	"context"
	"nurture/internal/config"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
)

func TestInitMinioDisabled(t *testing.T) {
	old := config.Conf.Minio
	t.Cleanup(func() {
		config.Conf.Minio = old
	})
	config.Conf.Minio = config.Minio{Enable: false}

	if client := InitMinio(); client != nil {
		t.Fatal("InitMinio() returned client when minio is disabled")
	}
}

func TestMinio(t *testing.T) {
	// 加载配置
	config.LoadConfig()
	if !config.Conf.Minio.Enable {
		t.Skip("skip minio integration test: minio.enable=false")
	}

	// 初始化 minio
	client := InitMinio()
	buckets, err := client.ListBuckets(context.Background())
	assert.NoError(t, err)
	for _, bucket := range buckets {
		t.Logf("bucket: %v", bucket)
	}
	// 检查存储桶是否存在
	bucketExists, err := client.BucketExists(context.Background(), config.Conf.Minio.Bucket)
	assert.NoError(t, err)
	assert.True(t, bucketExists)
	//查看文件列表
	objects := client.ListObjects(context.Background(), config.Conf.Minio.Bucket, minio.ListObjectsOptions{
		Recursive: true,
	})
	for object := range objects {
		t.Logf("object: %v", object)
	}
}
