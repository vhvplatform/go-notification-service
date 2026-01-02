package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"github.com/vhvplatform/go-notification-service/internal/domain"
	"github.com/vhvplatform/go-notification-service/internal/repository"
	"github.com/vhvplatform/go-notification-service/internal/scheduler"
	"github.com/vhvplatform/go-notification-service/internal/shared/errors"
	"github.com/vhvplatform/go-notification-service/internal/shared/logger"
)

// ScheduleHandler handles scheduled notification requests
type ScheduleHandler struct {
	repo      *repository.ScheduledNotificationRepository
	scheduler *scheduler.NotificationScheduler
	log       *logger.Logger
}

// NewScheduleHandler creates a new schedule handler
func NewScheduleHandler(repo *repository.ScheduledNotificationRepository, scheduler *scheduler.NotificationScheduler, log *logger.Logger) *ScheduleHandler {
	return &ScheduleHandler{
		repo:      repo,
		scheduler: scheduler,
		log:       log,
	}
}

// GetSchedules godoc
// @Summary Get scheduled notifications
// @Description Get list of scheduled notifications
// @Tags scheduled
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Success 200 {object} map[string]interface{} "List of schedules"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/scheduled [get]
func (h *ScheduleHandler) GetSchedules(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, errors.NewValidationError("tenant_id is required", nil))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	schedules, total, err := h.repo.FindByTenantID(c.Request.Context(), tenantID, page, pageSize)
	if err != nil {
		h.log.Error("Failed to get schedules", "error", err)
		c.JSON(http.StatusInternalServerError, errors.NewInternalError("Failed to get schedules", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      schedules,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// CreateSchedule godoc
// @Summary Create scheduled notification
// @Description Create a new scheduled notification
// @Tags scheduled
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param schedule body object true "Schedule request"
// @Success 201 {object} map[string]interface{} "Schedule created"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/scheduled [post]
func (h *ScheduleHandler) CreateSchedule(c *gin.Context) {
	var sched domain.ScheduledNotification
	if err := c.ShouldBindJSON(&sched); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewValidationError("Invalid request", err))
		return
	}

	// Validate cron expression
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(sched.Schedule)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewValidationError("Invalid cron expression", err))
		return
	}

	// Set next run time
	sched.NextRunAt = schedule.Next(time.Now())
	sched.IsActive = true

	// Add schedule
	if err := h.scheduler.AddSchedule(&sched); err != nil {
		h.log.Error("Failed to create schedule", "error", err)
		c.JSON(http.StatusInternalServerError, errors.NewInternalError("Failed to create schedule", err))
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Schedule created successfully",
		"data":    sched,
	})
}

// UpdateSchedule godoc
// @Summary Update scheduled notification
// @Description Update an existing scheduled notification
// @Tags scheduled
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param id path string true "Schedule ID"
// @Param schedule body object true "Schedule update"
// @Success 200 {object} map[string]interface{} "Schedule updated"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "Schedule not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/scheduled/{id} [put]
func (h *ScheduleHandler) UpdateSchedule(c *gin.Context) {
	id := c.Param("id")

	var sched domain.ScheduledNotification
	if err := c.ShouldBindJSON(&sched); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewValidationError("Invalid request", err))
		return
	}

	// Validate cron expression if changed
	if sched.Schedule != "" {
		parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		schedule, err := parser.Parse(sched.Schedule)
		if err != nil {
			c.JSON(http.StatusBadRequest, errors.NewValidationError("Invalid cron expression", err))
			return
		}
		sched.NextRunAt = schedule.Next(time.Now())
	}

	// Get existing schedule
	existing, err := h.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		h.log.Error("Failed to find schedule", "error", err)
		c.JSON(http.StatusNotFound, errors.NewNotFoundError("Schedule not found", err))
		return
	}

	// Update fields
	existing.Schedule = sched.Schedule
	existing.Request = sched.Request
	existing.IsActive = sched.IsActive
	existing.NextRunAt = sched.NextRunAt

	if err := h.repo.Update(c.Request.Context(), existing); err != nil {
		h.log.Error("Failed to update schedule", "error", err)
		c.JSON(http.StatusInternalServerError, errors.NewInternalError("Failed to update schedule", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Schedule updated successfully",
		"data":    existing,
	})
}

// DeleteSchedule godoc
// @Summary Delete scheduled notification
// @Description Delete a scheduled notification
// @Tags scheduled
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param id path string true "Schedule ID"
// @Success 200 {object} map[string]interface{} "Schedule deleted"
// @Failure 404 {object} map[string]interface{} "Schedule not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/scheduled/{id} [delete]
func (h *ScheduleHandler) DeleteSchedule(c *gin.Context) {
	id := c.Param("id")

	if err := h.scheduler.RemoveSchedule(id); err != nil {
		h.log.Error("Failed to delete schedule", "error", err)
		c.JSON(http.StatusInternalServerError, errors.NewInternalError("Failed to delete schedule", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Schedule deleted successfully",
	})
}
