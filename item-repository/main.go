package main

import (
	"context"
	"item-repository/config"
	"item-repository/handler"
	inventory "item-repository/pb"
	"item-repository/registry"
	"item-repository/repository"
	"item-repository/router"
	"item-repository/service"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func main() {
	config.LoadConfig()
	db, err := config.ConnectDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		return
	}
	defer config.CloseDatabase(db)

	repo := repository.NewInventoryRepository(db)
	svc := service.NewItemService(repo)
	hde := handler.NewInventoryHandler(svc)
	grpcHandler := handler.NewInventoryGrpcHandler(repo)

	r := gin.Default()
	router.RegisterRoutes(r, hde)
	log.Println("Starting HTTP server on ", ":"+config.APPConfig.ServicePort)
	go func() {
		if err := r.Run(":" + config.APPConfig.ServicePort); err != nil {
			log.Fatalf("Failed to run HTTP server: %v", err)
		}
	}()

	grpcServer := grpc.NewServer()
	go func() {
		lis, err := net.Listen("tcp", ":"+config.APPConfig.GRPCPort)
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}
		inventory.RegisterInventoryServiceServer(grpcServer, grpcHandler)
		log.Println("Starting gRPC server on ", config.APPConfig.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC server: %v", err)
		}
	}()

	consumerGroup, err := config.NewKafkaConsumer()
	if err != nil {
		log.Fatalf("Failed to start Kafka consumer: %v", err)
		return
	}
	defer consumerGroup.Close()

	ctx, cancel := context.WithCancel(context.Background())
	topics := []string{"orders-created"}

	go func() {
		for {
			log.Printf("Starting Kafka consumer for topics: %v", topics)
			if err := consumerGroup.Consume(ctx, topics, hde); err != nil {
				log.Printf("Error consuming messages: %v", err)
			}
			if ctx.Err() != nil {
				log.Printf("Kafka consumer context error: %v", ctx.Err())
				return
			}
		}
	}()

	reg, err := registry.NewServiceRegistry([]string{config.APPConfig.EtcdEndpoints})
	if err != nil {
		log.Fatalf("Failed to connect to etcd: %v", err)
	}

	err = reg.Register("inventory-grpc-service", config.APPConfig.PODIP+":"+config.APPConfig.GRPCPort, 5)
	if err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	<-sigterm
	log.Println("Shutting down inventory service...")
	if err := reg.Deregister(); err != nil {
		log.Printf("Failed to deregister service: %v", err)
	}
	grpcServer.GracefulStop()
	cancel()
}
