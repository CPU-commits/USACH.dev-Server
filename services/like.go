package services

import (
	"errors"
	"net/http"

	"github.com/CPU-commits/USACH.dev-Server/db"
	"github.com/CPU-commits/USACH.dev-Server/models"
	"github.com/CPU-commits/USACH.dev-Server/res"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type LikeService struct{}

func (l *LikeService) ToggleLike(
	like *models.Like,
	idUser,
	idRepository primitive.ObjectID,
	plus bool,
) *res.ErrorRes {
	// Plus one or -one to repo
	var toSum int
	if like == nil {
		// Insert like
		likeModel := likesModel.NewLikeModel(idUser, idRepository, plus)

		_, err := likesModel.Use().InsertOne(db.Ctx, likeModel)
		if err != nil {
			return &res.ErrorRes{
				Err:        err,
				StatusCode: http.StatusServiceUnavailable,
			}
		}
		if plus {
			toSum = 1
		} else {
			toSum = -1
		}
	} else {
		// Update like
		_, err := likesModel.Use().UpdateByID(db.Ctx, like.ID, bson.D{{
			Key: "$set",
			Value: bson.M{
				"plus": plus,
			},
		}})
		if err != nil {
			return &res.ErrorRes{
				Err:        err,
				StatusCode: http.StatusServiceUnavailable,
			}
		}

		if plus {
			toSum = 2
		} else {
			toSum = -2
		}
	}

	_, err := repoModel.Use().UpdateByID(db.Ctx, idRepository, bson.D{{
		Key: "$inc",
		Value: bson.M{
			"stars": toSum,
		},
	}})
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}

	return nil
}

func (l *LikeService) GetLike(
	idUser,
	idRepository primitive.ObjectID,
) (*models.Like, *res.ErrorRes) {
	var like *models.Like

	cursor := likesModel.Use().FindOne(db.Ctx, bson.D{
		{
			Key:   "user",
			Value: idUser,
		},
		{
			Key:   "repository",
			Value: idRepository,
		},
	})
	if err := cursor.Decode(&like); err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return nil, &res.ErrorRes{
				Err:        err,
				StatusCode: http.StatusServiceUnavailable,
			}
		}
	}

	return like, nil
}

func NewLikeService() *LikeService {
	return &LikeService{}
}
