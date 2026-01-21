package middleware

import (
	"claraverse/internal/services"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// ExecutionLimiter middleware checks daily execution limits based on user tier
type ExecutionLimiter struct {
	tierService *services.TierService
	redis       *redis.Client
}

// NewExecutionLimiter creates a new execution limiter middleware
func NewExecutionLimiter(tierService *services.TierService, redisClient *redis.Client) *ExecutionLimiter {
	return &ExecutionLimiter{
		tierService: tierService,
		redis:       redisClient,
	}
}

// CheckLimit verifies if user can execute another workflow today
func (el *ExecutionLimiter) CheckLimit(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	ctx := context.Background()

	// Get user's tier limits
	limits := el.tierService.GetLimits(ctx, userIDStr)

	// If unlimited executions, skip check
	if limits.MaxExecutionsPerDay == -1 {
		return c.Next()
	}

	// Get today's execution count from Redis
	today := time.Now().UTC().Format("2006-01-02")
	key := fmt.Sprintf("executions:%s:%s", userIDStr, today)

	// Get current count
	count, err := el.redis.Get(ctx, key).Int64()
	if err != nil && err != redis.Nil {
		log.Printf("⚠️  Failed to get execution count from Redis: %v", err)
		// On Redis error, allow execution but log warning
		return c.Next()
	}

	// Check if limit exceeded
	if count >= limits.MaxExecutionsPerDay {
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"error":     "Daily execution limit exceeded",
			"limit":     limits.MaxExecutionsPerDay,
			"used":      count,
			"reset_at":  getNextMidnightUTC(),
		})
	}

	// Store current count in context for post-execution increment
	c.Locals("execution_count_key", key)

	return c.Next()
}

// IncrementCount increments the execution counter after successful execution start
func (el *ExecutionLimiter) IncrementCount(userID string) error {
	if el.redis == nil {
		return nil // Redis not available, skip increment
	}

	ctx := context.Background()
	today := time.Now().UTC().Format("2006-01-02")
	key := fmt.Sprintf("executions:%s:%s", userID, today)

	// Increment counter
	pipe := el.redis.Pipeline()
	pipe.Incr(ctx, key)

	// Set expiry to end of day + 1 day (to allow historical querying)
	midnight := getNextMidnightUTC()
	expiryDuration := time.Until(midnight) + 24*time.Hour
	pipe.Expire(ctx, key, expiryDuration)

	_, err := pipe.Exec(ctx)
	if err != nil {
		log.Printf("⚠️  Failed to increment execution count: %v", err)
		return err
	}

	log.Printf("✅ Incremented execution count for user %s (key: %s)", userID, key)
	return nil
}

// GetRemainingExecutions returns how many executions user has left today
func (el *ExecutionLimiter) GetRemainingExecutions(userID string) (int64, error) {
	if el.redis == nil {
		return -1, nil // Redis not available, return unlimited
	}

	ctx := context.Background()

	// Get user's tier limits
	limits := el.tierService.GetLimits(ctx, userID)
	if limits.MaxExecutionsPerDay == -1 {
		return -1, nil // Unlimited
	}

	// Get today's count
	today := time.Now().UTC().Format("2006-01-02")
	key := fmt.Sprintf("executions:%s:%s", userID, today)

	count, err := el.redis.Get(ctx, key).Int64()
	if err == redis.Nil {
		return limits.MaxExecutionsPerDay, nil // No executions today
	}
	if err != nil {
		return -1, err
	}

	remaining := limits.MaxExecutionsPerDay - count
	if remaining < 0 {
		return 0, nil
	}

	return remaining, nil
}

// getNextMidnightUTC returns the next midnight UTC
func getNextMidnightUTC() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
}
