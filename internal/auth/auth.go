package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/constants"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	return argon2id.CreateHash(password, argon2id.DefaultParams)
}

func CheckPasswordHash(password string, hash string) (bool, error) {
	match, _, err := argon2id.CheckHash(password, hash)

	return match, err
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	now := time.Now()
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer: "chirpy",
		Subject: userID.String(),
		ID: uuid.New().String(),
		IssuedAt: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
	})

	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString string, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error validating token: %v", err)
	}

	if !token.Valid {
		return uuid.UUID{}, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return uuid.UUID{}, fmt.Errorf("invalid token claims")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("invalid user ID: %v", err)
	}

	return userID, nil
}

func GetBearerToken(r *http.Request) (string, error) {
	token := r.Header.Get(constants.AUTHORIZATION_HEADER)
	if token == "" {
		return "", fmt.Errorf("no authorization header provided")
	}
	
	if !strings.HasPrefix(token, constants.BEARER_TOKEN_PREFIX) {
		return "", fmt.Errorf("invalid authorization header format")
	}

	trimmedToken := strings.TrimPrefix(token, constants.BEARER_TOKEN_PREFIX)
	return trimmedToken, nil
}

func MakeRefreshToken() (string, error) {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("error generating refresh token: %v", err)
	}

	refreshToken := hex.EncodeToString(randomBytes)
	return refreshToken, nil
}


func GetAPIKey(r *http.Request) (string, error) {
	token := r.Header.Get(constants.AUTHORIZATION_HEADER)
	if token == "" {
		return "", fmt.Errorf("no authorization header provided")
	}

	if !strings.HasPrefix(token, constants.POLKA_WEBHOOK_TOKEN_PREFIX) {
		return "", fmt.Errorf("invalid authorization header format")
	}

	trimmedToken := strings.TrimPrefix(token, constants.POLKA_WEBHOOK_TOKEN_PREFIX)
	return trimmedToken, nil
}
