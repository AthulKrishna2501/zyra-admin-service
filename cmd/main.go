package main

import (
	"github.com/AthulKrishna2501/zyra-admin-service/internals/app/config"
	"github.com/AthulKrishna2501/zyra-admin-service/internals/app/grpc"
	"github.com/AthulKrishna2501/zyra-admin-service/internals/core/database"
	"github.com/AthulKrishna2501/zyra-admin-service/internals/core/repository"
	"github.com/AthulKrishna2501/zyra-admin-service/internals/logger"
	"github.com/gin-gonic/gin"
)

func main() {
	log := logger.NewLogrusLogger()
	configEnv, err := config.LoadConfig()
	if err != nil {
		log.Error("Error in config .env: %v", err)
		return
	}

	config.InitRedis()

	db := database.ConnectDatabase(configEnv)
	if db == nil {
		log.Error("Failed to connect to database")
		return
	}

	AdminRepo := repository.NewAdminRepository(db)

	err = grpc.StartgRPCServer(AdminRepo, log)

	if err != nil {
		log.Error("Failed to start gRPC server", err.Error())
		return
	}

	router := gin.Default()
	log.Info("HTTP Server started on port 3006")
	router.Run(":3006")

}
