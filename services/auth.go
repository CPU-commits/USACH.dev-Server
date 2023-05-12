package services

import (
	"errors"
	"net/http"

	"github.com/CPU-commits/USACH.dev-Server/db"
	"github.com/CPU-commits/USACH.dev-Server/forms"
	"github.com/CPU-commits/USACH.dev-Server/models"
	"github.com/CPU-commits/USACH.dev-Server/res"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct{}

func (auth *AuthService) Login(loginForm *forms.LoginForm) (map[string]interface{}, *res.ErrorRes) {
	user, errRes := userService.FindByEmail(loginForm.Email)
	if errRes != nil {
		return nil, errRes
	}
	err := bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(loginForm.Password),
	)
	if err != nil {
		return nil, &res.ErrorRes{
			Err:        errors.New("credenciales inválidas"),
			StatusCode: http.StatusUnauthorized,
		}
	}
	if !user.Status {
		return nil, &res.ErrorRes{
			Err:        errors.New("activa tu cuenta a través del correo que te envíamos"),
			StatusCode: http.StatusUnauthorized,
		}
	}
	// Create tokens
	tokenStr, refreshTokenStr, errRes := signToken(
		user.ID,
		user.Role,
		user.FullName,
	)
	if errRes != nil {
		return nil, errRes
	}
	// Make response
	response := make(map[string]interface{})
	response["token"] = tokenStr
	response["refresh_token"] = refreshTokenStr
	response["user"] = map[string]string{
		"name":  user.FullName,
		"email": user.Email,
		"role":  user.Role,
		"_id":   user.ID.Hex(),
	}

	return response, nil
}

func (auth *AuthService) RefreshToken(r *http.Request) (string, *res.ErrorRes) {
	token, err := VerifyToken(r, "refresh")
	if err != nil {
		return "", &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusUnauthorized,
		}
	}
	// Data
	claims, _ := ExtractTokenMetadata(token)

	idObjectUser, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		return "", &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	// Get user
	var user models.User

	cursor := userModel.Use().FindOne(db.Ctx, bson.D{{
		Key:   "_id",
		Value: idObjectUser,
	}})
	if err := cursor.Decode(&user); err != nil {
		return "", &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	// Create tokens
	tokenStr, _, errRes := signToken(
		idObjectUser,
		user.Role,
		user.FullName,
	)
	if errRes != nil {
		return "", errRes
	}

	return tokenStr, nil
}

func NewAuthService() *AuthService {
	return &AuthService{}
}
