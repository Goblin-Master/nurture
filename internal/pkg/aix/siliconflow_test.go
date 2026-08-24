package aix

import "testing"

func TestNormalizeSiliconFlowEmbeddingURL(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "base v1",
			in:   "https://api.siliconflow.cn/v1",
			want: "https://api.siliconflow.cn/v1/embeddings",
		},
		{
			name: "full embeddings endpoint",
			in:   "https://api.siliconflow.cn/v1/embeddings",
			want: "https://api.siliconflow.cn/v1/embeddings",
		},
		{
			name: "legacy singular endpoint",
			in:   "https://api.siliconflow.cn/v1/embedding",
			want: "https://api.siliconflow.cn/v1/embeddings",
		},
		{
			name: "empty defaults to siliconflow",
			in:   "",
			want: "https://api.siliconflow.cn/v1/embeddings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeSiliconFlowEmbeddingURL(tt.in); got != tt.want {
				t.Fatalf("normalizeSiliconFlowEmbeddingURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
