package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/CPU-commits/USACH.dev-Server/db"
	"github.com/CPU-commits/USACH.dev-Server/models"
	"github.com/CPU-commits/USACH.dev-Server/res"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Service
type CommentService struct{}

func (*CommentService) GetComments(
	idDiscussion,
	idUser string,
) ([]*models.CommentRes, *res.ErrorRes) {
	// ObjectID
	idObjDiscussion, err := primitive.ObjectIDFromHex(idDiscussion)
	if err != nil {
		return nil, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	// Get discussion and has access
	errRes := discussionService.HasAccess(idDiscussion, idUser)
	if errRes != nil {
		return nil, errRes
	}
	// Get comments
	var comments []*models.CommentRes

	cursor, err := commentModel.Use().Aggregate(db.Ctx, mongo.Pipeline{
		bson.D{{
			Key: "$match",
			Value: bson.M{
				"discussion": idObjDiscussion,
				"is_res":     false,
			},
		}},
		bson.D{{
			Key: "$lookup",
			Value: bson.M{
				"from":         models.USERS_COLLECTION,
				"localField":   "author",
				"foreignField": "_id",
				"as":           "author",
				"pipeline": bson.A{
					bson.D{{
						Key: "$project",
						Value: bson.M{
							"full_name": 1,
							"username":  1,
						},
					}},
				},
			},
		}},
		bson.D{{
			Key: "$lookup",
			Value: bson.M{
				"from":         models.COMMENTS_COLLECTION,
				"localField":   "responses",
				"foreignField": "_id",
				"as":           "responses",
				"pipeline": bson.A{
					bson.D{{
						Key: "$lookup",
						Value: bson.M{
							"from":         models.USERS_COLLECTION,
							"localField":   "author",
							"foreignField": "_id",
							"as":           "author",
							"pipeline": bson.A{
								bson.D{{
									Key: "$project",
									Value: bson.M{
										"full_name": 1,
										"username":  1,
									},
								}},
							},
						},
					}},
					bson.D{{
						Key:   "$sort",
						Value: bson.M{"created_at": 1},
					}},
					bson.D{{
						Key: "$addFields",
						Value: bson.M{
							"author": bson.M{
								"$arrayElemAt": bson.A{"$author", 0},
							},
						},
					}},
				},
			},
		}},
		bson.D{{
			Key: "$addFields",
			Value: bson.M{
				"author": bson.M{
					"$arrayElemAt": bson.A{"$author", 0},
				},
			},
		}},
		bson.D{{Key: "$sort", Value: bson.M{"created_at": -1}}},
	})
	if err != nil {
		return nil, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	if err := cursor.All(db.Ctx, &comments); err != nil {
		fmt.Printf("err: %v\n", err)
		return nil, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}

	return comments, nil
}

func (*CommentService) Comment(
	comment,
	discussion,
	idUser,
	replyComment string,
) (*primitive.ObjectID, *res.ErrorRes) {
	// ID Objects
	idObjUser, err := primitive.ObjectIDFromHex(idUser)
	if err != nil {
		return nil, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	idObjDiscussion, err := primitive.ObjectIDFromHex(discussion)
	if err != nil {
		return nil, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	// Check comment to reply
	var idObjComment primitive.ObjectID

	if replyComment != "" {
		idObjComment, err = primitive.ObjectIDFromHex(replyComment)
		if err != nil {
			return nil, &res.ErrorRes{
				Err:        err,
				StatusCode: http.StatusBadRequest,
			}
		}

		existsComment, err := commentModel.Exists(bson.D{{
			Key:   "_id",
			Value: idObjComment,
		}})
		if err != nil {
			return nil, &res.ErrorRes{
				Err:        err,
				StatusCode: http.StatusServiceUnavailable,
			}
		}
		if !existsComment {
			return nil, &res.ErrorRes{
				Err:        errors.New("el comentario que tratas de responder no existe"),
				StatusCode: http.StatusNotFound,
			}
		}
	}
	// Upload comment
	modelComment := commentModel.NewModel(
		comment,
		idObjDiscussion,
		idObjUser,
		replyComment != "",
	)
	insertedId, err := commentModel.Upload(modelComment)
	if err != nil {
		return nil, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	// Update comment reply
	if replyComment != "" {
		_, err = commentModel.Use().UpdateOne(
			db.Ctx,
			bson.D{{
				Key:   "_id",
				Value: idObjComment,
			}},
			bson.D{{
				Key: "$addToSet",
				Value: bson.M{
					"responses": insertedId,
				},
			}},
		)
		if err != nil {
			return nil, &res.ErrorRes{
				Err:        err,
				StatusCode: http.StatusServiceUnavailable,
			}
		}
	}
	// Queue
	modelQueue, err := modelComment.ToQueue(insertedId.Hex(), replyComment)
	if err != nil {
		return nil, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	jsonModel, err := json.Marshal(modelQueue)
	if err != nil {
		return nil, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusInternalServerError,
		}
	}

	pubSubClient.Emit("notifications:ws:comment", jsonModel)
	return &insertedId, nil
}

func NewCommentService() *CommentService {
	return &CommentService{}
}
