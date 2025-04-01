package models

import (
	"time"

	"github.com/google/uuid"
)

type AdminWallet struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
	Email     string    `json:"email"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
