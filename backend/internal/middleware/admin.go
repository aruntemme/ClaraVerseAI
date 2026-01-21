package middleware

import (
	"claraverse/internal/config"
	"log"

	"github.com/gofiber/fiber/v2"
)

// AdminMiddleware checks if the authenticated user is a superadmin
func AdminMiddleware(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}

		// Check role from context (set by JWT middleware)
		role, hasRole := c.Locals("user_role").(string)

		// First check role field (preferred method)
		if hasRole && role == "admin" {
			c.Locals("is_superadmin", true)
			log.Printf("✅ Admin access granted to user %s (role: %s)", userID, role)
			return c.Next()
		}

		// Fallback: Check if user is in superadmin list (legacy support)
		isSuperadmin := false
		for _, adminID := range cfg.SuperadminUserIDs {
			if adminID == userID {
				isSuperadmin = true
				break
			}
		}

		if !isSuperadmin {
			log.Printf("🚫 Non-admin user %s attempted to access admin endpoint (role: %s)", userID, role)
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Admin access required",
			})
		}

		// Store admin flag for handlers to use
		c.Locals("is_superadmin", true)
		return c.Next()
	}
}

// IsSuperadmin is a helper function to check if a user ID is a superadmin
func IsSuperadmin(userID string, cfg *config.Config) bool {
	for _, adminID := range cfg.SuperadminUserIDs {
		if adminID == userID {
			return true
		}
	}
	return false
}
