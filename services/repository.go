package services

import (
	"archive/zip"
	"context"
	"errors"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/CPU-commits/USACH.dev-Server/db"
	"github.com/CPU-commits/USACH.dev-Server/forms"
	"github.com/CPU-commits/USACH.dev-Server/models"
	"github.com/CPU-commits/USACH.dev-Server/res"
	"github.com/CPU-commits/USACH.dev-Server/utils"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/sync/semaphore"
)

type RepositoryService struct{}

func (r *RepositoryService) IsRepoOwner(idUser, idRepo primitive.ObjectID) (bool, error) {
	isOwner, err := repoModel.Exists(bson.D{
		{
			Key:   "owner",
			Value: idUser,
		},
		{
			Key:   "_id",
			Value: idRepo,
		},
	})
	if err != nil {
		return false, err
	}

	return isOwner, nil
}

func (r *RepositoryService) GetRepoAccess(
	idRepository primitive.ObjectID,
) (map[string]interface{}, *res.ErrorRes) {
	opts := options.FindOne().SetProjection(bson.D{
		{
			Key:   "access",
			Value: 1,
		},
		{
			Key:   "custom_access",
			Value: 1,
		},
	})

	repository, err := r.GetRepositoryById(idRepository, opts)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, &res.ErrorRes{
				Err:        errors.New("no existe el repositorio"),
				StatusCode: http.StatusNotFound,
			}
		}
		return nil, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	// Response
	response := map[string]interface{}{
		"access":        repository.Access,
		"custom_access": repository.CustomAccess,
	}

	return response, nil
}

func (r *RepositoryService) HasRepoAccess(
	idObjRepository primitive.ObjectID,
	idUser string,
) (bool, *res.ErrorRes) {
	// Repo access
	repoAccess, errRes := r.GetRepoAccess(idObjRepository)
	if errRes != nil {
		return false, errRes
	}
	if repoAccess["access"] == "private-group" {
		idObjUser, err := primitive.ObjectIDFromHex(idUser)
		if err != nil {
			return false, nil
		}

		hasAccess, err := utils.AnyMatch(
			repoAccess["custom_access"],
			func(x interface{}) bool {
				return x.(primitive.ObjectID) != idObjUser
			},
		)
		if err != nil {
			return false, &res.ErrorRes{
				Err:        err,
				StatusCode: http.StatusInternalServerError,
			}
		}
		return hasAccess, nil
	}
	return true, nil
}

func (r *RepositoryService) GetRepositoryId(
	username,
	repoName string,
) (*primitive.ObjectID, *res.ErrorRes) {
	var repository *models.Repository

	// Get username
	user, errRes := userService.GetByUsername(username, false)
	if errRes != nil {
		return nil, errRes
	}

	cursor := repoModel.Use().FindOne(db.Ctx, bson.D{
		{
			Key:   "owner",
			Value: user.ID,
		},
		{
			Key:   "name",
			Value: repoName,
		},
	})
	if err := cursor.Decode(&repository); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, &res.ErrorRes{
				Err:        errors.New("no existe el repositorio"),
				StatusCode: http.StatusNotFound,
			}
		}
		return nil, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}

	return &repository.ID, nil
}

func (r *RepositoryService) GetRepositoryById(
	idRepository primitive.ObjectID,
	opts ...*options.FindOneOptions,
) (*models.Repository, error) {
	var repository *models.Repository

	cursor := repoModel.Use().FindOne(db.Ctx, bson.D{{
		Key:   "_id",
		Value: idRepository,
	}}, opts...)
	if err := cursor.Decode(&repository); err != nil {
		return nil, err
	}

	return repository, nil
}

func (r *RepositoryService) addView(idUser string, idRepository primitive.ObjectID) error {
	_, err := mem.Get(REPOSITORY_VIEW + idUser)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			err := mem.Set(REPOSITORY_VIEW+idUser, "1", time.Hour*5)
			if err != nil {
				return err
			}
			_, err = repoModel.Use().UpdateByID(db.Ctx, idRepository, bson.D{{
				Key: "$inc",
				Value: bson.M{
					"views": 1,
				},
			}})
			if err != nil {
				return err
			}
		}
		return err
	}

	return nil
}

func (r *RepositoryService) GetRepository(
	username,
	repositoryName,
	idUserREQ string,
) (*models.RepositoryRes, *models.Like, *res.ErrorRes) {
	idObjRepository, errRes := r.GetRepositoryId(username, repositoryName)
	if errRes != nil {
		return nil, nil, errRes
	}
	// Get repo
	var repository []*models.RepositoryRes

	filterRepo := bson.D{{
		Key:   "_id",
		Value: idObjRepository,
	}}
	opts := options.Aggregate().SetCollation(&options.Collation{
		Locale: "es",
	})
	cursor, err := repoModel.Use().Aggregate(db.Ctx, mongo.Pipeline{
		bson.D{{Key: "$match", Value: filterRepo}},
		bson.D{{
			Key: "$lookup",
			Value: bson.M{
				"from":         models.SYSTEM_FILE_COLLECTION,
				"localField":   "system_file",
				"foreignField": "_id",
				"as":           "system_file",
				"pipeline": bson.A{
					bson.M{
						"$sort": bson.M{
							"name": 1,
						},
					},
					bson.M{
						"$sort": bson.M{
							"is_directory": -1,
						},
					},
				},
			}}},
		bson.D{{
			Key: "$lookup",
			Value: bson.M{
				"from":         models.USERS_COLLECTION,
				"localField":   "owner",
				"foreignField": "_id",
				"as":           "owner",
				"pipeline": bson.A{bson.M{
					"$project": bson.M{
						"username":  1,
						"full_name": 1,
					},
				}},
			},
		}},
		bson.D{{Key: "$unwind", Value: bson.M{"path": "$owner"}}},
	}, opts)
	if err != nil {
		return nil, nil, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	if err := cursor.All(db.Ctx, &repository); err != nil {
		return nil, nil, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	// Get like
	var like *models.Like

	idObjUser, err := primitive.ObjectIDFromHex(idUserREQ)
	if err == nil {
		like, errRes = likeService.GetLike(idObjUser, repository[0].ID)
		if errRes != nil {
			return nil, nil, errRes
		}
	}
	// Add view
	go r.addView(idUserREQ, repository[0].ID)

	return repository[0], like, nil
}

func (r *RepositoryService) GetRepositories(
	idUser,
	search string,
	page int,
	total bool,
) ([]models.RepositoryRes, int64, *res.ErrorRes) {
	idObjUser, err := primitive.ObjectIDFromHex(idUser)
	if err != nil && idUser != "" {
		return nil, 0, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	// Get repositories
	var repositories []models.RepositoryRes
	// Filter
	orFilter := bson.A{
		bson.M{"access": "public"},
	}
	if idUser != "" {
		orFilter = append(orFilter, bson.M{
			"access": "private-group",
			"custom_access": bson.M{
				"$in": bson.A{idObjUser},
			},
		})
	}

	andFilter := bson.A{bson.D{{
		Key:   "$or",
		Value: orFilter,
	}}}
	if search != "" {
		andFilter = append(andFilter, bson.D{{
			Key: "$or",
			Value: bson.A{
				bson.M{"name": bson.M{"$regex": search, "$options": "i"}},
				bson.M{"tags": bson.M{"$regex": search, "$options": "i"}},
			},
		}})
	}
	filter := bson.D{{
		Key: "$match",
		Value: bson.M{
			"$and": andFilter,
		},
	}}
	// Pipeline
	cursor, err := repoModel.Use().Aggregate(db.Ctx, mongo.Pipeline{
		filter,
		bson.D{{
			Key: "$project",
			Value: bson.D{
				{
					Key:   "system_file",
					Value: 0,
				},
				{
					Key:   "description",
					Value: 0,
				},
				{
					Key:   "custom_access",
					Value: 0,
				},
				{
					Key:   "downloads",
					Value: 0,
				},
				{
					Key:   "content",
					Value: 0,
				},
				{
					Key:   "links",
					Value: 0,
				},
			},
		}},
		bson.D{{
			Key: "$sort",
			Value: bson.D{
				{
					Key:   "updated_date",
					Value: -1,
				},
				{
					Key:   "stars",
					Value: -1,
				},
			},
		}},
		bson.D{{
			Key:   "$skip",
			Value: int64(page * 20),
		}},
		bson.D{{
			Key:   "$limit",
			Value: 20,
		}},
		bson.D{{
			Key: "$lookup",
			Value: bson.M{
				"from":         models.USERS_COLLECTION,
				"localField":   "owner",
				"foreignField": "_id",
				"as":           "owner",
				"pipeline": bson.A{bson.M{
					"$project": bson.M{"username": 1, "full_name": 1},
				}},
			},
		}},
		bson.D{{
			Key: "$unwind",
			Value: bson.M{
				"path": "$owner",
			},
		}},
	})
	if err != nil {
		return nil, 0, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	if err := cursor.All(db.Ctx, &repositories); err != nil {
		return nil, 0, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	// Total
	var totalElements int64
	if total {
		totalElements, err = repoModel.Use().CountDocuments(db.Ctx, bson.D{{
			Key:   "$or",
			Value: orFilter,
		}})
		if err != nil {
			return nil, 0, &res.ErrorRes{
				Err:        err,
				StatusCode: http.StatusServiceUnavailable,
			}
		}
	}

	return repositories, totalElements, nil
}

func (r *RepositoryService) GetUserRepositories(
	username,
	idUser string,
) ([]*models.Repository, *res.ErrorRes) {
	idObjUser, err := primitive.ObjectIDFromHex(idUser)
	if err != nil && idUser != "" {
		return nil, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	// Get repos by access
	var repositories []*models.Repository

	opts := options.Find().SetProjection(bson.D{
		{
			Key:   "system_file",
			Value: 0,
		},
		{
			Key:   "description",
			Value: 0,
		},
		{
			Key:   "custom_access",
			Value: 0,
		},
		{
			Key:   "downloads",
			Value: 0,
		},
		{
			Key:   "content",
			Value: 0,
		},
		{
			Key:   "links",
			Value: 0,
		},
	}).SetSort(bson.D{{
		Key:   "updated_date",
		Value: -1,
	}})
	// Owner
	var isUserOwner bool
	if idUser != "" {
		isOwner, errRes := userService.isOwner(username, idObjUser)
		if errRes != nil {
			return nil, errRes
		}
		isUserOwner = isOwner
	}
	// Get user
	user, errRes := userService.GetByUsername(username, false)
	if errRes != nil {
		return nil, errRes
	}
	// Filter
	filter := bson.D{{
		Key:   "owner",
		Value: user.ID,
	}}
	if !isUserOwner {
		filter = append(filter, bson.E{
			Key: "$or",
			Value: bson.A{
				bson.M{"access": "public"},
				bson.M{
					"access": "private-group",
					"custom_access": bson.M{
						"$in": bson.A{idObjUser},
					},
				},
			},
		})
	}
	cursor, err := repoModel.Use().Find(db.Ctx, filter, opts)
	if err != nil {
		return nil, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	if err := cursor.All(db.Ctx, &repositories); err != nil {
		return nil, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}

	return repositories, nil
}

func (r *RepositoryService) DownloadRepository(
	repository,
	child string,
	w io.Writer,
) (string, *res.ErrorRes) {
	zipWritter := zip.NewWriter(w)
	defer zipWritter.Close()

	if child != "" {
		return "", systemFileService.DownloadChild(child, zipWritter)
	}
	// Update downloads
	idObjRepository, err := primitive.ObjectIDFromHex(repository)
	if err != nil {
		return "", &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	_, err = repoModel.Use().UpdateByID(db.Ctx, idObjRepository, bson.D{{
		Key: "$inc",
		Value: bson.M{
			"downloads": 1,
		},
	}})
	if err != nil {
		return "", &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}

	return "", systemFileService.DownloadRepo(repository, zipWritter)
}

func (r *RepositoryService) ExistsRepoUser(
	idUser primitive.ObjectID,
	repoName string,
) (bool, *res.ErrorRes) {
	var repo *models.Repository

	opts := options.FindOne().SetProjection(bson.D{{
		Key:   "_id",
		Value: 1,
	}})
	pointer := repoModel.Use().FindOne(db.Ctx, bson.D{
		{
			Key:   "owner",
			Value: idUser,
		},
		{
			Key:   "name",
			Value: repoName,
		},
	}, opts)
	if err := pointer.Decode(&repo); err != nil {
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

func (r *RepositoryService) UploadRepository(
	repository *forms.RepositoryForm,
	idUser string,
) *res.ErrorRes {
	idObjUser, err := primitive.ObjectIDFromHex(idUser)
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	// Check if exists
	existsRepo, errRes := r.ExistsRepoUser(idObjUser, repository.Name)
	if errRes != nil {
		return errRes
	}
	if existsRepo {
		return &res.ErrorRes{
			Err:        errors.New("ya existe este repositorio a tu nombre"),
			StatusCode: http.StatusConflict,
		}
	}

	modelRepository := repoModel.NewModel(repository, idObjUser)
	// Insert
	_, err = repoModel.Use().InsertOne(db.Ctx, modelRepository)
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}

	return nil
}

func (r *RepositoryService) UpdateRepository(
	idRepository,
	idUser string,
	repositoryForm *forms.UpdateRepositoryForm,
) *res.ErrorRes {
	idObjRepository, err := primitive.ObjectIDFromHex(idRepository)
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
	// Exists repository and has access
	isRepoOwner, err := r.IsRepoOwner(idObjUser, idObjRepository)
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	if !isRepoOwner {
		return &res.ErrorRes{
			Err:        errors.New("no eres dueño del repositorio o no existe"),
			StatusCode: http.StatusUnauthorized,
		}
	}
	// Update repo
	var update bson.D
	update = append(update, bson.E{
		Key:   "updated_date",
		Value: primitive.NewDateTimeFromTime(time.Now()),
	})

	if repositoryForm.Content != "" {
		update = append(update, bson.E{
			Key:   "content",
			Value: repositoryForm.Content,
		})
	}
	if repositoryForm.Description != "" {
		update = append(update, bson.E{
			Key:   "description",
			Value: repositoryForm.Description,
		})
	}
	if repositoryForm.Access != "" {
		update = append(update, bson.E{
			Key:   "access",
			Value: repositoryForm.Access,
		})
		if repositoryForm.Access != "private-group" {
			update = append(update, bson.E{
				Key:   "custom_access",
				Value: bson.A{},
			})
		}
	}
	if repositoryForm.Tags != nil {
		update = append(update, bson.E{
			Key:   "tags",
			Value: repositoryForm.Tags,
		})
	}
	if repositoryForm.CustomAccess != nil && repositoryForm.Access == "private-group" {
		var customAccess []primitive.ObjectID
		// Check if exists all users
		var lock sync.RWMutex
		var wg sync.WaitGroup
		sem := semaphore.NewWeighted(int64(10))
		// Ctx with cancel if error
		ctx, cancel := context.WithCancel(context.Background())
		// Ctx error
		errKey := "error"
		ctx = context.WithValue(ctx, errKey, nil)

		for _, username := range repositoryForm.CustomAccess {
			wg.Add(1)
			if err := sem.Acquire(ctx, 1); err != nil {
				wg.Done()
				// Close go routines
				cancel()
				if errors.Is(err, context.Canceled) {
					if errRes := ctx.Value(errKey); errRes != nil {
						return errRes.(*res.ErrorRes)
					}
				}
				return &res.ErrorRes{
					Err:        err,
					StatusCode: http.StatusBadRequest,
				}
			}
			go func(
				username string,
				wg *sync.WaitGroup,
				lock *sync.RWMutex,
			) {
				defer wg.Done()

				user, errRes := userService.GetByUsername(username, false)
				if errRes != nil {
					ctx = context.WithValue(ctx, errKey, errRes)
					cancel()
					return
				}
				if user == nil {
					ctx = context.WithValue(ctx, errKey, &res.ErrorRes{
						Err:        errors.New("el usuario no existe"),
						StatusCode: http.StatusNotFound,
					})
					cancel()
					return
				}
				lock.Lock()
				customAccess = append(customAccess, user.ID)
				lock.Unlock()

				// Free semaphore
				sem.Release(1)
			}(username, &wg, &lock)
		}
		// Close all
		wg.Wait()
		cancel()
		// Catch error
		if err := ctx.Value(errKey); err != nil {
			return err.(*res.ErrorRes)
		}

		update = append(update, bson.E{
			Key:   "custom_access",
			Value: customAccess,
		})
	}

	_, err = repoModel.Use().UpdateByID(db.Ctx, idObjRepository, bson.D{{
		Key:   "$set",
		Value: update,
	}})
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}

	return nil
}

func (r *RepositoryService) ToggleLike(
	idUser,
	idRepository string,
	plus bool,
) *res.ErrorRes {
	idObjRepository, err := primitive.ObjectIDFromHex(idRepository)
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
	// Check if exists repo
	exists, err := repoModel.Exists(bson.D{{
		Key:   "_id",
		Value: idObjRepository,
	}})
	if !exists {
		return &res.ErrorRes{
			Err:        errors.New("no existe el repositorio"),
			StatusCode: http.StatusNotFound,
		}
	}
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	// Check if user like´s repo
	like, errRes := likeService.GetLike(idObjUser, idObjRepository)
	if errRes != nil {
		return errRes
	}
	// Like
	if like != nil && like.Plus != plus || like == nil {
		errRes = likeService.ToggleLike(
			like,
			idObjUser,
			idObjRepository,
			plus,
		)
		if errRes != nil {
			return errRes
		}
	}

	return nil
}

func (r *RepositoryService) DeleteRepository(
	idRepository,
	idUser string,
) *res.ErrorRes {
	idObjUser, err := primitive.ObjectIDFromHex(idUser)
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	idObjRepository, err := primitive.ObjectIDFromHex(idRepository)
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	// IsRepoOwner
	exists, err := r.IsRepoOwner(idObjUser, idObjRepository)
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	if !exists {
		return &res.ErrorRes{
			Err:        errors.New("no eres dueño del repositorio o no existe"),
			StatusCode: http.StatusUnauthorized,
		}
	}
	// Delete repository
	_, err = repoModel.Use().DeleteOne(db.Ctx, bson.D{{
		Key:   "_id",
		Value: idObjRepository,
	}})
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}

	return nil
}

func NewRepositoryService() *RepositoryService {
	return &RepositoryService{}
}
