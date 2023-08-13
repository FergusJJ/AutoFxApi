package handler

import (
	"api/config"
	"api/internal/storage/postgres"
	"api/pkg/whop"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func HandleCreateAccountWrapper(c *fiber.Ctx, pg postgres.PGAccount) error {
	log.Print(c.Locals("user"))
	payload := &CreateAccountRequest{}
	if err := c.BodyParser(payload); err != nil {
		payload := invalidRequestResponse{
			ResponseCode: fiber.StatusBadRequest,
			Message:      "invalid data in request body",
		}
		log.Print(err)
		return c.Status(fiber.StatusBadRequest).JSON(payload)
	}
	_, _, err := whop.CheckLicenseKey(payload.LicenseKey)
	if err != nil {
		log.Print(err)
		return c.SendStatus(fiber.StatusNotFound)
	}
	account := newAccount(payload.LicenseKey, payload.Email)
	err = pg.CreateAccount(account)
	if err != nil {
		payload := invalidRequestResponse{
			ResponseCode: fiber.StatusInternalServerError,
			Message:      "unable to create account",
		}
		log.Print(err)
		return c.Status(fiber.StatusInternalServerError).JSON(payload)
	}
	resPayload := AccountCreatedResponse{ID: account.ID, LicenseKey: account.License}
	return c.Status(fiber.StatusCreated).JSON(resPayload)
}

func HandleDeleteAccountWrapper(c *fiber.Ctx, pg postgres.PGAccount) error {
	payload := &AccountRequest{}
	if err := c.BodyParser(payload); err != nil {
		payload := invalidRequestResponse{
			ResponseCode: fiber.StatusBadRequest,
			Message:      "invalid data in request body",
		}
		log.Print(err)
		return c.Status(fiber.StatusBadRequest).JSON(payload)
	}
	_, _, err := whop.CheckLicenseKey(payload.LicenseKey)
	if err != nil {
		log.Print(err)
		return c.SendStatus(fiber.StatusNotFound)
	}
	ok, err := pg.CheckRelation(payload.ID, payload.LicenseKey)
	if err != nil {
		log.Print(err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if !ok {
		return c.SendStatus(fiber.StatusForbidden)
	}
	err = pg.DeleteAccount(payload.ID)
	if err != nil {
		payload := invalidRequestResponse{
			ResponseCode: fiber.StatusInternalServerError,
			Message:      "unable to delete account",
		}
		log.Print(err)
		return c.Status(fiber.StatusInternalServerError).JSON(payload)
	}
	return c.SendStatus(fiber.StatusOK)
}

func HandleGetAccountWrapper(c *fiber.Ctx, pg postgres.PGAccount) error {
	payload := &AccountRequest{}
	if err := c.BodyParser(payload); err != nil {
		payload := invalidRequestResponse{
			ResponseCode: fiber.StatusBadRequest,
			Message:      "invalid data in request body",
		}
		log.Print(err)
		return c.Status(fiber.StatusBadRequest).JSON(payload)
	}
	_, _, err := whop.CheckLicenseKey(payload.LicenseKey)
	if err != nil {
		log.Print(err)
		return c.SendStatus(fiber.StatusNotFound)
	}
	ok, err := pg.CheckRelation(payload.ID, payload.LicenseKey)
	if err != nil {
		log.Print(err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if !ok {
		return c.SendStatus(fiber.StatusForbidden)
	}
	account, err := pg.GetAccountByID(payload.ID)
	if err != nil {
		payload := invalidRequestResponse{
			ResponseCode: fiber.StatusInternalServerError,
			Message:      "unable to get account",
		}
		log.Print(err)
		return c.Status(fiber.StatusInternalServerError).JSON(payload)
	}
	return c.Status(fiber.StatusOK).JSON(*account)
}

func HandleUpdateAccountWrapper(c *fiber.Ctx, pg postgres.PGAccount) error {
	payload := &UpdateAccountRequest{}
	if err := c.BodyParser(payload); err != nil {
		payload := invalidRequestResponse{
			ResponseCode: fiber.StatusBadRequest,
			Message:      "invalid data in request body",
		}
		log.Print(err)
		return c.Status(fiber.StatusBadRequest).JSON(payload)
	}
	_, _, err := whop.CheckLicenseKey(payload.LicenseKey)
	if err != nil {
		log.Print(err)
		return c.SendStatus(fiber.StatusNotFound)
	}
	//might have to rethink this, if update account is called after licensekey change, the relation wont match the new license key, so will have to pass in old one as well?
	ok, err := pg.CheckRelation(payload.ID, payload.LicenseKey)
	if err != nil {
		log.Print(err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if !ok {
		return c.SendStatus(fiber.StatusForbidden)
	}
	var account *postgres.Account = &postgres.Account{
		Email:   payload.Email,
		ID:      payload.ID,
		License: payload.LicenseKey,
	}
	err = pg.UpdateAccount(account)
	if err != nil {
		payload := invalidRequestResponse{
			ResponseCode: fiber.StatusInternalServerError,
			Message:      "unable to update account",
		}
		log.Print(err)
		return c.Status(fiber.StatusInternalServerError).JSON(payload)
	}
	return c.SendStatus(fiber.StatusCreated)
}

// this should be called every time the session is validated with whop, return new tokens
func HandleAuthAccountWrapper(c *fiber.Ctx, pgManager postgres.PGManager) error {
	license := c.Query("license")
	if license == "" {
		payload := invalidRequestResponse{
			ResponseCode: 400,
			Message:      "Required query parameter \"license\" is not present.",
		}
		return c.Status(fiber.StatusBadRequest).JSON(payload)
	}
	accountId, err := pgManager.GetAccountID(license)
	if err != nil {
		payload := invalidRequestResponse{
			ResponseCode: fiber.StatusInternalServerError,
			Message:      "unable to authorise account",
		}
		log.Print(err)
		return c.Status(fiber.StatusInternalServerError).JSON(payload)
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
		payload := invalidRequestResponse{
			ResponseCode: 401,
			Message:      "invalid data in request body",
		}
		return c.Status(fiber.StatusBadRequest).JSON(payload)
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
