package services

import (
	"context"
	"fmt"
	"time"

	pb "github.com/AthulKrishna2501/proto-repo/admin"
	"github.com/AthulKrishna2501/zyra-admin-service/internals/app/config"
	adminModel "github.com/AthulKrishna2501/zyra-admin-service/internals/core/models"
	"github.com/AthulKrishna2501/zyra-admin-service/internals/core/repository"
	"github.com/AthulKrishna2501/zyra-admin-service/internals/logger"
	"github.com/AthulKrishna2501/zyra-client-service/internals/core/models"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AdminService struct {
	pb.UnimplementedAdminServiceServer
	AdminRepo   repository.AdminRepository
	redisClient *redis.Client
	log         logger.Logger
}

func NewAdminService(AdminRepo repository.AdminRepository, logger logger.Logger) *AdminService {
	return &AdminService{AdminRepo: AdminRepo, redisClient: config.RedisClient, log: logger}
}

func (s *AdminService) ApproveRejectCategory(ctx context.Context, req *pb.ApproveRejectCategoryRequest) (*pb.ApproveRejectCategoryResponse, error) {
	s.log.Info("Admin Service: Received gRPC request - VendorID=%s, CategoryID=%s, Status=%s",
		req.VendorId, req.CategoryId, req.Status)

	if req.VendorId == "" || req.CategoryId == "" || req.Status == "" {
		return nil, status.Errorf(codes.InvalidArgument, "VendorID, CategoryID, and Status are required")
	}

	if req.Status != "approved" && req.Status != "rejected" {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid status. Allowed values: 'approved', 'rejected'")
	}

	if req.Status == "approved" {
		if err := s.AdminRepo.AddVendorCategory(ctx, req.VendorId, req.CategoryId); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	err := s.AdminRepo.UpdateRequestStatus(ctx, req.VendorId, req.Status)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update status: %v", err)
	}

	err = s.AdminRepo.DeleteRequest(ctx, req.VendorId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete category request %v", err)
	}

	s.log.Info("Admin Service: Successfully updated category request - VendorID=%s, CategoryID=%s, Status=%s",
		req.VendorId, req.CategoryId, req.Status)

	return &pb.ApproveRejectCategoryResponse{
		Message: fmt.Sprintf("Category request has been %s", req.Status),
	}, nil

}

func (s *AdminService) BlockUser(ctx context.Context, req *pb.BlockUnblockUserRequest) (*pb.BlockUnblockUserResponse, error) {
	if req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "User ID cannot be empty")
	}

	err := s.redisClient.SAdd(ctx, "blocked_users", req.UserId).Err()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to block user: %v", err)
	}

	return &pb.BlockUnblockUserResponse{
		Message: fmt.Sprintf("User %s has been blocked", req.UserId),
	}, nil
}

func (s *AdminService) UnblockUser(ctx context.Context, req *pb.BlockUnblockUserRequest) (*pb.BlockUnblockUserResponse, error) {
	if req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "User ID cannot be empty")
	}

	err := s.redisClient.SRem(ctx, "blocked_users", req.UserId).Err()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to unblock user: %v", err)
	}

	return &pb.BlockUnblockUserResponse{
		Message: fmt.Sprintf("User %s has been unblocked", req.UserId),
	}, nil
}

func (s *AdminService) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	users, err := s.AdminRepo.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}

	var userList []*pb.User
	for _, user := range users {
		FullName := user.FirstName + " " + user.LastName
		userList = append(userList, &pb.User{
			UserId:    user.UserId,
			Name:      FullName,
			Email:     user.Email,
			Role:      user.Role,
			IsBlocked: user.IsBlocked,
		})
	}

	return &pb.ListUsersResponse{Users: userList}, nil
}

func (s *AdminService) ViewRequests(ctx context.Context, req *pb.ViewRequestsReq) (*pb.ViewRequestsResponse, error) {
	requests, err := s.AdminRepo.GetRequests(ctx)
	if err != nil {
		return nil, err
	}

	var pbRequests []*pb.CategoryRequest
	for _, r := range requests {
		pbRequests = append(pbRequests, &pb.CategoryRequest{
			VendorId:   r.VendorID.String(),
			CategoryId: r.CategoryID.String(),
			Name:       r.CategoryName,
			VendorName: r.VendorName,
			Date:       r.CreatedAt.String(),
		})

	}

	return &pb.ViewRequestsResponse{
		Requests: pbRequests,
	}, nil
}

func (s *AdminService) AddCategory(ctx context.Context, req *pb.AddCategoryRequest) (*pb.AddCategoryResponse, error) {
	err := s.AdminRepo.CreateCategory(ctx, req.CategoryName)
	if err != nil {
		return nil, err
	}

	return &pb.AddCategoryResponse{
		Message: "category added successfully",
	}, nil
}

func (s *AdminService) AdminDashBoard(ctx context.Context, req *pb.AdminDashBoardRequest) (*pb.AdminDashBoardResponse, error) {
	stats, err := s.AdminRepo.GetAdminDashboard(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.AdminDashBoardResponse{
		TotalVendors:  stats.TotalVendors,
		TotalClients:  stats.TotalClients,
		TotalBookings: stats.TotalBookings,
		TotalRevenue:  stats.TotalRevenue,
	}, nil
}

func (s *AdminService) ViewAdminWallet(ctx context.Context, req *pb.ViewAdminWalletRequest) (*pb.ViewAdminWalletResponse, error) {
	wallet, err := s.AdminRepo.GetAdminWallet(ctx, req.Email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve admin wallet %v", err.Error())
	}

	return &pb.ViewAdminWalletResponse{
		Balance:          float32(wallet.Balance),
		TotalDeposits:    float32(wallet.TotalDeposits),
		TotalWithdrawals: float32(wallet.TotalWithdrawals),
	}, nil
}

func (s *AdminService) GetAdminWalletTransactions(ctx context.Context, req *pb.GetAdminTransactionRequest) (*pb.GetAdminTransactionResponse, error) {
	walletTransactions, err := s.AdminRepo.GetAllAdminTransactions(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve admin wallet transactions: %v", err.Error())
	}

	var protoTransactions []*pb.AdminWalletTransaction
	for _, txn := range walletTransactions {
		protoTransactions = append(protoTransactions, &pb.AdminWalletTransaction{
			TransactionId: txn.TransactionID.String(),
			Date:          txn.Date.String(),
			Amount:        float32(txn.Amount),
			Type:          txn.Type,
			Status:        txn.Status,
		})
	}

	return &pb.GetAdminTransactionResponse{
		WalletTransactions: protoTransactions,
	}, nil
}

func (s *AdminService) ListCategory(ctx context.Context, req *pb.ListCategoryRequest) (*pb.ListCategoryResponse, error) {
	categories, err := s.AdminRepo.ListCategories(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to fetch categories: %v", err)
	}

	var categoryResponses []*pb.Category
	for _, cat := range categories {
		categoryResponses = append(categoryResponses, &pb.Category{
			CategoryId:   cat.CategoryID.String(),
			CategoryName: cat.CategoryName,
		})
	}

	return &pb.ListCategoryResponse{
		Categories: categoryResponses,
	}, nil
}

func (s *AdminService) GetAllBookings(ctx context.Context, req *pb.GetAllBookingsRequest) (*pb.GetAllBookingsResponse, error) {
	bookings, err := s.AdminRepo.GetAllBookings(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to fetch bookings: %v", err)
	}

	var pbBookings []*pb.Booking
	for _, booking := range bookings {
		pbBookings = append(pbBookings, &pb.Booking{
			BookingId: booking.BookingID.String(),
			Client: &pb.Client{
				FirstName: booking.Client.FirstName,
				LastName:  booking.Client.LastName,
			},
			Vendor: &pb.Vendor{
				FirstName: booking.Vendor.FirstName,
				LastName:  booking.Vendor.LastName,
			},
			Service: booking.Service,
			Date:    timestamppb.New(booking.Date),
			Price:   int32(booking.Price),
			Status:  booking.Status,
		})
	}

	return &pb.GetAllBookingsResponse{
		Bookings: pbBookings,
	}, nil
}

func (s *AdminService) GetFundRelease(ctx context.Context, req *pb.FundReleaseRequest) (*pb.FundReleaseResponse, error) {
	requests, err := s.AdminRepo.GetAllFundReleaseRequests(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch fund release requests %v:", err)
	}

	var pbRequest []*pb.FundReleaseRequests
	for _, req := range requests {
		pbRequest = append(pbRequest, &pb.FundReleaseRequests{
			RequestId: req.RequestID.String(),
			EventId:   req.EventID.String(),
			EventName: req.EventName,
			Amount:    float32(req.Amount),
			Tickets:   uint32(req.Tickets),
			Status:    req.Status,
		})
	}

	return &pb.FundReleaseResponse{
		Requests: pbRequest,
	}, nil

}

func (s *AdminService) ApproveFundRelease(ctx context.Context, req *pb.ApproveFundReleaseRequest) (*pb.ApproveFundReleaseResponse, error) {
	requestID := req.GetRequestId()
	newStatus := req.GetStatus()

	if newStatus == "approved" {
		details, err := s.AdminRepo.GetEventDetails(ctx, requestID)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to fetch event details %v :", err)
		}

		userID, err := s.AdminRepo.GetUserIDWithEventID(ctx, details.EventID)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to fetch userID %v", err)
		}

		userUUID, err := uuid.Parse(userID)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to parse user_id %v", err)
		}

		newTransaction := &models.Transaction{
			UserID:        userUUID,
			Purpose:       "Fund Release",
			AmountPaid:    int(details.Amount),
			PaymentMethod: "wallet",
			DateOfPayment: time.Now(),
			PaymentStatus: "refunded",
		}

		newAdminWalletTransaction := &adminModel.AdminWalletTransaction{
			Date:   time.Now(),
			Type:   "Fund Release",
			Amount: details.Amount,
			Status: "succeeded",
		}

		err = s.AdminRepo.DebitAmountFromAdminWallet(ctx, details.Amount, "admin@example.com")
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to debit amount from admin wallet %v ", err)
		}

		err = s.AdminRepo.CreateAdminWalletTransaction(ctx, newAdminWalletTransaction)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to create admin wallet transaction")
		}
		err = s.AdminRepo.CreditAmountToClientWallet(ctx, details.Amount, userID)

		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to credit amount to client wallet %v", err)
		}

		err = s.AdminRepo.CreateTransaction(ctx, newTransaction)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to create transaction: %v", err)
		}

	}

	err := s.AdminRepo.UpdateFundReleaseStatus(ctx, requestID, newStatus)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch fund release requests %v:", err)
	}

	return &pb.ApproveFundReleaseResponse{
		Message: fmt.Sprintf("Fund release request %s has been %s", requestID, newStatus),
	}, nil

}
