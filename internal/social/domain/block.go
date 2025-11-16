package domain

import (
	"time"

	"github.com/0xsj/result"
	"github.com/google/uuid"
)

// Block represents a block relationship between two users.
type Block struct {
	ID        string
	BlockerID string
	BlockedID string
	CreatedAt time.Time
	Reason    *string
}

// NewBlock creates a new block relationship.
func NewBlock(blockerID, blockedID string, reason *string) result.Result[*Block] {
	// Validate: cannot block yourself
	if blockerID == blockedID {
		return result.Err[*Block](ErrCannotBlockSelf())
	}

	// Validate: valid user IDs
	if blockerID == "" || blockedID == "" {
		return result.Err[*Block](ErrInvalidUserID())
	}

	block := &Block{
		ID:        uuid.New().String(),
		BlockerID: blockerID,
		BlockedID: blockedID,
		CreatedAt: time.Now().UTC(),
		Reason:    reason,
	}

	return result.Ok(block)
}

// UpdateReason updates the block reason.
func (b *Block) UpdateReason(reason string) {
	b.Reason = &reason
}
