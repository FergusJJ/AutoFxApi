package handler

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	monitormanager "api/internal/monitor-manager"
	"api/internal/storage"
	"api/pkg/whop"

	"github.com/gofiber/fiber/v2"
)

/*

on connection sendHeartbeat function starts


*/

type API_KEY string

const (
	registerAction   = "register"
	unregisterAction = "unregister"
	ackAction        = "ack"
)

// var register = make(chan *websocket.Conn)
// var unregister = make(chan *websocket.Conn)
// var heartbeat = make(chan *websocket.Conn)
// var ack = make(chan *websocket.Conn)
// var errResp = make(chan *websocket.Conn)

// var activeCids = make(map[string]API_KEY)
// var wsClients = make(map[*websocket.Conn]client)
// var AcceptedContentTypes = []string{"application/json"}

// could check within some cache first if want to avoid spamming whop
func HandleWhopValidate(c *fiber.Ctx) error {
	license := c.Query("license")
	if license == "" {
		payload := invalidRequestResponse{
			ResponseCode: 400,
			Message:      "Required query parameter \"license\" is not present.",
		}
		return c.Status(fiber.StatusBadRequest).JSON(payload)
	}
	id, err := whop.CheckLicenseKey(license)
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

	//

	_, licenseActive := ActiveClients[id]
	if licenseActive {
		payload := &invalidRequestResponse{}
		payload.ResponseCode = 403
		payload.Message = "This license key already has an active session"
		return c.Status(fiber.StatusCreated).JSON(payload)
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

func HandleConfigureMonitorWrapper(c *fiber.Ctx, redisClient *storage.RedisClientWithContext) error {

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
	err = redisClient.PushUpdate(storage.MonitorUpdateKey, payloadJsonBytes)
	if err != nil {
		log.Print(err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return nil

}

// get user by id, return list ok CopyPID:PID pairs
func HandleGetAllUserPositions(c *fiber.Ctx) error {
	userId := c.Params("id")
	log.Printf("user id: %s", userId)

	//probably check with something whether user is allowed in order to stop database queries that are unnecessary

	//if user doesn't exist then send 404

	data := map[string]string{}
	return c.Status(fiber.StatusOK).JSON(data)
}

// take list of CopyPID:PID pairs
func HandlePushNewUserPositions(c *fiber.Ctx) error {
	userId := c.Params("id")

	log.Printf("user id: %s", userId)
	data := map[string]string{}
	return c.Status(fiber.StatusCreated).JSON(data)
}

// take list of CopyPIDs
func HandlePopUserPositions(c *fiber.Ctx) error {
	userId := c.Params("id")

	log.Printf("user id: %s", userId)
	data := map[string]string{}
	return c.Status(fiber.StatusAccepted).JSON(data)
}
