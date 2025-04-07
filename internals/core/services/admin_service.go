package services

import (
	"context"
	"fmt"

	pb "github.com/AthulKrishna2501/proto-repo/admin"
	"github.com/AthulKrishna2501/zyra-admin-service/internals/app/config"
	"github.com/AthulKrishna2501/zyra-admin-service/internals/core/repository"
	"github.com/AthulKrishna2501/zyra-admin-service/internals/logger"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		userList = append(userList, &pb.User{
			UserId:          user.UserID,
			Email:           user.Email,
			Role:            user.Role,
			IsBlocked:       user.IsBlocked,
			IsEmailVerified: user.IsEmailVerified,
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
	amount, err := s.AdminRepo.GetWalletBalance(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	return &pb.ViewAdminWalletResponse{
		Balance: float32(amount),
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
