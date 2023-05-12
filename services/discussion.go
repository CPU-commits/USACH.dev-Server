package services

import (
	"errors"
	"mime/multipart"
	"net/http"

	"github.com/CPU-commits/USACH.dev-Server/db"
	"github.com/CPU-commits/USACH.dev-Server/forms"
	"github.com/CPU-commits/USACH.dev-Server/res"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DiscussionService struct{}

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

func NewDiscussionService() *DiscussionService {
	return &DiscussionService{}
}
