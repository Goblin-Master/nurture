package logic

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"nurture/internal/config"
	"nurture/internal/constant"
	"nurture/internal/dto"
	"nurture/internal/global"
	"nurture/internal/pkg/aix"
	"nurture/internal/repo"
	"path/filepath"
	"time"

	"github.com/minio/minio-go/v7"
)

type ICommonLogic interface {
	UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error)
	ChatStream(ctx context.Context, userID string, req dto.ChatStreamReq, streamFunc func(event dto.SSEEvent)) error
	UploadKnowledge(ctx context.Context, userID string, req dto.KnowledgeUploadReq) error
	GetChatHistory(ctx context.Context, userID string, req dto.ChatHistoryReq) (dto.ChatHistoryResp, error)
}

type CommonLogic struct {
	aiRepo *repo.AIRepo
}

func NewCommonLogic() *CommonLogic {
	return &CommonLogic{
		aiRepo: repo.NewAIRepo(),
	}
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

// ChatStream 流式对话
func (l *CommonLogic) ChatStream(ctx context.Context, userID string, req dto.ChatStreamReq,
	streamFunc func(event dto.SSEEvent)) error {

	// 1. 获取最近 3 轮对话历史（6 条消息）作为 AI 上下文
	history, err := l.aiRepo.GetRecentHistory(ctx, userID, req.SessionID, constant.AI_CONTEXT_MESSAGES)
	if err != nil {
		global.Log.Error(err)
		// 历史获取失败不阻断，继续对话
		history = []aix.ChatMessage{}
	}

	// 2. RAG 检索（暂略）
	var ragContext string

	// 3. 构建消息
	messages := global.AIX.BuildMessages(history, req.Message, req.Images, ragContext)

	// 4. 流式对话
	fullResponse, err := global.AIX.StreamChat(ctx, messages, func(chunk string) {
		streamFunc(dto.SSEEvent{
			Type:    constant.SSE_TYPE_CONTENT,
			Content: chunk,
		})
	})
	if err != nil {
		global.Log.Error(err)
		streamFunc(dto.SSEEvent{
			Type:  constant.SSE_TYPE_ERROR,
			Error: ErrChatStream.Error(),
		})
		return ErrChatStream
	}

	// 5. 保存对话历史
	now := time.Now().Unix()
	_ = l.aiRepo.SaveMessage(ctx, userID, req.SessionID, aix.ChatMessage{
		Role:      "user",
		Content:   req.Message,
		Images:    req.Images,
		Timestamp: now,
	})
	_ = l.aiRepo.SaveMessage(ctx, userID, req.SessionID, aix.ChatMessage{
		Role:      "assistant",
		Content:   fullResponse,
		Timestamp: now,
	})

	// 6. 发送完成事件
	streamFunc(dto.SSEEvent{
		Type:      constant.SSE_TYPE_DONE,
		SessionID: req.SessionID,
	})

	return nil
}

// UploadKnowledge 上传知识库
func (l *CommonLogic) UploadKnowledge(ctx context.Context, userID string, req dto.KnowledgeUploadReq) error {
	// 构建 CollectionName
	var collectionName string
	switch req.SpaceType {
	case constant.SPACE_TYPE_PRIVATE:
		collectionName = fmt.Sprintf(constant.COLLECTION_USER_PREFIX, userID)
	case constant.SPACE_TYPE_PUBLIC:
		collectionName = constant.COLLECTION_PUBLIC
	default:
		return ErrInvalidSpaceType
	}

	// 添加文档
	err := l.aiRepo.AddDocuments(ctx, collectionName, req.Content)
	if err != nil {
		global.Log.Error(err)
		return ErrKnowledgeUpload
	}
	return nil
}

// GetChatHistory 获取完整对话历史（供前端展示）
func (l *CommonLogic) GetChatHistory(ctx context.Context, userID string, req dto.ChatHistoryReq) (dto.ChatHistoryResp, error) {
	var resp dto.ChatHistoryResp

	history, err := l.aiRepo.GetFullHistory(ctx, userID, req.SessionID)
	if err != nil {
		global.Log.Error(err)
		return resp, ErrDefault
	}

	resp.Messages = make([]dto.ChatMessageItem, len(history))
	for i, msg := range history {
		resp.Messages[i] = dto.ChatMessageItem{
			Role:      msg.Role,
			Content:   msg.Content,
			Images:    msg.Images,
			Timestamp: msg.Timestamp,
		}
	}

	return resp, nil
}

func (l *CommonLogic) buildCollections(userID string, cfg dto.KBConfig) []string {
	var collections []string
	if !cfg.Enable {
		return collections
	}
	if cfg.SearchPrivate {
		collections = append(collections, fmt.Sprintf(constant.COLLECTION_USER_PREFIX, userID))
	}
	if cfg.SearchPublic {
		collections = append(collections, constant.COLLECTION_PUBLIC)
	}
	return collections
}
