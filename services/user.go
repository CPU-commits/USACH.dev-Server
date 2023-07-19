package services

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/CPU-commits/USACH.dev-Server/db"
	"github.com/CPU-commits/USACH.dev-Server/forms"
	"github.com/CPU-commits/USACH.dev-Server/models"
	"github.com/CPU-commits/USACH.dev-Server/notifications/email"
	"github.com/CPU-commits/USACH.dev-Server/res"
	"github.com/CPU-commits/USACH.dev-Server/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct{}

func (u *UserService) isOwner(
	username string,
	idUser primitive.ObjectID,
) (bool, *res.ErrorRes) {
	var user *models.User

	opts := options.FindOne().SetProjection(bson.D{{
		Key:   "_id",
		Value: 1,
	}})
	pointer := userModel.Use().FindOne(db.Ctx, bson.D{
		{
			Key:   "_id",
			Value: idUser,
		},
		{
			Key:   "username",
			Value: username,
		},
	}, opts)
	if err := pointer.Decode(&user); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	return true, nil
}

func (u *UserService) FindByEmail(email string) (*models.User, *res.ErrorRes) {
	var user *models.User

	cursor := userModel.Use().FindOne(db.Ctx, bson.D{{
		Key:   "email",
		Value: email,
	}})
	if err := cursor.Decode(&user); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, &res.ErrorRes{
				Err:        errors.New("no existe el usuario"),
				StatusCode: http.StatusNotFound,
			}
		} else {
			return nil, &res.ErrorRes{
				Err:        err,
				StatusCode: http.StatusServiceUnavailable,
			}
		}
	}
	return user, nil
}

func (u *UserService) GetByUsername(username string, getProfile bool) (*models.UserRes, *res.ErrorRes) {
	var user []*models.UserRes

	// Pipeline
	pipeline := mongo.Pipeline{
		bson.D{{
			Key: "$match",
			Value: bson.M{
				"username": username,
				"status":   true,
			},
		}},
		bson.D{{
			Key: "$project",
			Value: bson.M{
				"status":   0,
				"email":    0,
				"password": 0,
			},
		}},
	}
	if getProfile {
		pipeline = append(
			pipeline,
			bson.D{{
				Key: "$lookup",
				Value: bson.M{
					"from":         models.PROFILE_COLLECTION,
					"localField":   "profile",
					"foreignField": "_id",
					"as":           "profile",
				},
			}},
			bson.D{{
				Key: "$addFields",
				Value: bson.M{
					"profile": bson.M{
						"$arrayElemAt": bson.A{"$profile", 0},
					},
				},
			}},
		)
	} else {
		pipeline = append(pipeline, bson.D{{
			Key: "$project",
			Value: bson.M{
				"profile": 0,
			},
		}})
	}

	pointer, err := userModel.Use().Aggregate(db.Ctx, pipeline)
	if err != nil {
		return nil, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	if err := pointer.All(db.Ctx, &user); err != nil {
		return nil, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	if len(user) == 0 {
		return nil, nil
	}
	return user[0], nil
}

func (u *UserService) GetUser(username string, getProfile bool) (*models.UserRes, *res.ErrorRes) {
	// Get user
	user, err := u.GetByUsername(username, getProfile)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (uS *UserService) GetAvatar(username string, w io.Writer) *res.ErrorRes {
	user, err := uS.GetByUsername(username, true)
	if err != nil {
		return err
	}
	if user.Profile != nil {
		avatar, err := utils.GetFile(user.Profile.Avatar)
		if err != nil {
			return &res.ErrorRes{
				Err:        err,
				StatusCode: http.StatusInternalServerError,
			}
		}
		w.Write(avatar)
	}

	return nil
}

func (u *UserService) CreateUser(userForm *forms.UserForm) *res.ErrorRes {
	if !strings.Contains(userForm.Email, "@usach.cl") {
		return &res.ErrorRes{
			Err:        errors.New("el correo debe ser usach"),
			StatusCode: http.StatusBadRequest,
		}
	}
	// Check if exists the email
	var userExists *models.User
	cursor := userModel.Use().FindOne(db.Ctx, bson.D{{
		Key:   "email",
		Value: userForm.Email,
	}})
	if err := cursor.Decode(&userExists); err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return &res.ErrorRes{
				Err:        err,
				StatusCode: http.StatusBadRequest,
			}
		}
	}
	if userExists != nil {
		return &res.ErrorRes{
			Err:        errors.New("el usuario ya existe"),
			StatusCode: http.StatusBadRequest,
		}
	}
	// Create model
	passwordHashed, err := bcrypt.GenerateFromPassword(
		[]byte(userForm.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	newUserModel := userModel.NewModel(
		userForm.FullName,
		userForm.Email,
		string(passwordHashed),
		models.USER,
	)
	// Insert model
	insertedUser, err := userModel.Use().InsertOne(db.Ctx, newUserModel)
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	idUser := insertedUser.InsertedID.(primitive.ObjectID)
	// Create token to confirm user
	finishDate := time.Now().Add(24 * 7 * time.Hour)
	userToken, err := usersTokenModel.NewModel(
		idUser,
		primitive.NewDateTimeFromTime(finishDate),
		[]string{models.PERMISSION_CONFIRM_USER},
	)
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusInternalServerError,
		}
	}
	_, err = usersTokenModel.Use().InsertOne(db.Ctx, userToken)
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	// Send email to confirm user
	err = email.SendEmail(&email.MailConfig{
		From:     "info@usach.dev",
		To:       userForm.Email,
		Subject:  "Confirma tu correo - USACH.dev",
		Template: email.TEMPLATE_VALIDATE_USER,
		TemplateParams: map[string]string{
			"{{ BACKEND_URL }}":   "https://backend.usach.dev",
			"{{ CONFIRM_TOKEN }}": userToken.Token,
		},
	})
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}

	return nil
}

func (u *UserService) ActivateUser(token string) *res.ErrorRes {
	// Search token
	var userToken *models.UserToken

	cursor := usersTokenModel.Use().FindOne(db.Ctx, bson.D{{
		Key:   "token",
		Value: token,
	}})
	if err := cursor.Decode(&userToken); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return &res.ErrorRes{
				Err:        errors.New("el token no es válido"),
				StatusCode: http.StatusUnauthorized,
			}
		} else {
			return &res.ErrorRes{
				Err:        err,
				StatusCode: http.StatusServiceUnavailable,
			}
		}
	}
	// Check if is not expired
	now := time.Now()
	if now.After(userToken.FinishDate.Time()) {
		return &res.ErrorRes{
			Err:        errors.New("el token no es válido"),
			StatusCode: http.StatusUnauthorized,
		}
	}
	// Check permissions
	flag := false
	for _, permission := range userToken.Permissions {
		if permission == models.PERMISSION_CONFIRM_USER {
			flag = true
			break
		}
	}
	if !flag {
		return &res.ErrorRes{
			Err:        errors.New("el token no es válido"),
			StatusCode: http.StatusUnauthorized,
		}
	}
	// Active
	_, err := userModel.Use().UpdateByID(db.Ctx, userToken.User, bson.D{{
		Key: "$set",
		Value: bson.M{
			"status": true,
		},
	}})
	if err != nil {
		return &res.ErrorRes{
			Err:        errors.New("no se pudo activar tu cuenta, intenta nuevamente"),
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	// Get user
	var user models.User

	cursor = userModel.Use().FindOne(db.Ctx, bson.D{{
		Key:   "_id",
		Value: userToken.User,
	}})
	if err := cursor.Decode(&user); err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}

	return nil
}

func (*UserService) UpdateProfile(
	idUser string,
	profile *forms.ProfileForm,
	avatarFile *multipart.FileHeader,
) *res.ErrorRes {
	// ObjectID
	idObjUser, err := primitive.ObjectIDFromHex(idUser)
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	// Upload avatar
	var avatar string

	if avatarFile != nil {
		avatar, err = utils.UploadFile(avatarFile)
		if err != nil {
			return &res.ErrorRes{
				Err:        err,
				StatusCode: http.StatusInternalServerError,
			}
		}
	}
	// Set profile
	exists, err := profileModel.Exists(bson.D{{
		Key:   "user",
		Value: idObjUser,
	}})
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	if !exists {
		modelProfile := profileModel.NewModel(
			idObjUser,
			profile,
			avatar,
		)
		insertedId, err := profileModel.Use().InsertOne(db.Ctx, modelProfile)
		if err != nil {
			return &res.ErrorRes{
				Err:        err,
				StatusCode: http.StatusServiceUnavailable,
			}
		}
		_, err = userModel.Use().UpdateOne(
			db.Ctx,
			bson.D{{
				Key:   "_id",
				Value: idObjUser,
			}},
			bson.D{{
				Key: "$set",
				Value: bson.M{
					"profile": insertedId.InsertedID,
				},
			},
			})
		if err != nil {
			return &res.ErrorRes{
				Err:        err,
				StatusCode: http.StatusServiceUnavailable,
			}
		}
	} else {
		var update bson.D
		if profile != nil && profile.Description != "" {
			update = append(update, bson.E{
				Key:   "description",
				Value: profile.Description,
			})
		}
		if avatar != "" {
			update = append(update, bson.E{
				Key:   "avatar",
				Value: avatar,
			})
		}

		_, err := profileModel.Use().UpdateOne(db.Ctx, bson.D{{
			Key:   "user",
			Value: idObjUser,
		}}, bson.D{{
			Key:   "$set",
			Value: update,
		}})
		if err != nil {
			return &res.ErrorRes{
				Err:        err,
				StatusCode: http.StatusServiceUnavailable,
			}
		}
	}
	return nil
}

func NewUserService() *UserService {
	return &UserService{}
}
