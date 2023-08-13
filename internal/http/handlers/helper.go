package handler

import (
	"api/config"
	monitormanager "api/internal/monitor-manager"
	"api/internal/storage/postgres"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func checkMonitorManagerMessagePayload(payload *monitormanager.MonitorManagerMessage) (bool, string) {
	if payload.Option != 0 && payload.Option != 1 {
		return false, `invalid "option" field, value must be 1 or 0`
	}
	if payload.Type != monitormanager.ICMARKETS {
		return false, `invalid "type" field`
	}
	return true, ""
}

func newAccount(license, email string) *postgres.Account {
	return &postgres.Account{
		Email:     email,
		License:   license,
		CreatedAt: time.Now().UTC(),
	}
}

func generateTokenPair(license string, id int) (string, string, int64, error) {
	signingKey, err := config.Config("JWT_SIGNING_KEY")
	if err != nil {
		return "", "", -1, err
	}
	refreshSigningKey, err := config.Config("JWT_REFRESH_SIGNING_KEY")
	if err != nil {
		return "", "", -1, err
	}

	accessTokenClaims := jwt.MapClaims{
		"license": license,
		"sub":     strconv.Itoa(id),
		"exp":     time.Now().Add(time.Hour * 6).Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	signedAccessToken, err := accessToken.SignedString([]byte(signingKey))
	if err != nil {
		return "", "", -1, err
	}
	refreshExpiry := time.Now().Add(time.Hour * 24 * 7).Unix()
	refreshTokenClaims := jwt.MapClaims{
		"sub": strconv.Itoa(id),
		"exp": refreshExpiry,
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	signedRefreshToken, err := refreshToken.SignedString([]byte(refreshSigningKey))
	if err != nil {
		return "", "", -1, err
	}
	return signedAccessToken, signedRefreshToken, refreshExpiry, nil

}
