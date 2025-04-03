package grpc

import (
	"net"

	"github.com/AthulKrishna2501/proto-repo/admin"
	"github.com/AthulKrishna2501/zyra-admin-service/internals/core/repository"
	"github.com/AthulKrishna2501/zyra-admin-service/internals/core/services"
	"github.com/AthulKrishna2501/zyra-admin-service/internals/logger"
	"google.golang.org/grpc"
)

func StartgRPCServer(AdminRepo repository.AdminRepository, log logger.Logger) error {
	go func() {
		lis, err := net.Listen("tcp", ":5005")
		if err != nil {
			log.Error("Failed to listen on port 5005: %v", err)
			return
		}

		grpcServer := grpc.NewServer(
			grpc.MaxRecvMsgSize(1024*1024*100),
			grpc.MaxSendMsgSize(1024*1024*100),
		)
		adminService := services.NewAdminService(AdminRepo, log)
		admin.RegisterAdminServiceServer(grpcServer, adminService)

		log.Info("gRPC Server started on port 5005")
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("Failed to serve gRPC: %v", err)
			return
		}
	}()

	return nil
}
