package aix

import "errors"

var (
	ErrChatDisabled        = errors.New("aix chat is disabled")
	ErrEmbeddingDisabled   = errors.New("aix embedding is disabled")
	ErrVectorStoreDisabled = errors.New("aix vector store is disabled")
)
