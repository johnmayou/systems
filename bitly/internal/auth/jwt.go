package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var jwtSigningMethod = jwt.SigningMethodHS256

func IssueJwtToken(userID, email, jwtSecret string) (string, error) {
	token := jwt.NewWithClaims(
		jwtSigningMethod,
		jwt.MapClaims{
			"sub":   userID,
			"email": email,
			"roles": []string{"user"},
			"exp":   time.Now().Add(time.Hour).Unix(),
			"iat":   time.Now().Unix(),
			"jti":   uuid.New().String(),
		},
	)
	return token.SignedString([]byte(jwtSecret))
}
