package models

import (
	"time"

	"github.com/AthulKrishna2501/zyra-auth-service/internals/core/models"
	"github.com/google/uuid"
)

type AdminWallet struct {
	Email            string    `json:"email"`
	Balance          float64   `gorm:"default:0" json:"balance"`
	TotalDeposits    float64   `gorm:"default:0" json:"total_deposits"`
	TotalWithdrawals float64   `gorm:"default:0" json:"total_withdrawals"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type AdminWalletTransaction struct {
	TransactionID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()"`
	Date          time.Time `gorm:"date"`
	Type          string    `gorm:"type:varchar(255)"`
	Amount        float64
	Status        string `gorm:"type:varchar(255)"`
}

type DashboardStats struct {
	TotalVendors  int32
	TotalClients  int32
	TotalBookings int32
	TotalRevenue  int64
}

type Booking struct {
	ID               uuid.UUID          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	BookingID        uuid.UUID          `gorm:"type:uuid;default:gen_random_uuid()"`
	ClientID         uuid.UUID          `gorm:"type:uuid;not null"`
	Client           models.UserDetails `gorm:"foreignKey:ClientID;references:UserID"`
	VendorID         uuid.UUID          `gorm:"type:uuid;not null"`
	Vendor           models.UserDetails `gorm:"foreignKey:VendorID;references:UserID"`
	Service          string             `gorm:"type:varchar(255)"`
	Date             time.Time          `gorm:"type:date;not null"`
	Status           string             `gorm:"type:varchar(50);not null"`
	Price            int                `gorm:"not null"`
	IsVendorApproved bool
	IsClientApproved bool
	IsFundReleased   bool
	CreatedAt        time.Time `gorm:"autoCreateTime"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime"`
}

type UserInfo struct {
	UserId    string `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	IsBlocked bool   `json:"is_blocked"`
}

type FundRelease struct {
	RequestID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()"`
	EventID   uuid.UUID `gorm:"type:uuid"`
	EventName string    `gorm:"typevarchar(255)"`
	Amount    float64
	Tickets   uint
	Status    string    `gorm:"type:varchar(255)"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type EventDetails struct {
	EventID string  `json:"event_id"`
	Amount  float64 `json:"amount"`
}
