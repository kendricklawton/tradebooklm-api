package middleware

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/workos/workos-go/v6/pkg/usermanagement"
)

func mustGetenv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("Fatal Error in connect_connector.go: %s environment variable not set.\n", k)
	}
	return v
}

func CORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins: []string{
			os.Getenv("TRADEBOOKLM_WEB_URL"),
		},
		AllowMethods:     []string{"DELETE", "GET", "PATCH", "POST"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

var (
	jwks           *keyfunc.JWKS
	workosClientID string
)

func InitAuth() {

	var (
		workosAPIKey   = mustGetenv("WORKOS_API_KEY")
		workosClientID = mustGetenv("WORKOS_CLIENT_ID")
	)

	usermanagement.SetAPIKey(workosAPIKey)

	jwksURL, err := usermanagement.GetJWKSURL(workosClientID)
	if err != nil {
		log.Fatalf("Failed to get JWKS URL from WorkOS: %v", err)
	}

	jwks, err = keyfunc.Get(jwksURL.String(), keyfunc.Options{})
	if err != nil {
		log.Fatalf("Failed to create JWKS from URL (%s): %v", jwksURL, err)
	}
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			workosClientID = mustGetenv("WORKOS_CLIENT_ID")
		)

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token format required"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, jwks.Keyfunc)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
			c.Abort()
			return
		}

		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		issuer, _ := claims.GetIssuer()
		if !strings.EqualFold(issuer, "https://api.workos.com/user_management/"+workosClientID) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token issuer"})
			c.Abort()
			return
		}

		workosId, err := claims.GetSubject()
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token subject (user ID)"})
			c.Abort()
			return
		}

		c.Set("workos_id", workosId)
		c.Next()
	}
}

func WebhookMiddleware() gin.HandlerFunc {
	var (
		internalSecret = mustGetenv("INTERNAL_API_SECRET")
	)

	return func(c *gin.Context) {
		providedKey := c.GetHeader("X-Internal-Api-Key")
		if providedKey != internalSecret {
			log.Printf("WARN: Unauthorized internal API attempt from %s", c.ClientIP())
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		c.Next()
	}
}
