package middleware

import (
	"claraverse/pkg/auth"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
)

// LocalAuthMiddleware verifies local JWT tokens
// Supports both Authorization header and query parameter (for WebSocket connections)
func LocalAuthMiddleware(jwtAuth *auth.LocalJWTAuth) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip auth if JWT secret is not configured (development mode ONLY)
		environment := os.Getenv("ENVIRONMENT")

		if jwtAuth == nil {
			// CRITICAL: Never allow auth bypass in production
			if environment == "production" {
				log.Fatal("❌ CRITICAL SECURITY ERROR: JWT auth not configured in production environment. Authentication is required.")
			}

			// Only allow bypass in development/testing
			if environment != "development" && environment != "testing" && environment != "" {
				return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
					"error": "Authentication service unavailable",
				})
			}

			log.Println("⚠️  Auth skipped: JWT not configured (development mode)")
			c.Locals("user_id", "dev-user")
			c.Locals("user_email", "dev@localhost")
			c.Locals("user_role", "user")
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

		// Verify JWT token
		user, err := jwtAuth.VerifyAccessToken(token)
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

// OptionalLocalAuthMiddleware makes authentication optional
// Supports both Authorization header and query parameter (for WebSocket)
func OptionalLocalAuthMiddleware(jwtAuth *auth.LocalJWTAuth) fiber.Handler {
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

		// Skip validation if JWT auth is not configured (development mode ONLY)
		environment := os.Getenv("ENVIRONMENT")

		if jwtAuth == nil {
			// CRITICAL: Never allow auth bypass in production
			if environment == "production" {
				log.Fatal("❌ CRITICAL SECURITY ERROR: JWT auth not configured in production environment")
			}

			// Only allow in development/testing
			if environment != "development" && environment != "testing" && environment != "" {
				c.Locals("user_id", "anonymous")
				log.Println("⚠️  JWT unavailable, proceeding as anonymous")
				return c.Next()
			}

			c.Locals("user_id", "dev-user")
			c.Locals("user_email", "dev@localhost")
			c.Locals("user_role", "user")
			log.Println("⚠️  Auth skipped: JWT not configured (dev mode)")
			return c.Next()
		}

		// Verify JWT token
		user, err := jwtAuth.VerifyAccessToken(token)
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

// RateLimitedAuthMiddleware combines rate limiting with authentication
// Rate limit: 5 attempts per 15 minutes per IP
// Note: This function is currently unused. Apply rate limiting separately in routes if needed.
