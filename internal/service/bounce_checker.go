package service

import (
	"context"

	"github.com/vhvcorp/go-notification-service/internal/repository"
)

// BounceChecker checks if emails have bounced
type BounceChecker struct {
	repo *repository.BounceRepository
}

// NewBounceChecker creates a new bounce checker
func NewBounceChecker(repo *repository.BounceRepository) *BounceChecker {
	return &BounceChecker{repo: repo}
}

// IsEmailBounced checks if an email has hard bounced recently
func (bc *BounceChecker) IsEmailBounced(ctx context.Context, email string) (bool, error) {
	// Check for hard bounces in the last 30 days
	bounces, err := bc.repo.FindRecentHardBounces(ctx, email, 30)
	if err != nil {
		return false, err
	}

	return len(bounces) > 0, nil
}
