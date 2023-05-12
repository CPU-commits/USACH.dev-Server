package services

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/CPU-commits/USACH.dev-Server/res"
	"github.com/CPU-commits/USACH.dev-Server/settings"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var jwtKey = settings.GetSettings().JWT_SECRET_KEY
var jwtKeyRefresh = settings.GetSettings().JWT_SECRET_REFRESH

// Custom claims
type Claims struct {
	UserID string `json:"_id"`
	Role   string `json:"role,omitempty"`
	Name   string `json:"name,omitempty"`
	jwt.RegisteredClaims
}

func extractToken(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	strArr := strings.Split(bearerToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

func signToken(
	idUser primitive.ObjectID,
	role string,
	name string,
) (tokenStr, refreshTokenStr string, errorRes *res.ErrorRes) {
	// Create token access
	claims := Claims{
		idUser.Hex(),
		role,
		name,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "usach.dev",
			Subject:   "access",
			ID:        uuid.New().String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// To string
	tokenStr, err := token.SignedString([]byte(settingsData.JWT_SECRET_KEY))
	if err != nil {
		return "", "", &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusInternalServerError,
		}
	}
	// Create refresh token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		idUser.Hex(),
		"",
		"",
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "usach.dev",
			Subject:   "refresh",
			ID:        uuid.New().String(),
		},
	})
	// To string
	refreshTokenStr, err = refreshToken.SignedString([]byte(settingsData.JWT_SECRET_REFRESH))
	if err != nil {
		return "", "", &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusInternalServerError,
		}
	}
	return
}

func VerifyToken(r *http.Request, kind string) (*jwt.Token, error) {
	tokenString := extractToken(r)
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		if kind == "access" {
			return []byte(jwtKey), nil
		}
		return []byte(jwtKeyRefresh), nil
	})
	if token == nil {
		return nil, errors.New("Unauthorized")
	}
	if !token.Valid {
		return nil, errors.New("Unauthorized")
	}
	if err != nil {
		return nil, err
	}
	return token, nil
}

func ExtractTokenMetadata(token *jwt.Token) (*Claims, error) {
	claim := token.Claims.(jwt.MapClaims)
	return &Claims{
		fmt.Sprintf("%v", claim["_id"]),
		fmt.Sprintf("%v", claim["role"]),
		fmt.Sprintf("%v", claim["name"]),
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Unix(int64(claim["exp"].(float64)), 0)),
			NotBefore: jwt.NewNumericDate(time.Unix(int64(claim["nbf"].(float64)), 0)),
			Issuer:    claim["iss"].(string),
			Subject:   claim["sub"].(string),
			ID:        claim["jti"].(string),
		},
	}, nil
}

func NewClaimsFromContext(ctx *gin.Context) (*Claims, bool) {
	user, exists := ctx.Get("user")
	if !exists {
		return &Claims{}, false
	}
	return &Claims{
		UserID: user.(*Claims).UserID,
		Role:   user.(*Claims).Role,
		Name:   user.(*Claims).Name,
	}, true
}
