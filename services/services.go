package services

import (
	"github.com/CPU-commits/USACH.dev-Server/models"
	"github.com/CPU-commits/USACH.dev-Server/settings"
	"github.com/CPU-commits/USACH.dev-Server/stack"
)

// Models
var (
	userModel       = models.NewUsersModel()
	usersTokenModel = models.NewUsersTokenModel()
	repoModel       = models.NewRepositoryModel()
	systemFileModel = models.NewSystemFileModel()
	likesModel      = models.NewLikesModel()
	discussionModel = models.NewDiscussionModel()
)

// Services
var (
	userService       = NewUserService()
	repoService       = NewRepositoryService()
	profileService    = NewProfileService()
	systemFileService = NewSystemFileService()
	likeService       = NewLikeService()
)

// Settings
var settingsData = settings.GetSettings()

// Stack
var mem = stack.NewStack()
