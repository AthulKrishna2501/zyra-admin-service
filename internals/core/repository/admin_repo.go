package repository

import (
	"context"
	"errors"

	"github.com/AthulKrishna2501/zyra-vendor-service/internals/core/models"
	"gorm.io/gorm"
)

type User struct {
	UserID          string `gorm:"type:uuid"`
	Email           string
	Role            string
	IsBlocked       bool
	IsEmailVerified bool
}

type AdminStorage struct {
	DB *gorm.DB
}

type AdminRepository interface {
	UpdateCategoryRequestStatus(ctx context.Context, vendorID, categoryID, status string) error
	GetAllUsers(ctx context.Context) ([]User, error)
}

func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &AdminStorage{
		DB: db,
	}
}

func (r *AdminStorage) UpdateCategoryRequestStatus(ctx context.Context, vendorID, categoryID, status string) error {
	result := r.DB.WithContext(ctx).
		Model(&models.CategoryRequest{}).
		Where("vendor_id = ? AND category_id = ?", vendorID, categoryID).
		Update("status", status)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("category request not found")
	}
	return nil
}

func (r *AdminStorage) GetAllUsers(ctx context.Context) ([]User, error) {
	var users []User
	result := r.DB.WithContext(ctx).
		Select("user_id, email, role, is_blocked, is_email_verified").
		Find(&users)

	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}
