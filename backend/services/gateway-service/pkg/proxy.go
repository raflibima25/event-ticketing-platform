package pkg

import (
	"context"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/idtoken"
)

// ProxyHandler creates a reverse proxy handler for backend services
func ProxyHandler(targetURL string) gin.HandlerFunc {
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	return func(c *gin.Context) {
		// Build target URL
		target := targetURL + c.Request.URL.Path
		if c.Request.URL.RawQuery != "" {
			target += "?" + c.Request.URL.RawQuery
		}

		// Create new request
		proxyReq, err := http.NewRequest(c.Request.Method, target, c.Request.Body)
		if err != nil {
			log.Printf("[Proxy Error] Failed to create request: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create proxy request",
			})
			return
		}

		// Copy headers from original request
		for key, values := range c.Request.Header {
			// Skip host header as it will be set by http.Client
			if strings.ToLower(key) == "host" {
				continue
			}
			for _, value := range values {
				proxyReq.Header.Add(key, value)
			}
		}

		// Add user context headers from JWT middleware
		if userID, exists := c.Get("user_id"); exists {
			proxyReq.Header.Set("X-User-ID", userID.(string))
		}
		if email, exists := c.Get("email"); exists {
			proxyReq.Header.Set("X-User-Email", email.(string))
		}
		if role, exists := c.Get("role"); exists {
			proxyReq.Header.Set("X-User-Role", role.(string))
		}

		// Add correlation ID
		if correlationID, exists := c.Get("correlation_id"); exists {
			proxyReq.Header.Set("X-Correlation-ID", correlationID.(string))
		} else if correlationID := c.GetHeader("X-Request-ID"); correlationID != "" {
			proxyReq.Header.Set("X-Correlation-ID", correlationID)
		}

		// Add identity token for Cloud Run service-to-service authentication
		// This allows the gateway to call private Cloud Run services
		if strings.Contains(targetURL, "run.app") {
			tokenSource, err := idtoken.NewTokenSource(context.Background(), targetURL)
			if err != nil {
				log.Printf("[Proxy Warning] Failed to create token source for %s: %v", targetURL, err)
			} else {
				token, err := tokenSource.Token()
				if err != nil {
					log.Printf("[Proxy Warning] Failed to get identity token for %s: %v", targetURL, err)
				} else {
					proxyReq.Header.Set("Authorization", "Bearer "+token.AccessToken)
				}
			}
		}

		// Execute request
		resp, err := client.Do(proxyReq)
		if err != nil {
			log.Printf("[Proxy Error] Request failed: %v", err)
			c.JSON(http.StatusBadGateway, gin.H{
				"error":   "Backend service unavailable",
				"service": targetURL,
			})
			return
		}
		defer resp.Body.Close()

		// Copy response headers
		for key, values := range resp.Header {
			for _, value := range values {
				c.Writer.Header().Add(key, value)
			}
		}

		// Set status code
		c.Status(resp.StatusCode)

		// Copy response body
		if _, err := io.Copy(c.Writer, resp.Body); err != nil {
			log.Printf("[Proxy Error] Failed to copy response body: %v", err)
		}
	}
}
