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
	reactionModel   = models.NewReactionModel()
	commentModel    = models.NewCommentModel()
	profileModel    = models.NewProfileModel()
)

// Services
var (
	userService       = NewUserService()
	repoService       = NewRepositoryService()
	systemFileService = NewSystemFileService()
	likeService       = NewLikeService()
	discussionService = NewDiscussionService()
)

// Settings
var settingsData = settings.GetSettings()

// Stack
var mem = stack.NewStack()

// Tasks
var pubSubClient = stack.NewPubSubClient()
