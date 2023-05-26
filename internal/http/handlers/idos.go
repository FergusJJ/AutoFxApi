package handler

import "github.com/gofiber/websocket/v2"

type Client struct {
	Ts     int
	WsConn *websocket.Conn
}

var WsClients = make(map[string]*Client)

type apiAuthRequest struct {
	ApiKey string `json:"apiKey"`
}

type invalidRequestResponse struct {
	ResponseCode int    `json:"responseCode"`
	Message      string `json:"message"`
}

type validLicenseKeyResponse struct {
	ResponseCode int    `json:"responseCode"`
	Cid          string `json:"cid"`
}
