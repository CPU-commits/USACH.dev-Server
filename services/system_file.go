package services

import (
	"archive/zip"
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"

	"github.com/CPU-commits/USACH.dev-Server/db"
	"github.com/CPU-commits/USACH.dev-Server/forms"
	"github.com/CPU-commits/USACH.dev-Server/models"
	"github.com/CPU-commits/USACH.dev-Server/res"
	"github.com/CPU-commits/USACH.dev-Server/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type SystemFileService struct{}

func (s *SystemFileService) FileHasRef(nameFile string) (bool, error) {
	hasRef, err := systemFileModel.Exists(bson.D{{
		Key:   "content",
		Value: nameFile,
	}})
	if err != nil {
		return false, err
	}
	if hasRef {
		return true, nil
	}

	hasRef, err = discussionModel.Exists(bson.D{{
		Key:   "image",
		Value: nameFile,
	}})
	if err != nil {
		return false, err
	}

	return hasRef, nil
}

func (s *SystemFileService) GetFolder(
	username,
	repoName,
	idFolder,
	idUserREQ string,
) (*models.SystemFileRes, map[string]interface{}, *res.ErrorRes) {
	repository, _, errRes := repoService.GetRepository(username, repoName, idUserREQ)
	if errRes != nil {
		return nil, nil, errRes
	}

	// Folder in repository
	folder, errRes := s.GetElementById(idFolder)
	if errRes != nil {
		return nil, nil, errRes
	}

	inRepo, err := s.isElementInRepo(folder, repository.ID)
	if err != nil {
		return nil, nil, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	if !inRepo {
		return nil, nil, &res.ErrorRes{
			Err:        errors.New("no existe la carpeta"),
			StatusCode: http.StatusNotFound,
		}
	}
	// Get rest of elements
	var folderChildrens []*models.SystemFileRes

	idObjFolder, _ := primitive.ObjectIDFromHex(idFolder)

	cursor, err := systemFileModel.Use().Aggregate(db.Ctx, mongo.Pipeline{
		bson.D{{
			Key: "$match",
			Value: bson.M{
				"_id": idObjFolder,
			},
		}},
		bson.D{{
			Key: "$lookup",
			Value: bson.M{
				"from":         models.SYSTEM_FILE_COLLECTION,
				"localField":   "childrens",
				"foreignField": "_id",
				"as":           "childrens",
			},
		}},
	})
	if err != nil {
		return nil, nil, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	if err := cursor.All(db.Ctx, &folderChildrens); err != nil {
		return nil, nil, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	// Build new repo
	repoMap := map[string]interface{}{
		"name":  repository.Name,
		"_id":   repository.ID,
		"owner": repository.Owner,
	}

	return folderChildrens[0], repoMap, nil
}

func (s *SystemFileService) GetElementById(idElement string) (*models.SystemFile, *res.ErrorRes) {
	idElementObj, err := primitive.ObjectIDFromHex(idElement)
	if err != nil {
		return nil, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	// Get element
	var element *models.SystemFile

	cursor := systemFileModel.Use().FindOne(db.Ctx, bson.D{{
		Key:   "_id",
		Value: idElementObj,
	}})
	if err := cursor.Decode(&element); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, &res.ErrorRes{
				Err:        err,
				StatusCode: http.StatusNotFound,
			}
		}
		return nil, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}

	return element, nil
}

func (s *SystemFileService) DownloadRepo(
	idRepo string,
	zipWritter *zip.Writer,
) *res.ErrorRes {
	idRepoObj, err := primitive.ObjectIDFromHex(idRepo)
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	repository, err := repoService.GetRepositoryById(idRepoObj)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return &res.ErrorRes{
				Err:        errors.New("el repositorio no existe"),
				StatusCode: http.StatusNotFound,
			}
		}
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	for _, child := range repository.SystemFile {
		errRes := s.DownloadChild(child.Hex(), zipWritter)
		if errRes != nil {
			return errRes
		}
	}
	return nil
}

func (s *SystemFileService) DownloadChild(
	idChild string,
	zipWritter *zip.Writer,
) *res.ErrorRes {
	element, errRes := s.GetElementById(idChild)
	if errRes != nil {
		return errRes
	}
	if !element.IsDirectory {
		file, err := utils.GetFile(element.Content)
		if err != nil {
			return &res.ErrorRes{
				Err:        err,
				StatusCode: http.StatusInternalServerError,
			}
		}
		zipFile, err := zipWritter.Create(element.Name)
		if err != nil {
			return &res.ErrorRes{
				Err:        err,
				StatusCode: http.StatusInternalServerError,
			}
		}
		_, err = zipFile.Write(file)
		if err != nil {
			return &res.ErrorRes{
				Err:        err,
				StatusCode: http.StatusInternalServerError,
			}
		}
	} else if element.IsDirectory && len(element.Childrens) > 0 {
		// Create new zip
		buffer := bytes.NewBuffer(nil)
		newZipWritter := zip.NewWriter(buffer)

		for _, child := range element.Childrens {
			errRes = s.DownloadChild(child.Hex(), newZipWritter)
			if errRes != nil {
				return errRes
			}
		}
		zipFile, err := zipWritter.Create(element.Name + ".zip")
		if err != nil {
			return &res.ErrorRes{
				Err:        err,
				StatusCode: http.StatusInternalServerError,
			}
		}
		// Close zip
		newZipWritter.Close()

		_, err = zipFile.Write(buffer.Bytes())
		if err != nil {
			return &res.ErrorRes{
				Err:        err,
				StatusCode: http.StatusInternalServerError,
			}
		}
	}

	return nil
}

func (s *SystemFileService) getFolderChildrens(
	folder *models.SystemFile,
) (childrens []primitive.ObjectID, err error) {
	if folder.Childrens == nil || err != nil {
		return nil, nil
	}

	for _, children := range folder.Childrens {
		childrens = append(childrens, children)
		element, errRes := s.GetElementById(children.Hex())
		if errRes != nil {
			return nil, err
		}
		if element.IsDirectory {
			childrensToAppend, err := s.getFolderChildrens(element)
			if err != nil {
				return nil, err
			}
			childrens = append(childrens, childrensToAppend...)
		}
	}

	return
}

func (s *SystemFileService) isElementInRepo(
	element *models.SystemFile,
	idRepository primitive.ObjectID,
) (bool, error) {
	hasParent, err := systemFileModel.Exists(bson.D{{
		Key: "childrens",
		Value: bson.M{
			"$in": bson.A{element.ID},
		},
	}})
	if err != nil {
		return false, err
	}
	if hasParent {
		var parentElement *models.SystemFile

		cursor := systemFileModel.Use().FindOne(db.Ctx, bson.D{{
			Key: "childrens",
			Value: bson.M{
				"$in": bson.A{element.ID},
			},
		}})
		if err := cursor.Decode(&parentElement); err != nil {
			return false, err
		}

		return s.isElementInRepo(parentElement, idRepository)
	}
	repository, err := repoService.GetRepositoryById(idRepository)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, err
	}
	isElementInSystemFile, err := utils.AnyMatch(
		repository.SystemFile,
		func(x interface{}) bool {
			elementRepo := x.(primitive.ObjectID)
			return elementRepo == element.ID
		},
	)
	if err != nil {
		return false, err
	}
	return isElementInSystemFile, nil
}

func (s *SystemFileService) NewRepoElement(
	element *forms.SystemFileForm,
	repository,
	parent,
	idUser string,
	file *multipart.FileHeader,
) (map[string]interface{}, *res.ErrorRes) {
	// Check if repo exists
	idRepositoryObj, err := primitive.ObjectIDFromHex(repository)
	if err != nil {
		return nil, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}

	existsRepo, err := repoModel.Exists(bson.D{{
		Key:   "_id",
		Value: idRepositoryObj,
	}})
	if err != nil {
		return nil, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	if !existsRepo {
		return nil, &res.ErrorRes{
			Err:        errors.New("el repositorio no existe"),
			StatusCode: http.StatusNotFound,
		}
	}
	// Check if is owner
	idUserObj, err := primitive.ObjectIDFromHex(idUser)
	if err != nil {
		return nil, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	isOwner, err := repoService.IsRepoOwner(idUserObj, idRepositoryObj)
	if err != nil {
		return nil, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	if !isOwner {
		return nil, &res.ErrorRes{
			Err:        errors.New("no eres dueño del repositorio"),
			StatusCode: http.StatusUnauthorized,
		}
	}
	// Insert into repo
	response := make(map[string]interface{})

	newElementModel, err := systemFileModel.NewModel(element, file)
	if err != nil {
		return nil, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusInternalServerError,
		}
	}
	insertedSF, err := systemFileModel.Use().InsertOne(db.Ctx, newElementModel)
	if err != nil {
		return nil, &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	if parent == "" {
		_, err := repoModel.Use().UpdateByID(db.Ctx, idRepositoryObj, bson.D{{
			Key: "$addToSet",
			Value: bson.M{
				"system_file": insertedSF.InsertedID,
			},
		}})
		if err != nil {
			return nil, &res.ErrorRes{
				Err:        err,
				StatusCode: http.StatusServiceUnavailable,
			}
		}
	} else {
		element, errRes := s.GetElementById(parent)
		if errRes != nil {
			return nil, errRes
		}
		if !element.IsDirectory {
			return nil, &res.ErrorRes{
				Err:        errors.New("parent no es una carpeta"),
				StatusCode: http.StatusBadRequest,
			}
		}
		inRepo, err := s.isElementInRepo(element, idRepositoryObj)
		if err != nil {
			return nil, &res.ErrorRes{
				Err:        err,
				StatusCode: http.StatusServiceUnavailable,
			}
		}
		if !inRepo {
			return nil, &res.ErrorRes{
				Err:        errors.New("el elemento no pertenece al repositorio"),
				StatusCode: http.StatusUnauthorized,
			}
		}

		_, err = systemFileModel.Use().UpdateByID(db.Ctx, element.ID, bson.D{{
			Key: "$addToSet",
			Value: bson.M{
				"childrens": insertedSF.InsertedID,
			},
		}})
		if err != nil {
			return nil, &res.ErrorRes{
				Err:        err,
				StatusCode: http.StatusServiceUnavailable,
			}
		}
	}
	// Make response
	response["_id"] = insertedSF.InsertedID.(primitive.ObjectID).Hex()

	return response, nil
}

func (s *SystemFileService) DeleteElement(
	idRepository,
	idElement,
	idUser string,
) *res.ErrorRes {
	// IdObjects
	idRepositoryObj, err := primitive.ObjectIDFromHex(idRepository)
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	idUserObj, err := primitive.ObjectIDFromHex(idUser)
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusBadRequest,
		}
	}
	// Check owner
	isOwner, err := repoService.IsRepoOwner(idUserObj, idRepositoryObj)
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	if !isOwner {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusUnauthorized,
		}
	}
	// Get element by id
	element, errRes := s.GetElementById(idElement)
	if errRes != nil {
		return errRes
	}
	// Check element in repo
	isElementInRepo, err := s.isElementInRepo(element, idRepositoryObj)
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	if !isElementInRepo {
		return &res.ErrorRes{
			Err:        errors.New("el elemento no pertenece al repositorio"),
			StatusCode: http.StatusUnauthorized,
		}
	}
	// Get elements to delete
	var toDelete []primitive.ObjectID
	toDelete = append(toDelete, element.ID)

	if element.IsDirectory {
		childrens, err := s.getFolderChildrens(element)
		if err != nil {
			return nil
		}
		toDelete = append(toDelete, childrens...)
	}
	// Delete elements
	_, err = systemFileModel.Use().DeleteMany(db.Ctx, bson.D{{
		Key: "_id",
		Value: bson.M{
			"$in": toDelete,
		},
	}})
	if err != nil {
		return &res.ErrorRes{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
		}
	}
	_, err = repoModel.Use().UpdateByID(db.Ctx, idRepositoryObj, bson.D{{
		Key: "$pull",
		Value: bson.M{
			"system_file": element.ID,
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

func NewSystemFileService() *SystemFileService {
	return &SystemFileService{}
}
