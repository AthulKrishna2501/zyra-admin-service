package services

import (
	"context"
	"fmt"
	"log"

	pb "github.com/AthulKrishna2501/proto-repo/admin"
	"github.com/AthulKrishna2501/zyra-admin-service/internals/app/config"
	"github.com/AthulKrishna2501/zyra-admin-service/internals/core/repository"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AdminService struct {
	pb.UnimplementedAdminServiceServer
	AdminRepo   repository.AdminRepository
	redisClient *redis.Client
}

func NewAdminService(AdminRepo repository.AdminRepository) *AdminService {
	return &AdminService{AdminRepo: AdminRepo, redisClient: config.RedisClient}
}

func (s *AdminService) ApproveRejectCategory(ctx context.Context, req *pb.ApproveRejectCategoryRequest) (*pb.ApproveRejectCategoryResponse, error) {
	log.Printf("Admin Service: Received gRPC request - VendorID=%s, CategoryID=%s, Status=%s",
		req.VendorId, req.CategoryId, req.Status)

	if req.VendorId == "" || req.CategoryId == "" || req.Status == "" {
		log.Println("Admin Service: ERROR - Missing required fields in gRPC request")
		return nil, status.Errorf(codes.InvalidArgument, "VendorID, CategoryID, and Status are required")
	}

	if req.Status != "approved" && req.Status != "rejected" {
		log.Println("Admin Service: ERROR - Invalid status value")
		return nil, status.Errorf(codes.InvalidArgument, "Invalid status. Allowed values: 'approved', 'rejected'")
	}

	err := s.AdminRepo.UpdateCategoryRequestStatus(ctx, req.VendorId, req.CategoryId, req.Status)
	if err != nil {
		log.Printf("Admin Service: ERROR - Failed to update category request: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to update category request: %v", err)
	}

	log.Printf("Admin Service: Successfully updated category request - VendorID=%s, CategoryID=%s, Status=%s",
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
