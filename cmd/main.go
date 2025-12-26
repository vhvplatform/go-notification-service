package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vhvplatform/go-notification-service/internal/consumer"
	"github.com/vhvplatform/go-notification-service/internal/dlq"
	"github.com/vhvplatform/go-notification-service/internal/handler"
	"github.com/vhvplatform/go-notification-service/internal/middleware"
	"github.com/vhvplatform/go-notification-service/internal/repository"
	"github.com/vhvplatform/go-notification-service/internal/scheduler"
	"github.com/vhvplatform/go-notification-service/internal/service"
	"github.com/vhvplatform/go-notification-service/internal/shared/config"
	"github.com/vhvplatform/go-notification-service/internal/shared/logger"
	"github.com/vhvplatform/go-notification-service/internal/shared/mongodb"
	"github.com/vhvplatform/go-notification-service/internal/shared/rabbitmq"
	"github.com/vhvplatform/go-notification-service/internal/webhook"
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
	failedNotificationRepo := repository.NewFailedNotificationRepository(mongoClient)
	scheduledNotificationRepo := repository.NewScheduledNotificationRepository(mongoClient)
	preferencesRepo := repository.NewPreferencesRepository(mongoClient)
	bounceRepo := repository.NewBounceRepository(mongoClient)

	// Get configuration from environment
	smtpPoolSize, _ := strconv.Atoi(getEnv("SMTP_POOL_SIZE", "10"))
	emailWorkers, _ := strconv.Atoi(getEnv("EMAIL_WORKERS", "5"))
	rateLimitPerTenant, _ := strconv.ParseFloat(getEnv("RATE_LIMIT_PER_TENANT", "100"), 64)
	rateLimitBurst, _ := strconv.Atoi(getEnv("RATE_LIMIT_BURST", "200"))

	// Initialize services
	emailConfig := service.EmailConfig{
		SMTPHost:     cfg.SMTP.Host,
		SMTPPort:     cfg.SMTP.Port,
		SMTPUsername: cfg.SMTP.Username,
		SMTPPassword: cfg.SMTP.Password,
		FromEmail:    cfg.SMTP.FromEmail,
		FromName:     cfg.SMTP.FromName,
		PoolSize:     smtpPoolSize,
	}
	emailService := service.NewEmailService(emailConfig, notificationRepo, templateRepo, log)
	defer emailService.Close()

	smsConfig := service.SMSConfig{
		Provider:    getEnv("SMS_PROVIDER", "twilio"),
		TwilioSID:   getEnv("TWILIO_SID", ""),
		TwilioToken: getEnv("TWILIO_TOKEN", ""),
		TwilioFrom:  getEnv("TWILIO_FROM", ""),
		AWSSNSARN:   getEnv("AWS_SNS_ARN", ""),
		AWSRegion:   getEnv("AWS_REGION", ""),
	}
	smsService := service.NewSMSService(smsConfig, notificationRepo, log)

	webhookService := service.NewWebhookService(notificationRepo, log)
	notificationService := service.NewNotificationService(notificationRepo, emailService, webhookService, smsService, log)

	// Initialize Dead Letter Queue
	deadLetterQueue := dlq.NewDeadLetterQueue(failedNotificationRepo, log)

	// Initialize Bounce Checker (can be integrated into email service if needed)
	_ = service.NewBounceChecker(bounceRepo)

	// Initialize Bulk Email Service
	bulkEmailService := service.NewBulkEmailService(emailService, emailWorkers, log)
	bulkEmailService.Start()
	defer bulkEmailService.Stop()

	// Initialize Scheduler
	notificationScheduler := scheduler.NewNotificationScheduler(notificationService, scheduledNotificationRepo, log)
	if err := notificationScheduler.Start(); err != nil {
		log.Error("Failed to start scheduler", "error", err)
	}
	defer notificationScheduler.Stop()

	// Initialize HTTP handlers
	notificationHandler := handler.NewNotificationHandler(notificationService, log)
	smsHandler := handler.NewSMSHandler(notificationService, log)
	bulkHandler := handler.NewBulkHandler(bulkEmailService, log)
	preferencesHandler := handler.NewPreferencesHandler(preferencesRepo, log)
	scheduleHandler := handler.NewScheduleHandler(scheduledNotificationRepo, notificationScheduler, log)
	dlqHandler := handler.NewDLQHandler(deadLetterQueue, notificationService, log)
	bounceHandler := webhook.NewBounceHandler(bounceRepo, log)

	// Initialize rate limiter
	rateLimiter := middleware.NewTenantRateLimiter(rateLimitPerTenant, rateLimitBurst)

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

	// Metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API routes with rate limiting
	v1 := router.Group("/api/v1")
	v1.Use(middleware.RateLimitMiddleware(rateLimiter))
	{
		// Notifications
		notifications := v1.Group("/notifications")
		{
			notifications.POST("/email", notificationHandler.SendEmail)
			notifications.POST("/webhook", notificationHandler.SendWebhook)
			notifications.POST("/sms", smsHandler.SendSMS)
			notifications.GET("", notificationHandler.GetNotifications)
			notifications.GET("/:id", notificationHandler.GetNotification)
		}

		// Bulk operations
		bulk := v1.Group("/notifications/bulk")
		{
			bulk.POST("/email", bulkHandler.SendBulkEmail)
		}

		// Preferences
		preferences := v1.Group("/preferences")
		{
			preferences.GET("/:user_id", preferencesHandler.GetPreferences)
			preferences.PUT("/:user_id", preferencesHandler.UpdatePreferences)
		}

		// Scheduled notifications
		scheduled := v1.Group("/scheduled")
		{
			scheduled.GET("", scheduleHandler.GetSchedules)
			scheduled.POST("", scheduleHandler.CreateSchedule)
			scheduled.PUT("/:id", scheduleHandler.UpdateSchedule)
			scheduled.DELETE("/:id", scheduleHandler.DeleteSchedule)
		}

		// Dead Letter Queue
		dlqRoutes := v1.Group("/dlq")
		{
			dlqRoutes.GET("", dlqHandler.GetFailedNotifications)
			dlqRoutes.POST("/:id/retry", dlqHandler.RetryNotification)
		}
	}

	// Webhooks (no rate limiting for external providers)
	webhooks := router.Group("/webhooks")
	{
		webhooks.POST("/ses", bounceHandler.HandleSESWebhook)
		webhooks.POST("/sendgrid", bounceHandler.HandleSendGridWebhook)
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

// getEnv retrieves environment variable or returns default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
