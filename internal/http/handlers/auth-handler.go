package handler

import (
	"api/config"
	"api/internal/storage/postgres"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// this should be called every time the session is validated with whop, return new tokens
func HandleAuthAccountWrapper(c *fiber.Ctx, pgManager postgres.PGManager) error {
	license := c.Query("license")
	if license == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Required query parameter \"license\" is not present."})
	}
	accountId, err := pgManager.GetAccountID(license)
	if err != nil {
		log.Print(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "unable to authorise account"})
	}
	signedAccessToken, signedRefreshToken, refreshExp, err := generateTokenPair(license, accountId)
	if err != nil {
		log.Print(err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	userData := &postgres.UserData{
		AccountID:          accountId,
		RefreshToken:       signedRefreshToken,
		RefreshTokenExpiry: refreshExp,
	}
	err = pgManager.UpsertUserData(userData)
	if err != nil {
		log.Print(err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"accessToken": signedAccessToken, "refreshToken": signedRefreshToken})
}

func HandlerRefreshTokenWrapper(c *fiber.Ctx, pgManager postgres.PGManager) error {
	type tokenReqBody struct {
		RefreshToken string `json:"refreshToken"`
	}
	tokenReq := &tokenReqBody{}
	if err := c.BodyParser(tokenReq); err != nil {
		log.Print(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid data in request body"})
	}
	refreshSigningKey, err := config.Config("JWT_REFRESH_SIGNING_KEY")
	if err != nil {
		log.Print(err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	tokenString := tokenReq.RefreshToken

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(refreshSigningKey), nil
	})
	if err != nil {
		log.Print(err)
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		expirationTime := int64(claims["exp"].(float64))
		currentTime := time.Now().Unix()
		if currentTime > expirationTime {
			//need to get new refreshtoken
			log.Print("attempted to get new auth with expired refresh token")
			return c.SendStatus(fiber.StatusUnauthorized)
		}
		subInt, err := strconv.Atoi(claims["sub"].(string))
		if err != nil {
			log.Print(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		/*CHECK WHETHER TOKEN MATCHES user_data TOKEN*/
		currentUserData, err := pgManager.GetUserDataByAccountID(subInt)
		if err != nil {
			log.Print(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		if currentUserData.RefreshToken != tokenString {
			log.Printf("user %d attempted to get new auth with outdated refresh token", currentUserData.AccountID)
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		account, err := pgManager.GetAccountByID(subInt)
		if err != nil {
			log.Print(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		signedAccessToken, signedRefreshToken, refreshExp, err := generateTokenPair(account.License, subInt)
		if err != nil {
			log.Print(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		userData := &postgres.UserData{
			AccountID:          subInt,
			RefreshToken:       signedRefreshToken,
			RefreshTokenExpiry: refreshExp,
		}
		err = pgManager.UpsertUserData(userData)
		if err != nil {
			log.Print(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{"accessToken": signedAccessToken, "refreshToken": signedRefreshToken})
	}
	return c.SendStatus(fiber.StatusUnauthorized)
}
