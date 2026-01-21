package jobs

import (
	"claraverse/internal/database"
	"claraverse/internal/models"
	"claraverse/internal/services"
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// PromoExpirationChecker handles expiration of promotional pro subscriptions
type PromoExpirationChecker struct {
	mongoDB     *database.MongoDB
	userService *services.UserService
	tierService *services.TierService
}

// NewPromoExpirationChecker creates a new promo expiration checker
func NewPromoExpirationChecker(
	mongoDB *database.MongoDB,
	userService *services.UserService,
	tierService *services.TierService,
) *PromoExpirationChecker {
	return &PromoExpirationChecker{
		mongoDB:     mongoDB,
		userService: userService,
		tierService: tierService,
	}
}

// Run checks for expired promotional subscriptions and downgrades users
func (p *PromoExpirationChecker) Run(ctx context.Context) error {
	if p.mongoDB == nil || p.userService == nil || p.tierService == nil {
		log.Println("⚠️  [PROMO-EXPIRATION] Promo expiration checker disabled (requires MongoDB, UserService, TierService)")
		return nil
	}

	log.Println("⏰ [PROMO-EXPIRATION] Checking for expired promotional subscriptions...")
	startTime := time.Now()

	// Find users collection
	collection := p.mongoDB.Database().Collection("users")

	// Find promo users with expired subscriptions
	// Criteria:
	// - subscriptionTier = "pro"
	// - subscriptionExpiresAt < now
	// - dodoSubscriptionId is empty (not a paid subscriber)
	filter := bson.M{
		"subscriptionTier": models.TierPro,
		"subscriptionExpiresAt": bson.M{
			"$lt": time.Now().UTC(),
		},
		"$or": []bson.M{
			{"dodoSubscriptionId": ""},
			{"dodoSubscriptionId": bson.M{"$exists": false}},
		},
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		log.Printf("❌ [PROMO-EXPIRATION] Failed to query users: %v", err)
		return err
	}
	defer cursor.Close(ctx)

	expiredCount := 0
	for cursor.Next(ctx) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			log.Printf("⚠️  [PROMO-EXPIRATION] Failed to decode user: %v", err)
			continue
		}

		if err := p.expirePromoSubscription(ctx, &user); err != nil {
			log.Printf("⚠️  [PROMO-EXPIRATION] Failed to expire promo for user %s: %v", user.SupabaseUserID, err)
			continue
		}

		expiredCount++
		log.Printf("✅ [PROMO-EXPIRATION] Expired promo subscription for user %s (promo ended %v ago)",
			user.SupabaseUserID, time.Since(*user.SubscriptionExpiresAt).Round(time.Hour))
	}

	duration := time.Since(startTime)
	log.Printf("✅ [PROMO-EXPIRATION] Check complete: expired %d promotional subscriptions in %v", expiredCount, duration)

	return nil
}

// expirePromoSubscription downgrades a user from promo pro to free tier
func (p *PromoExpirationChecker) expirePromoSubscription(ctx context.Context, user *models.User) error {
	// Update user to free tier with cancelled status
	collection := p.mongoDB.Database().Collection("users")

	update := bson.M{
		"$set": bson.M{
			"subscriptionTier":   models.TierFree,
			"subscriptionStatus": models.SubStatusCancelled,
		},
	}

	_, err := collection.UpdateOne(ctx, bson.M{"_id": user.ID}, update)
	if err != nil {
		return err
	}

	// Invalidate tier cache so user immediately sees free tier on next request
	if p.tierService != nil {
		p.tierService.InvalidateCache(user.SupabaseUserID)
	}

	return nil
}

// GetNextRunTime returns when the job should run next (hourly)
func (p *PromoExpirationChecker) GetNextRunTime() time.Time {
	return time.Now().UTC().Add(1 * time.Hour)
}
