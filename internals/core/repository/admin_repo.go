package repository

import (
	"context"
	"errors"
	"fmt"
	"log"

	adminModel "github.com/AthulKrishna2501/zyra-admin-service/internals/core/models"
	auth "github.com/AthulKrishna2501/zyra-auth-service/internals/core/models"
	clientModel "github.com/AthulKrishna2501/zyra-client-service/internals/core/models"
	"github.com/AthulKrishna2501/zyra-vendor-service/internals/core/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminStorage struct {
	DB *gorm.DB
}

type AdminRepository interface {
	UpdateCategoryRequestStatus(ctx context.Context, vendorID, categoryID, status string) error
	UpdateRequestStatus(ctx context.Context, vendorID, status string) error
	GetAllUsers(ctx context.Context) ([]adminModel.UserInfo, error)
	ListCategories(ctx context.Context) ([]models.Category, error)
	AddVendorCategory(ctx context.Context, VendorID, CategoryID string) error
	GetRequests(ctx context.Context) ([]models.CategoryRequest, error)
	CreateCategory(ctx context.Context, name string) error
	DeleteRequest(ctx context.Context, vendorID string) error
	GetAdminDashboard(ctx context.Context) (*adminModel.DashboardStats, error)
	GetAdminWallet(ctx context.Context, email string) (*adminModel.AdminWallet, error)
	GetAllBookings(ctx context.Context) ([]adminModel.Booking, error)
	GetAllAdminTransactions(ctx context.Context) ([]adminModel.AdminWalletTransaction, error)
	GetAllFundReleaseRequests(ctx context.Context) ([]adminModel.FundRelease, error)
	UpdateFundReleaseStatus(ctx context.Context, requestID, status string) error
	GetEventDetails(ctx context.Context, requestID string) (*adminModel.EventDetails, error)
	GetUserIDWithEventID(ctx context.Context, eventID string) (string, error)
	CreateTransaction(ctx context.Context, newTransaction *clientModel.Transaction) error
	CreditAmountToClientWallet(ctx context.Context, amount float64, userID string) error
	DebitAmountFromAdminWallet(ctx context.Context, amount float64, adminEmail string) error
	CreateAdminWalletTransaction(ctx context.Context, newAdminWalletTransaction *adminModel.AdminWalletTransaction) error
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

func (r *AdminStorage) GetAllUsers(ctx context.Context) ([]adminModel.UserInfo, error) {
	var users []adminModel.UserInfo

	result := r.DB.WithContext(ctx).
		Table("users").
		Select(`
			users.user_id,
			users.email,
			users.role,
			users.is_blocked,
			user_details.first_name,
			user_details.last_name
		`).
		Joins("JOIN user_details ON user_details.user_id = users.user_id").
		Scan(&users)

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
	result := r.DB.WithContext(ctx).
		Joins("JOIN categories c ON c.category_id = category_requests.category_id").
		Joins("JOIN user_details u ON u.user_id = category_requests.vendor_id").
		Select("category_requests.vendor_id, category_requests.category_id, category_requests.created_at,c.category_name as category_name, u.first_name as vendor_name").
		Find(&CatRequests)

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

func (r *AdminStorage) GetAdminWallet(ctx context.Context, email string) (*adminModel.AdminWallet, error) {
	var wallet adminModel.AdminWallet

	err := r.DB.Where("email = ?", email).First(&wallet).Error

	if err != nil {
		return nil, err
	}

	return &wallet, nil
}

func (r *AdminStorage) ListCategories(ctx context.Context) ([]models.Category, error) {
	var categories []models.Category

	err := r.DB.Statement.DB.WithContext(ctx).
		Find(&categories).Error

	if err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *AdminStorage) GetAllBookings(ctx context.Context) ([]adminModel.Booking, error) {
	var bookings []adminModel.Booking

	query := `
        SELECT 
            b.booking_id,
            b.client_id,
            b.vendor_id,
            b.service,
            b.date,
            b.price,
            b.status,
            c.first_name AS client_first_name,
            c.last_name AS client_last_name,
            v.first_name AS vendor_first_name,
            v.last_name AS vendor_last_name
        FROM bookings b
        JOIN user_details c ON b.client_id = c.user_id
        JOIN user_details v ON b.vendor_id = v.user_id
    `

	err := r.DB.WithContext(ctx).Raw(query).Scan(&bookings).Error
	if err != nil {
		return nil, err
	}

	return bookings, nil
}

func (r *AdminStorage) GetAllAdminTransactions(ctx context.Context) ([]adminModel.AdminWalletTransaction, error) {
	var transactions []adminModel.AdminWalletTransaction

	err := r.DB.WithContext(ctx).
		Find(&transactions).Error

	if err != nil {
		return nil, err
	}

	return transactions, nil
}

func (r *AdminStorage) GetAllFundReleaseRequests(ctx context.Context) ([]adminModel.FundRelease, error) {
	var requests []adminModel.FundRelease
	err := r.DB.WithContext(ctx).Find(&requests).Error
	if err != nil {
		return nil, err
	}

	return requests, nil

}

func (r *AdminStorage) UpdateFundReleaseStatus(ctx context.Context, requestID, status string) error {
	err := r.DB.WithContext(ctx).Model(adminModel.FundRelease{}).Where("request_id = ?", requestID).Update("status", status).Error
	if err != nil {
		return err
	}

	return nil

}

func (r *AdminStorage) GetEventDetails(ctx context.Context, requestID string) (*adminModel.EventDetails, error) {
	var eventDetails adminModel.EventDetails
	err := r.DB.WithContext(ctx).Model(&adminModel.FundRelease{}).Where("request_id = ?", requestID).Scan(&eventDetails).Error
	if err != nil {
		return nil, err
	}

	return &eventDetails, nil

}
func (r *AdminStorage) GetUserIDWithEventID(ctx context.Context, eventID string) (string, error) {
	var HostedBy string
	err := r.DB.WithContext(ctx).
		Model(&clientModel.Event{}).
		Select("hosted_by").
		Where("event_id = ?", eventID).
		Scan(&HostedBy).Error

	if err != nil {
		return "", err
	}

	return HostedBy, nil
}

func (r *AdminStorage) CreateTransaction(ctx context.Context, newTransaction *clientModel.Transaction) error {
	return r.DB.WithContext(ctx).Create(newTransaction).Error

}

func (r *AdminStorage) CreditAmountToClientWallet(ctx context.Context, amount float64, userID string) error {
	return r.DB.
		Model(&models.Wallet{}).Where("client_id = ?", userID).
		Updates(map[string]interface{}{
			"wallet_balance":        gorm.Expr("wallet_balance + ?", amount),
			"total_deposits": gorm.Expr("total_deposits + ?", amount),
		}).Error

}

func (r *AdminStorage) CreateAdminWalletTransaction(ctx context.Context, newAdminWalletTransaction *adminModel.AdminWalletTransaction) error {
	return r.DB.WithContext(ctx).Create(newAdminWalletTransaction).Error
}

func (r *AdminStorage) DebitAmountFromAdminWallet(ctx context.Context, amount float64, adminEmail string) error {
	return r.DB.
		Model(&adminModel.AdminWallet{}).Where("email = ?", adminEmail).
		Updates(map[string]interface{}{
			"balance":        gorm.Expr("balance - ?", amount),
			"total_deposits": gorm.Expr("total_withdrawals + ?", amount),
		}).Error
}
