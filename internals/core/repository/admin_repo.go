package repository

import (
	"context"
	"errors"
	"log"

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
	UpdateRequestStatus(ctx context.Context, vendorID, status string) error
	GetAllUsers(ctx context.Context) ([]User, error)
	AddVendorCategory(ctx context.Context, VendorID, CategoryID string) error
	GetRequests(ctx context.Context) ([]models.CategoryRequest, error)
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

func (r *AdminStorage) UpdateRequestStatus(ctx context.Context, vendorID, status string) error {
	if err := r.DB.WithContext(ctx).Update("status", true).Where("user_id=?", vendorID); err != nil {
		return err.Error
	}

	return nil
}

func (r *AdminStorage) AddVendorCategory(ctx context.Context, VendorID, CategoryID string) error {
	vendorCategory := models.VendorCategory{
		VendorID:   VendorID,
		CategoryID: CategoryID,
	}
	err := r.DB.WithContext(ctx).Create(&vendorCategory).Error
	if err != nil {
		log.Printf("Error adding vendor category: %v", err)
		return err
	}

	log.Println("Vendor category added successfully")
	return nil
}

func (r *AdminStorage) GetRequests(ctx context.Context) ([]models.CategoryRequest, error) {
	var CatRequests []models.CategoryRequest
	result := r.DB.WithContext(ctx).Select("vendor_id,category_id").Find(&CatRequests)

	if result.Error != nil {
		return nil, result.Error
	}

	return CatRequests, nil
}
