package handler

import (
	"api/internal/storage/postgres"
	"api/pkg/whop"
	"log"

	"github.com/gofiber/fiber/v2"
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
