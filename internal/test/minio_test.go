package test

import (
	"context"
	"nurture/internal/config"
	"nurture/internal/global"
	"os"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
)

func TestMinio(t *testing.T) {
	if os.Getenv("NURTURE_RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("skip integration test: set NURTURE_RUN_INTEGRATION_TESTS=1 to run")
	}
	// 加载配置
	config.LoadConfig()
	// 初始化全局配置
	global.Init()

	// 初始化 minio
	buckets, err := global.MIO.ListBuckets(context.Background())
	assert.NoError(t, err)
	for _, bucket := range buckets {
		t.Logf("bucket: %v", bucket)
	}
	// 检查存储桶是否存在
	bucketExists, err := global.MIO.BucketExists(context.Background(), config.Conf.Minio.Bucket)
	assert.NoError(t, err)
	assert.True(t, bucketExists)
	//查看文件列表
	objects := global.MIO.ListObjects(context.Background(), config.Conf.Minio.Bucket, minio.ListObjectsOptions{
		Recursive: true,
	})
	for object := range objects {
		t.Logf("object: %v", object)
	}
}
