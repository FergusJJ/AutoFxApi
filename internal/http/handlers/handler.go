package handler

import (
	"database/sql"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	monitormanager "api/internal/monitor-manager"
	"api/internal/storage/postgres"
	cache "api/internal/storage/redis"
	"api/pkg/whop"

	"github.com/gofiber/fiber/v2"
)

func HandleWhopValidateWrapper(c *fiber.Ctx, pg postgres.PGAccount) error {
	license := c.Query("license")
	if license == "" {
		payload := invalidRequestResponse{
			ResponseCode: 400,
			Message:      "Required query parameter \"license\" is not present.",
		}
		return c.Status(fiber.StatusBadRequest).JSON(payload)
	}

	id, email, err := whop.CheckLicenseKey(license)
	if err != nil {
		if err.Error() == "WHOP_VALIDATE_LICENSE_API_KEY does not exist" {
			payload := invalidRequestResponse{
				ResponseCode: fiber.StatusServiceUnavailable, //503
				Message:      "",
			}
			return c.Status(fiber.StatusBadRequest).JSON(payload)
		}
		payload := &invalidRequestResponse{}
		codeMessagePair := strings.Split(err.Error(), "|")
		if len(codeMessagePair) == 2 {
			respCode, err2 := strconv.Atoi(codeMessagePair[0])
			if err2 != nil {
				payload := invalidRequestResponse{
					ResponseCode: fiber.StatusServiceUnavailable, //503
					Message:      err2.Error(),
				}

				return c.Status(fiber.StatusBadRequest).JSON(payload)
			}
			payload.ResponseCode = respCode
			payload.Message = codeMessagePair[1]
		} else {
			payload.ResponseCode = 404
			payload.Message = codeMessagePair[0]
		}

		return c.Status(fiber.StatusBadRequest).JSON(payload)
	}
	_, licenseActive := ActiveClients[id]
	if licenseActive {
		payload := &invalidRequestResponse{}
		payload.ResponseCode = 403
		payload.Message = "This license key already has an active session"
		return c.Status(fiber.StatusCreated).JSON(payload)
	}
	account := newAccount(license, email)
	err = pg.CreateAccount(account)
	if err != nil && err != sql.ErrNoRows {
		log.Print(err)
		payload := invalidRequestResponse{
			ResponseCode: fiber.StatusInternalServerError,
			Message:      "unable to create account",
		}
		log.Print(err)
		return c.Status(fiber.StatusCreated).JSON(payload)
	}

	if account.ID != 0 {
		log.Printf("created account, id: %d", account.ID)
	} else {
		log.Println("account already exists")
	}

	//
	ActiveClients[id] = &Client{
		Ts:     int(time.Now().UnixMilli()),
		WsConn: nil,
		Id:     id,
		Pool:   []*Pool{},
	}

	payload := validLicenseKeyResponse{
		ResponseCode: 201,
		Cid:          id,
	}
	return c.Status(fiber.StatusCreated).JSON(payload)

	// cid := uuid.New().String()

}

func HandleConfigureMonitorWrapper(c *fiber.Ctx, redisClient *cache.RedisClientWithContext) error {

	payload := &monitormanager.MonitorManagerMessage{}
	if err := c.BodyParser(payload); err != nil {
		payload := invalidRequestResponse{
			ResponseCode: fiber.StatusBadRequest,
			Message:      "invalid data in request body",
		}
		log.Print(err)
		return c.Status(fiber.StatusBadRequest).JSON(payload)

	}
	if payload.Name == "" {
		payload := invalidRequestResponse{
			ResponseCode: fiber.StatusBadRequest, //400
			Message:      `invalid "name" field`,
		}
		return c.Status(fiber.StatusBadRequest).JSON(payload)
	}
	valid, errMsg := checkMonitorManagerMessagePayload(payload)
	if !valid {
		payload := invalidRequestResponse{
			ResponseCode: fiber.StatusBadRequest,
			Message:      errMsg,
		}
		return c.Status(fiber.StatusBadRequest).JSON(payload)
	}

	payloadJsonBytes, err := json.Marshal(payload)
	if err != nil {
		log.Print(err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	err = redisClient.PushUpdate(cache.MonitorUpdateKey, payloadJsonBytes)
	if err != nil {
		log.Print(err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return nil

}
