package handler

import (
	"api/internal/storage/postgres"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

/*
WHOP/VALIDATE
/AUTH
gets tokens, user data also created/
now give access via jwt to account info and user_data info
*/
func HandleDeleteUserPositionWrapper(c *fiber.Ctx, pg postgres.PGPosition) error {
	type deletePositionReqBody struct {
		PositionID string `json:"positionID"`
	}
	var jwtSub int
	var subAsString string
	userToken := c.Locals("user").(*jwt.Token)
	if claims, ok := userToken.Claims.(jwt.MapClaims); ok {
		if sub, exists := claims["sub"]; exists {
			subAsString = sub.(string)
			subInt, err := strconv.Atoi(subAsString)
			if err != nil {
				log.Print(err)
				return c.SendStatus(fiber.StatusInternalServerError)
			}
			jwtSub = subInt
		}
	}
	deletePositionReq := &deletePositionReqBody{}
	if err := c.BodyParser(deletePositionReq); err != nil {
		log.Print(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid data in request body"})
	}
	if deletePositionReq.PositionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "failed to delete position cannot be empty"})
	}
	err := pg.DeleteUserPosition(jwtSub, deletePositionReq.PositionID)
	if err != nil {
		log.Print(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": fmt.Sprintf("failed to delete position with id: %s", deletePositionReq.PositionID)})
	}
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"message": fmt.Sprintf("position %s deleted", deletePositionReq.PositionID)})
}

func HandleGetAllUserPositionWrapper(c *fiber.Ctx, pg postgres.PGPosition) error {
	var jwtSub int
	var subAsString string
	userToken := c.Locals("user").(*jwt.Token)
	if claims, ok := userToken.Claims.(jwt.MapClaims); ok {
		if sub, exists := claims["sub"]; exists {
			subAsString = sub.(string)
			subInt, err := strconv.Atoi(subAsString)
			if err != nil {
				log.Print(err)
				return c.SendStatus(fiber.StatusInternalServerError)
			}
			jwtSub = subInt
		}
	}
	positions, err := pg.GetAllPositionsByAccountID(jwtSub)
	if err != nil {
		log.Print(err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	data := []map[string]string{}

	/*
		PositionID      string    `json:"positionID"`
		AccountID       int       `json:"accountID"`
		CopyPositionID  string    `json:"copyPositionID"`
		OpenedTimestamp time.Time `json:"openedTimestamp"`
		Symbol          string    `json:"symbol"`
		SymbolID        int       `json:"symbolID"`
		Volume          int       `json:"volume"`
		Side            string    `json:"Side"`
		AveragePrice    string    `json:"averagePrice"`
	*/

	for _, v := range positions {
		entry := map[string]string{
			"positionID":      v.PositionID,
			"copyPositionID":  v.CopyPositionID,
			"openedTimestamp": v.OpenedTimestamp,
			"symbol":          v.Symbol,
			"symbolID":        fmt.Sprint(v.SymbolID),
			"volume":          fmt.Sprint(v.Volume),
			"side":            v.Side,
			"averagePrice":    v.AveragePrice,
		}
		data = append(data, entry)
	}
	return c.Status(fiber.StatusOK).JSON(data)
}

func HandleCreateUserPositionWrapper(c *fiber.Ctx, pg postgres.PGPosition) error {
	type createPositionReqBody struct {
		PositionID      string `json:"positionID"`
		CopyPositionID  string `json:"copyPositionID"`
		OpenedTimestamp string `json:"openedTimestamp"`
		Symbol          string `json:"symbol"`
		SymbolID        string `json:"symbolID"`
		Volume          string `json:"volume"`
		Side            string `json:"Side"`
		AveragePrice    string `json:"averagePrice"`
	}
	var jwtSub int
	var subAsString string
	userToken := c.Locals("user").(*jwt.Token)
	if claims, ok := userToken.Claims.(jwt.MapClaims); ok {
		if sub, exists := claims["sub"]; exists {
			subAsString = sub.(string)
			subInt, err := strconv.Atoi(subAsString)
			if err != nil {
				log.Print(err)
				return c.SendStatus(fiber.StatusInternalServerError)
			}
			jwtSub = subInt
		}
	}
	createPositionReq := &createPositionReqBody{}
	if err := c.BodyParser(createPositionReq); err != nil {
		log.Print(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid data in request body"})
	}
	if createPositionReq.PositionID == "" || createPositionReq.CopyPositionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "failed to store position, fields cannot be empty"})
	}

	symbolIDInt, err := strconv.Atoi(createPositionReq.SymbolID)
	if err != nil {
		log.Print(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid data in request body, symbol"})
	}
	// volumeInt := int(newPos.Volume)
	volumeInt, err := strconv.Atoi(createPositionReq.Volume)
	if err != nil {
		log.Print(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid data in request body, volume"})
	}

	userPosition := &postgres.UserPosition{
		PositionID:      createPositionReq.PositionID,
		CopyPositionID:  createPositionReq.CopyPositionID,
		AccountID:       jwtSub,
		OpenedTimestamp: createPositionReq.OpenedTimestamp,
		Symbol:          createPositionReq.Symbol,
		SymbolID:        symbolIDInt,
		Volume:          volumeInt,
		Side:            createPositionReq.Side,
		AveragePrice:    createPositionReq.AveragePrice,
	}
	err = pg.CreateUserPosition(userPosition)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "unable to store duplicate position"})
		}
		log.Print(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "failed to store position"})
	}
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"message": "position stored"})
}
