package middleware

import (
	"api/config"
	"log"
	"strings"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// var jwtFilter = []string{"/account/get"}

func jwtProtectPaths() fiber.Handler {
	signingKey, err := config.Config("JWT_SIGNING_KEY")
	if err != nil {
		log.Fatal(err)
	}

	jwtMiddleware := jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(signingKey)},
	})
	return func(c *fiber.Ctx) error {
		// Check if the path starts with "/api/account"
		if strings.HasPrefix(c.Path(), "/api/account") {
			return jwtMiddleware(c)
		}
		if strings.HasPrefix(c.Path(), "/api/user") {
			return jwtMiddleware(c)
		}
		return c.Next()
	}
}

func UseMiddlewares(app *fiber.App) error {

	app.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	}))
	app.Use(keyauth.New(keyauth.Config{
		KeyLookup: "header:api_key",
		Validator: validateAPIKey,
		Next:      authFilter,
	}))
	app.Use(jwtProtectPaths())
	return nil
}
