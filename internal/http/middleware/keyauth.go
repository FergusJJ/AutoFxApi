package middleware

import (
	"api/config"
	"crypto/sha256"
	"crypto/subtle"
	"log"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
)

var protectedURLs = []*regexp.Regexp{
	regexp.MustCompile("^/internal$"),
}

func validateAPIKey(c *fiber.Ctx, key string) (bool, error) {
	var apiKey, err = config.Config("MONITOR_SECRET_KEY")
	if err != nil {
		log.Fatal(err)
	}
	hashedAPIKey := sha256.Sum256([]byte(apiKey))
	hashedKey := sha256.Sum256([]byte(key))
	if subtle.ConstantTimeCompare(hashedAPIKey[:], hashedKey[:]) == 1 {
		return true, nil
	}
	return false, keyauth.ErrMissingOrMalformedAPIKey

}

func authFilter(c *fiber.Ctx) bool {
	originalURL := strings.ToLower(c.OriginalURL())
	for _, pattern := range protectedURLs {
		if pattern.MatchString(originalURL) {
			return false
		}
	}
	return true
}
