package aix

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"nurture/internal/config"
	"strings"
)

type ZhipuEmbedder struct {
	apiKey  string
	baseURL string
	model   string
}

func NewZhipuEmbedder(cfg config.EmbeddingModel) *ZhipuEmbedder {
	return &ZhipuEmbedder{
		apiKey:  cfg.APIKey,
		baseURL: cfg.BaseURL,
		model:   cfg.Model,
	}
}

// EmbedDocuments 实现 embeddings.Embedder 接口
func (e *ZhipuEmbedder) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	var results [][]float32
	for _, text := range texts {
		vec, err := e.embedOne(ctx, text)
		if err != nil {
			return nil, err
		}
		results = append(results, vec)
	}
	return results, nil
}

func (e *ZhipuEmbedder) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	return e.embedOne(ctx, text)
}

func (e *ZhipuEmbedder) embedOne(ctx context.Context, text string) ([]float32, error) {
	// 确保 URL 正确拼接
	baseURL := strings.TrimRight(e.baseURL, "/")
	url := fmt.Sprintf("%s/embeddings", baseURL)

	payload := map[string]interface{}{
		"model": e.model,
		"input": text,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+e.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// 读取错误信息
		var errRes map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errRes)
		return nil, fmt.Errorf("api error: %s, detail: %v", resp.Status, errRes)
	}

	var res struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	if len(res.Data) == 0 {
		return nil, fmt.Errorf("no embedding data")
	}
	return res.Data[0].Embedding, nil
}
