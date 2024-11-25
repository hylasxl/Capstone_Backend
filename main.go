package main

import (
	"Capstone_Go_gRPC/configs"
	"Capstone_Go_gRPC/pkg/pb/authpb"
	"Capstone_Go_gRPC/pkg/pb/friendpb"
	"Capstone_Go_gRPC/pkg/pb/userAccountpb"
	"Capstone_Go_gRPC/pkg/service"
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"gorm.io/gorm"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type server struct {
	authpb.UnimplementedAuthServiceServer
	userAccountpb.UnimplementedUserAccountServer
	friendpb.UnimplementedFriendServiceServer
	DB *gorm.DB
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := godotenv.Load(); err != nil {
		log.Println("Warning: Couldn't load .env file, relying on system environment variables")
	}
	fixedIP := string("")
	lis, err := net.Listen("tcp", fixedIP+":8080")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()

	logServerIP()
	cloudinaryService, err := configs.InitCloudinary(ctx)
	if err != nil {
		log.Fatalf("Failed to init cloudinary: %v", err)
	}
	if err := cloudinaryService.CheckConnection(); err != nil {
		log.Fatalf("Failed to check connection with cloudinary: %v", err)
	} else {
		log.Println("Cloudinary connected successfully")
	}
	mysqlDB, err := configs.ConnectMySQL()
	if err != nil {
		log.Fatalf("Could not connect to MySQL: %v", err)
	}

	authSvc := &service.AuthServiceServer{DB: mysqlDB, CloudinaryClient: cloudinaryService}
	authpb.RegisterAuthServiceServer(s, authSvc)
	userAccountSvc := &service.UserAccountServiceServer{DB: mysqlDB}
	userAccountpb.RegisterUserAccountServer(s, userAccountSvc)
	friendSvc := &service.FriendServiceServer{DB: mysqlDB}
	friendpb.RegisterFriendServiceServer(s, friendSvc)

	go func() {
		log.Println("Starting gRPC server on " + fixedIP + ":8080")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	fmt.Println("Shutting down server...")
	s.GracefulStop()
	fmt.Println("Server stopped gracefully")
}

func logServerIP() {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatalf("failed to get network interfaces: %v", err)
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			fmt.Printf("Server is running on IP: %s\n", ipNet.IP.String())
		}
	}
}
