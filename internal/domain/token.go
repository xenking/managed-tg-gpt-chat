package domain

import "github.com/google/uuid"

type ActivitiesTokenStorage interface {
	Store(token []byte) uuid.UUID
	Pop(tokenID uuid.UUID) ([]byte, bool)
}
