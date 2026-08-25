package logic

import (
	"errors"
	"fmt"
	"nurture/internal/file/constant"
)

var (
	ErrFileOverSize = fmt.Errorf("文件大小不能超过%dMB", constant.FileMaxSize/1024/1024)
	ErrFileRead     = errors.New("文件读取失败")
	ErrFileUpload   = errors.New("文件上传失败")
)
