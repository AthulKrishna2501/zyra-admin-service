package repository

import (
	"context"
	"errors"
	"fmt"
	"log"

	adminModel "github.com/AthulKrishna2501/zyra-admin-service/internals/core/models"
	auth "github.com/AthulKrishna2501/zyra-auth-service/internals/core/models"
	"github.com/AthulKrishna2501/zyra-vendor-service/internals/core/models"
	"github.com/google/uuid"
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
	CreateCategory(ctx context.Context, name string) error
	DeleteRequest(ctx context.Context, vendorID string) error
	GetAdminDashboard(ctx context.Context) (*adminModel.DashboardStats, error)
}

func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &AdminStorage{
		DB: db,
	}
}

func (r *AdminStorage) UpdateCategoryRequestStatus(ctx context.Context, vendorID, categoryID, status string) error {
	err := r.DB.WithContext(ctx).Model(&auth.User{}).
		Where("user_id = ?", vendorID).
		Update("status", status)

	if err != nil {
		return errors.New("failed to updated user status")
	}

	err = r.DB.Where("vendor_id = ?", vendorID).Delete(&models.CategoryRequest{})

	if err != nil {
		return errors.New("failed to delete category request")
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
	result := r.DB.WithContext(ctx).Model(&auth.User{}).Where("user_id = ?", vendorID).Update("status", status)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no records updated, user_id %s not found", vendorID)
	}

	return nil
}

func (r *AdminStorage) AddVendorCategory(ctx context.Context, VendorID, CategoryID string) error {
	vendorUUID, err := uuid.Parse(VendorID)
	if err != nil {
		return fmt.Errorf("invalid vendor ID format: %v", err)
	}
	categoryUUID, err := uuid.Parse(CategoryID)
	if err != nil {
		return fmt.Errorf("invalid vendor ID format: %v", err)
	}

	vendorCategory := models.VendorCategory{
		VendorID:   vendorUUID,
		CategoryID: categoryUUID,
	}
	err = r.DB.WithContext(ctx).Create(&vendorCategory).Error
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

	if result.RowsAffected == 0 {
		return nil, errors.New("no requests in the queue")
	}

	return CatRequests, nil
}

func (r *AdminStorage) CreateCategory(ctx context.Context, name string) error {
	var category models.Category
	log.Print("Category to be added :", name)
	err := r.DB.WithContext(ctx).Where("category_name = ?", name).First(&category).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		newCategory := &models.Category{
			CategoryName: name,
		}

		if err := r.DB.WithContext(ctx).Create(newCategory).Error; err != nil {
			return err
		}
		return nil

	} else if err != nil {
		return err
	}

	return fmt.Errorf("category name '%s' already exists", name)

}

func (r *AdminStorage) DeleteRequest(ctx context.Context, vendorID string) error {
	if err := r.DB.WithContext(ctx).
		Where("vendor_id = ?", vendorID).
		Delete(&models.CategoryRequest{}).Error; err != nil {
		return fmt.Errorf("failed to delete category request for vendor_id %s", vendorID)
	}

	return nil

}

func (r *AdminStorage) GetAdminDashboard(ctx context.Context) (*adminModel.DashboardStats, error) {
	var stats adminModel.DashboardStats

	err := r.DB.WithContext(ctx).
		Raw(`
			SELECT 
				(SELECT COUNT(*) FROM users WHERE role = 'vendor') AS total_vendors,
				(SELECT COUNT(*) FROM users WHERE role = 'client') AS total_clients,
				(SELECT COUNT(*) FROM bookings) AS total_bookings,
				COALESCE(SUM(price), 0) AS total_revenue
			FROM bookings
		`).Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	return &stats, nil
}
