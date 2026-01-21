package middleware

import (
	"claraverse/pkg/auth"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware verifies Supabase JWT tokens
// Supports both Authorization header and query parameter (for WebSocket connections)
func AuthMiddleware(supabaseAuth *auth.SupabaseAuth) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// SECURITY: DEV_API_KEY bypass has been removed for security reasons.
		// Use proper Supabase authentication or separate development/staging environments.

		// Skip auth if Supabase is not configured (development mode ONLY)
		if supabaseAuth.URL == "" {
			environment := os.Getenv("ENVIRONMENT")

			// CRITICAL: Never allow auth bypass in production
			if environment == "production" {
				log.Fatal("❌ CRITICAL SECURITY ERROR: Supabase not configured in production environment. Authentication is required.")
			}

			// Only allow bypass in development/testing
			if environment != "development" && environment != "testing" && environment != "" {
				return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
					"error": "Authentication service unavailable",
				})
			}

			log.Println("⚠️  Auth skipped: Supabase not configured (development mode)")
			c.Locals("user_id", "dev-user")
			c.Locals("user_email", "dev@localhost")
			c.Locals("user_role", "authenticated")
			return c.Next()
		}

		// Try to extract token from multiple sources
		var token string

		// 1. Try Authorization header first
		authHeader := c.Get("Authorization")
		if authHeader != "" {
			extractedToken, err := auth.ExtractToken(authHeader)
			if err == nil {
				token = extractedToken
			}
		}

		// 2. Try query parameter (for WebSocket connections)
		if token == "" {
			token = c.Query("token")
		}

		// No token found
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing or invalid authorization token",
			})
		}

		// Verify token with Supabase
		user, err := supabaseAuth.VerifyToken(token)
		if err != nil {
			log.Printf("❌ Auth failed: %v", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		// Store user info in context
		c.Locals("user_id", user.ID)
		c.Locals("user_email", user.Email)
		c.Locals("user_role", user.Role)

		log.Printf("✅ Authenticated user: %s (%s)", user.Email, user.ID)
		return c.Next()
	}
}

// OptionalAuthMiddleware makes authentication optional
// Supports both Authorization header and query parameter (for WebSocket)
func OptionalAuthMiddleware(supabaseAuth *auth.SupabaseAuth) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Try to extract token from multiple sources
		var token string

		// 1. Try Authorization header first
		authHeader := c.Get("Authorization")
		if authHeader != "" {
			extractedToken, err := auth.ExtractToken(authHeader)
			if err == nil {
				token = extractedToken
			}
		}

		// 2. Try query parameter (for WebSocket connections)
		if token == "" {
			token = c.Query("token")
		}

		// If no token found, proceed as anonymous
		if token == "" {
			c.Locals("user_id", "anonymous")
			log.Println("🔓 Anonymous connection")
			return c.Next()
		}

		// Skip validation if Supabase is not configured (development mode ONLY)
		if supabaseAuth == nil || supabaseAuth.URL == "" {
			environment := os.Getenv("ENVIRONMENT")

			// CRITICAL: Never allow auth bypass in production
			if environment == "production" {
				log.Fatal("❌ CRITICAL SECURITY ERROR: Supabase not configured in production environment. Authentication is required.")
			}

			// Only allow in development/testing
			if environment != "development" && environment != "testing" && environment != "" {
				c.Locals("user_id", "anonymous")
				log.Println("⚠️  Supabase unavailable, proceeding as anonymous")
				return c.Next()
			}

			c.Locals("user_id", "dev-user-" + token[:min(8, len(token))])
			c.Locals("user_email", "dev@localhost")
			c.Locals("user_role", "authenticated")
			log.Println("⚠️  Auth skipped: Supabase not configured (dev mode)")
			return c.Next()
		}

		// Verify token with Supabase
		user, err := supabaseAuth.VerifyToken(token)
		if err != nil {
			log.Printf("⚠️  Token validation failed: %v (continuing as anonymous)", err)
			c.Locals("user_id", "anonymous")
			return c.Next()
		}

		// Store authenticated user info
		c.Locals("user_id", user.ID)
		c.Locals("user_email", user.Email)
		c.Locals("user_role", user.Role)

		log.Printf("✅ Authenticated user: %s (%s)", user.Email, user.ID)
		return c.Next()
	}
}
