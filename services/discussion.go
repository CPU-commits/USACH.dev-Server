package services

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"sync"

	"github.com/CPU-commits/USACH.dev-Server/db"
	"github.com/CPU-commits/USACH.dev-Server/forms"
	"github.com/CPU-commits/USACH.dev-Server/models"
	"github.com/CPU-commits/USACH.dev-Server/res"
	"github.com/CPU-commits/USACH.dev-Server/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DiscussionService struct{}

// Pipeline
func (d *DiscussionService) pipeline(options struct {
	Page          int
	Limit         int
	IDObjUser     primitive.ObjectID
	IsDiscussions bool
	AppendFirst   bson.D
}) mongo.Pipeline {
	pipeline := mongo.Pipeline{}

	if options.AppendFirst != nil {
		pipeline = append(pipeline, options.AppendFirst)
	}
	if options.IsDiscussions {
		pipeline = append(pipeline,
			bson.D{{Key: "$skip", Value: options.Page * 15}},
			bson.D{{Key: "$limit", Value: options.Limit}},
		)
	}
	pipeline = append(pipeline,
		bson.D{{Key: "$sort", Value: bson.M{"created_at": -1}}},
		bson.D{{
			Key: "$lookup",
			Value: bson.M{
				"from":         models.REACTION_COLLECTION,
				"localField":   "_id",
				"foreignField": "discussion",
				"as":           "reactions",
				"pipeline": bson.A{
					bson.D{{
						Key: "$group",
						Value: bson.M{
							"_id":   "$reaction",
							"count": bson.M{"$count": bson.M{}},
						},
					}},
					bson.D{{
						Key: "$project",
						Value: bson.M{
							"_id":      0,
							"reaction": "$_id",
							"count":    "$count",
						},
					}},
					bson.D{{
						Key: "$sort",
						Value: bson.M{
							"count": -1,
						},
					}},
				},
			},
		}},
		bson.D{{
			Key: "$lookup",
			Value: bson.M{
				"from": models.REACTION_COLLECTION,
				"let": bson.M{
					"id_discussion": "$_id",
					"id_user":       options.IDObjUser,
				},
				"pipeline": bson.A{
					bson.D{{
						Key: "$match",
						Value: bson.M{
							"$expr": bson.M{
								"$and": bson.A{
									bson.M{
										"$eq": bson.A{"$discussion", "$$id_discussion"},
									},
									bson.M{
										"$eq": bson.A{"$user", "$$id_user"},
									},
								},
							},
						},
					}},
					bson.D{{
						Key: "$project",
						Value: bson.M{
							"_id":      0,
							"reaction": "$reaction",
						},
					}},
				},
				"as": "user_reaction",
			},
		}},
		bson.D{{
			Key: "$unwind",
			Value: bson.M{
				"path":                       "$user_reaction",
				"preserveNullAndEmptyArrays": true,
			},
		}},
	)
	if options.IsDiscussions {
		pipeline = append(pipeline, bson.D{{
			Key:   "$project",
			Value: bson.M{"text": false},
		}})
	}

	return pipeline
}

func (d *DiscussionService) GetDiscussion(
	codeDiscussion,
	idUser string,
) (*models.DiscussionRes, *res.ErrorRes) {
	// Objects
	idObjUser, err := primitive.ObjectIDFromHex(idUser)
	if err != nil && idUser != "" {
		return nil, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}

	// Get discussion
	pipeline := d.pipeline(struct {
		Page          int
		Limit         int
		IDObjUser     primitive.ObjectID
		IsDiscussions bool
		AppendFirst   primitive.D
	}{
		IsDiscussions: false,
		AppendFirst: bson.D{{
			Key: "$match",
			Value: bson.M{
				"code": codeDiscussion,
			},
		}},
		IDObjUser: idObjUser,
	})
	// Find
	cursor, err := discussionModel.Use().Aggregate(db.Ctx, pipeline)
	if err != nil {
		return nil, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}

	var discussions []*models.DiscussionRes
	if err := cursor.All(db.Ctx, &discussions); err != nil {
		return nil, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	if len(discussions) == 0 {
		return nil, &res.ErrorRes{
			Err:        errors.New("no existe la discusión"),
			StatusCode: http.StatusNotFound,
		}
	}
	// Has access user
	discussion := discussions[0]
	if discussion.Repository != primitive.NilObjectID {
		hasAccess, err := repoService.HasRepoAccess(
			discussion.Repository,
			idUser,
		)
		if err != nil {
			return nil, err
		}
		if !hasAccess {
			return nil, &res.ErrorRes{
				Err:        errors.New("no tienes acceso a esta discución"),
				StatusCode: http.StatusUnauthorized,
			}
		}
	}

	return discussion, nil
}

func (d *DiscussionService) GetDiscussions(
	page int,
	sortBy,
	idRepository,
	idUser,
	search string,
	getTotal bool,
) ([]*models.DiscussionRes, int64, *res.ErrorRes) {
	// Objects
	idObjUser, err := primitive.ObjectIDFromHex(idUser)
	if err != nil && idUser != "" {
		return nil, 0, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	// AppendFirst
	var appendFirst bson.D
	// Query
	if idRepository != "" {
		idObjRepository, err := primitive.ObjectIDFromHex(idRepository)
		if err != nil {
			return nil, 0, &res.ErrorRes{
				Err:        errors.New("repository must be a mongo id"),
				StatusCode: http.StatusBadRequest,
			}
		}
		appendFirst = bson.D{{
			Key: "$match",
			Value: bson.M{
				"repository": idObjRepository,
			},
		}}
	} else if search != "" {
		appendFirst = bson.D{{
			Key: "$match",
			Value: bson.M{
				"title": bson.M{
					"$regex":   search,
					"$options": "i",
				},
			},
		}}
	}
	// Pipeline
	pipeline := d.pipeline(struct {
		Page          int
		Limit         int
		IDObjUser     primitive.ObjectID
		IsDiscussions bool
		AppendFirst   primitive.D
	}{
		Page:          page,
		Limit:         15,
		IDObjUser:     idObjUser,
		IsDiscussions: true,
		AppendFirst:   appendFirst,
	})
	// Get discussions
	var discussions []*models.DiscussionRes

	cursor, err := discussionModel.Use().Aggregate(db.Ctx, pipeline)
	if err != nil {
		return nil, 0, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	if err := cursor.All(db.Ctx, &discussions); err != nil {
		return nil, 0, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	// Filter by access
	if idRepository != "" {
		var locker sync.RWMutex
		errRes := utils.Concurrency(5, len(discussions), func(index int, ctx *utils.Context) {
			discussion := discussions[index]
			// Has access
			hasAccess, err := repoService.HasRepoAccess(
				discussion.Repository,
				idUser,
			)
			if err != nil {
				*ctx.Ctx = context.WithValue(*ctx.Ctx, ctx.Key, &res.ErrorRes{
					Err:        err.Err,
					StatusCode: err.StatusCode,
				})
				ctx.Cancel()
				return
			}
			if !hasAccess {
				locker.Lock()
				discussions[index] = nil
				locker.Unlock()
			}
		})
		if errRes != nil {
			return nil, 0, errRes
		}

		return utils.Filter(discussions, func(x *models.DiscussionRes) bool {
			return x != nil
		}), 0, nil
	}
	// Get total count data
	var total int64

	if getTotal {
		total, err = discussionModel.Use().CountDocuments(db.Ctx, bson.D{})
		if err != nil {
			return nil, 0, &res.ErrorRes{
				Err:        err,
				StatusCode: http.StatusServiceUnavailable,
			}
		}
	}

	return discussions, total, nil
}

func (*DiscussionService) HasAccess(
	idDiscussion,
	idUser string,
) *res.ErrorRes {
	// IDObjects
	idObjDiscussion, err := primitive.ObjectIDFromHex(idDiscussion)
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	// Discussion
	var discussion *models.Discussion

	opts := options.FindOne().SetProjection(bson.D{{
		Key:   "repository",
		Value: 1,
	}})
	cursor := discussionModel.Use().FindOne(db.Ctx, bson.D{{
		Key:   "_id",
		Value: idObjDiscussion,
	}}, opts)
	if err := cursor.Decode(&discussion); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return &res.ErrorRes{
				Err:        errors.New("discussion not found"),
				StatusCode: http.StatusNotFound,
			}
		}
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	// Has access
	if discussion.Repository == primitive.NilObjectID {
		return nil
	}

	hasAccess, errRes := repoService.HasRepoAccess(discussion.Repository, idUser)
	if errRes != nil {
		return errRes
	}
	if !hasAccess {
		return &res.ErrorRes{
			Err:        errors.New("unauthorized"),
			StatusCode: http.StatusUnauthorized,
		}
	}
	return nil
}

func (d *DiscussionService) GetImage(
	idDiscussion,
	image string,
	w io.Writer,
) *res.ErrorRes {
	// ObjectId
	idObjDiscussion, err := primitive.ObjectIDFromHex(idDiscussion)
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}

	existsDis, err := discussionModel.Exists(bson.D{
		{
			Key:   "_id",
			Value: idObjDiscussion,
		},
		{
			Key:   "image",
			Value: image,
		},
	})
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	if !existsDis {
		return &res.ErrorRes{
			Err:        errors.New("no existe el repositorio o la imagen"),
			StatusCode: http.StatusNotFound,
		}
	}
	// Read Image
	fileImage, err := utils.GetFile(image)
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusInternalServerError,
		}
	}
	w.Write(fileImage)

	return nil
}

func (d *DiscussionService) UploadDiscussion(
	discussion *forms.DiscussionForm,
	idUser string,
	image *multipart.FileHeader,
) *res.ErrorRes {
	idObjUser, err := primitive.ObjectIDFromHex(idUser)
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	if discussion.Repository != "" {
		idObjRepository, _ := primitive.ObjectIDFromHex(discussion.Repository)
		isOwner, err := repoService.IsRepoOwner(idObjUser, idObjRepository)
		if err != nil {
			return &res.ErrorRes{
				Err:        err,
				StatusCode: http.StatusServiceUnavailable,
			}
		}
		if !isOwner {
			return &res.ErrorRes{
				Err:        errors.New("no eres dueño del repositorio o no existe"),
				StatusCode: http.StatusUnauthorized,
			}
		}
	}
	// Model
	modelDis, err := discussionModel.NewModel(discussion, idObjUser, image)
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusInternalServerError,
		}
	}
	_, err = discussionModel.Use().InsertOne(db.Ctx, modelDis)
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}

	return nil
}

func (d *DiscussionService) React(
	idDiscussion,
	idUser,
	reaction string,
) *res.ErrorRes {
	// idObjects
	idObjDiscussion, err := primitive.ObjectIDFromHex(idDiscussion)
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	idObjUser, err := primitive.ObjectIDFromHex(idUser)
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	// Exists discussion
	exists, err := discussionModel.Exists(bson.D{{
		Key:   "_id",
		Value: idObjDiscussion,
	}})
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	if !exists {
		return &res.ErrorRes{
			Err:        errors.New("discussion not found"),
			StatusCode: http.StatusNotFound,
		}
	}
	// React
	filter := bson.D{
		{Key: "user", Value: idObjUser},
		{Key: "discussion", Value: idObjDiscussion},
	}

	existsReact, err := reactionModel.Exists(filter)
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	if existsReact {
		_, err = reactionModel.Use().UpdateOne(
			db.Ctx,
			filter,
			bson.D{{
				Key: "$set",
				Value: bson.M{
					"reaction": reaction,
				},
			}},
		)
		if err != nil {
			return &res.ErrorRes{
				Err:        err,
				StatusCode: http.StatusServiceUnavailable,
			}
		}
	} else {
		discussionModel := reactionModel.NewReactionModel(
			idObjUser,
			idObjDiscussion,
			reaction,
		)
		_, err = reactionModel.Use().InsertOne(db.Ctx, discussionModel)
		if err != nil {
			return &res.ErrorRes{
				Err:        err,
				StatusCode: http.StatusServiceUnavailable,
			}
		}
	}

	return nil
}

func (d *DiscussionService) DeleteReaction(idDiscussion, idUser string) *res.ErrorRes {
	// Objects
	idObjDiscussion, err := primitive.ObjectIDFromHex(idDiscussion)
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	idObjUser, err := primitive.ObjectIDFromHex(idUser)
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	// Delete reaction
	_, err = reactionModel.Use().DeleteOne(db.Ctx, bson.D{
		{Key: "user", Value: idObjUser},
		{Key: "discussion", Value: idObjDiscussion},
	})
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}

	return nil
}

func NewDiscussionService() *DiscussionService {
	return &DiscussionService{}
}
