package main

import (
	"log"
	"net/http"
	"pvz-service/internal/database"
	"pvz-service/internal/handlers"
	"pvz-service/internal/logger"
	"pvz-service/internal/middleware"
	"pvz-service/internal/repository"
	"pvz-service/internal/service"

	"pvz-service/internal/grpc"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {

	logger.Init("development")

	logger.Log.Info("Starting application...")

	db := database.SetupDB()

	logger.Log.Info("Database connected")

	router := gin.Default()
	repo := repository.NewRepository(db)

	go func() {
		logger.Log.Info("Starting gRPC server on :3000")
		grpc.StartGRPCServer(repo)
	}()

	go startMetricsServer()

	svc := service.NewService(repo)
	handler := handlers.NewHandler(svc)

	router.Use(middleware.PrometheusMiddleware())
	router.Use(middleware.LoggerMiddleware())

	router.POST("/dummyLogin", handler.DummyLogin)
	router.POST("/register", handler.Register)
	router.POST("/login", handler.Login)

	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.POST("/pvz", middleware.RequireModerator(), handler.CreatePVZ)
		protected.GET("/pvz", handler.GetPVZList)
		protected.POST("/receptions", middleware.RequireEmployee(), handler.CreateReception)
		protected.POST("/products", middleware.RequireEmployee(), handler.AddProduct)
		protected.POST("/pvz/:pvzId/close_last_reception", middleware.RequireEmployee(), handler.CloseReception)
		protected.POST("/pvz/:pvzId/delete_last_product", middleware.RequireEmployee(), handler.DeleteLastProduct)
	}

	logger.Log.Info("HTTP server starting on :8080")
	if err := router.Run(":8080"); err != nil {
		logger.Log.Fatal("Failed to start HTTP server:", err)
	}
}

func startMetricsServer() {
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":9000", nil))
}
