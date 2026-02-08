package jwt

import (
	"errors"
	"fmt"
	"time"

	"shortvideo/pkg/config"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secretKey   string
	expireHours int
}

func NewJWTManager() *JWTManager {
	jwtConfig := config.Get().JWT
	return NewJWTManagerWithConfig(jwtConfig.Secret, jwtConfig.ExpireHours)
}

func NewJWTManagerWithConfig(secretKey string, expireHours int) *JWTManager {
	return &JWTManager{
		secretKey:   secretKey,
		expireHours: expireHours,
	}
}

func (j *JWTManager) GenerateToken(userID int64) (string, error) {
	expireTime := time.Now().Add(time.Hour * time.Duration(j.expireHours))
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("签名方法不符合预期: %v", token.Header["alg"])
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("令牌无效")
	}

	return claims, nil
}

func (j *JWTManager) GetUserIDFromToken(tokenString string) (int64, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return 0, err
	}

	return claims.UserID, nil
}
