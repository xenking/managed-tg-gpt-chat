package adapters

import (
	"sync"

	"github.com/google/uuid"

	"github.com/xenking/managed-tg-gpt-chat/internal/domain"
)

type InMemoryTokenStorage struct {
	tokens map[uuid.UUID][]byte
	mu     sync.Mutex
}

var _ domain.ActivitiesTokenStorage = (*InMemoryTokenStorage)(nil)

func NewInMemoryTokenStorage() *InMemoryTokenStorage {
	return &InMemoryTokenStorage{tokens: make(map[uuid.UUID][]byte)}
}

func (ts *InMemoryTokenStorage) Store(token []byte) uuid.UUID {
	id := uuid.New()
	ts.mu.Lock()
	ts.tokens[id] = token
	ts.mu.Unlock()
	return id
}

func (ts *InMemoryTokenStorage) Pop(token uuid.UUID) ([]byte, bool) {
	ts.mu.Lock()
	data, ok := ts.tokens[token]
	if ok {
		delete(ts.tokens, token)
	}
	ts.mu.Unlock()
	return data, ok
}
