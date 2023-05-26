package handler

import (
	"strconv"
	"strings"
	"time"

	"api/pkg/whop"

	"github.com/gofiber/fiber/v2"
)

/*

on connection sendHeartbeat function starts


*/

type sockMessage struct {
	Action  string `json:"action"`
	Message string `json:"message"`
}

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

	_, licenseActive := WsClients[id]
	if licenseActive {
		payload := &invalidRequestResponse{}
		payload.ResponseCode = 403
		payload.Message = "This license key already has an active session"
		return c.Status(fiber.StatusCreated).JSON(payload)
	}
	//
	WsClients[id] = &Client{
		Ts:     int(time.Now().UnixMilli()),
		WsConn: nil,
	}
	payload := validLicenseKeyResponse{
		ResponseCode: 201,
		Cid:          id,
	}
	return c.Status(fiber.StatusCreated).JSON(payload)

	// cid := uuid.New().String()

}
