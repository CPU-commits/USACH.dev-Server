package services

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/CPU-commits/USACH.dev-Server/db"
	"github.com/CPU-commits/USACH.dev-Server/forms"
	"github.com/CPU-commits/USACH.dev-Server/res"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LinkService struct{}

func (l *LinkService) AddLink(
	idRepository,
	idUser string,
	link *forms.LinkForm,
) (primitive.ObjectID, *res.ErrorRes) {
	idObjUser, err := primitive.ObjectIDFromHex(idUser)
	if err != nil {
		return primitive.NilObjectID, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	idObjRepository, err := primitive.ObjectIDFromHex(idRepository)
	if err != nil {
		return primitive.NilObjectID, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	// Check repo owner
	isRepoOwner, err := repoService.IsRepoOwner(idObjUser, idObjRepository)
	if err != nil {
		return primitive.NilObjectID, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	if !isRepoOwner {
		return primitive.NilObjectID, &res.ErrorRes{
			Err:        errors.New("no eres dueño del repositorio o no existe"),
			StatusCode: http.StatusUnauthorized,
		}
	}
	// Create link
	newLink := repoModel.NewLinkModel(link)
	_, err = repoModel.Use().UpdateByID(db.Ctx, idObjRepository, bson.D{{
		Key: "$addToSet",
		Value: bson.M{
			"links": newLink,
		},
	}})
	if err != nil {
		return primitive.NilObjectID, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}

	return newLink.ID, nil
}

func (l *LinkService) DeleteLink(
	idRepository,
	idLink,
	idUser string,
) *res.ErrorRes {
	// ID Objects
	idObjRepository, err := primitive.ObjectIDFromHex(idRepository)
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	idObjLink, err := primitive.ObjectIDFromHex(idLink)
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
	// Repo owner
	exists, err := repoService.IsRepoOwner(idObjUser, idObjRepository)
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	if !exists {
		return &res.ErrorRes{
			Err:        errors.New("el repositorio no te pertenece o no existe"),
			StatusCode: http.StatusUnauthorized,
		}
	}
	// Delete link
	fmt.Printf("idObjLink: %v\n", idObjLink)
	fmt.Printf("idObjRepository: %v\n", idObjRepository)
	_, err = repoModel.Use().UpdateByID(db.Ctx, idObjRepository, bson.D{{
		Key: "$pull",
		Value: bson.M{
			"links": bson.M{
				"_id": idObjLink,
			},
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

func NewLinkService() *LinkService {
	return &LinkService{}
}
