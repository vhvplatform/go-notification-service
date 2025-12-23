package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/longvhv/saas-framework-go/pkg/config"
	"github.com/longvhv/saas-framework-go/pkg/logger"
	"github.com/longvhv/saas-framework-go/pkg/mongodb"
	"github.com/longvhv/saas-framework-go/pkg/rabbitmq"
	"github.com/longvhv/saas-framework-go/services/notification-service/internal/consumer"
	"github.com/longvhv/saas-framework-go/services/notification-service/internal/handler"
	"github.com/longvhv/saas-framework-go/services/notification-service/internal/repository"
	"github.com/longvhv/saas-framework-go/services/notification-service/internal/service"
)

func main() {
	// Initialize logger
	log := logger.NewLogger()
	defer log.Sync()

	log.Info("Starting Notification Service...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration", "error", err)
	}

	// Initialize MongoDB
	mongoClient, err := mongodb.NewMongoClient(cfg.MongoDB.URI, cfg.MongoDB.Database)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB", "error", err)
	}
	defer mongoClient.Disconnect(context.Background())

	// Initialize RabbitMQ
	rabbitMQClient, err := rabbitmq.NewRabbitMQClient(cfg.RabbitMQ.URL)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ", "error", err)
	}
	defer rabbitMQClient.Close()

	// Initialize repositories
	notificationRepo := repository.NewNotificationRepository(mongoClient)
	templateRepo := repository.NewTemplateRepository(mongoClient)

	// Initialize services
	emailConfig := service.EmailConfig{
		SMTPHost:     cfg.SMTP.Host,
		SMTPPort:     cfg.SMTP.Port,
		SMTPUsername: cfg.SMTP.Username,
		SMTPPassword: cfg.SMTP.Password,
		FromEmail:    cfg.SMTP.FromEmail,
		FromName:     cfg.SMTP.FromName,
	}
	emailService := service.NewEmailService(emailConfig, notificationRepo, templateRepo, log)
	webhookService := service.NewWebhookService(notificationRepo, log)
	notificationService := service.NewNotificationService(notificationRepo, emailService, webhookService, log)

	// Initialize HTTP handlers
	notificationHandler := handler.NewNotificationHandler(notificationService, log)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Health check endpoints
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})
	router.GET("/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		notifications := v1.Group("/notifications")
		{
			notifications.POST("/email", notificationHandler.SendEmail)
			notifications.POST("/webhook", notificationHandler.SendWebhook)
			notifications.GET("", notificationHandler.GetNotifications)
			notifications.GET("/:id", notificationHandler.GetNotification)
		}
	}

	// Start RabbitMQ consumer
	eventConsumer := consumer.NewEventConsumer(rabbitMQClient, notificationService, log)
	go func() {
		if err := eventConsumer.Start(); err != nil {
			log.Error("Failed to start event consumer", "error", err)
		}
	}()

	// Start HTTP server
	port := os.Getenv("NOTIFICATION_SERVICE_PORT")
	if port == "" {
		port = "8084"
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Info("Notification Service started", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", "error", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down Notification Service...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", "error", err)
	}

	log.Info("Notification Service stopped")
}
