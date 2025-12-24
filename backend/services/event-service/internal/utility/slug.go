package utility

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// GenerateSlug generates URL-friendly slug from title
func GenerateSlug(title string) string {
	// Convert to lowercase
	slug := strings.ToLower(title)

	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove special characters
	reg := regexp.MustCompile("[^a-z0-9-]+")
	slug = reg.ReplaceAllString(slug, "")

	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile("-+")
	slug = reg.ReplaceAllString(slug, "-")

	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")

	// Add unique suffix to ensure uniqueness
	suffix := time.Now().Format("20060102") + "-" + uuid.New().String()[:8]
	slug = slug + "-" + suffix

	return slug
}
