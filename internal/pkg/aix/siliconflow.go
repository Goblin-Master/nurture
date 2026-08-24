package aix

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"nurture/internal/config"
	"strings"
)

const defaultSiliconFlowBaseURL = "https://api.siliconflow.cn/v1"

type SiliconFlowEmbedder struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

func NewSiliconFlowEmbedder(cfg config.EmbeddingModel) *SiliconFlowEmbedder {
	return &SiliconFlowEmbedder{
		apiKey:  cfg.APIKey,
		baseURL: normalizeSiliconFlowEmbeddingURL(cfg.BaseURL),
		model:   cfg.Model,
		client:  http.DefaultClient,
	}
}

func (e *SiliconFlowEmbedder) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	results := make([][]float32, 0, len(texts))
	for _, text := range texts {
		vec, err := e.embedOne(ctx, text)
		if err != nil {
			return nil, err
		}
		results = append(results, vec)
	}
	return results, nil
}

func (e *SiliconFlowEmbedder) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	return e.embedOne(ctx, text)
}

func (e *SiliconFlowEmbedder) embedOne(ctx context.Context, text string) ([]float32, error) {
	payload := map[string]any{
		"model": e.model,
		"input": text,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+e.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("siliconflow embedding api error: status=%s body=%s", resp.Status, strings.TrimSpace(string(b)))
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
		return nil, fmt.Errorf("siliconflow embedding api returned no data")
	}
	return res.Data[0].Embedding, nil
}

func normalizeSiliconFlowEmbeddingURL(baseURL string) string {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base == "" {
		base = defaultSiliconFlowBaseURL
	}
	if strings.HasSuffix(base, "/embeddings") {
		return base
	}
	if strings.HasSuffix(base, "/embedding") {
		return strings.TrimSuffix(base, "/embedding") + "/embeddings"
	}
	return base + "/embeddings"
}
