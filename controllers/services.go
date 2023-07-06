package controllers

import (
	"github.com/CPU-commits/USACH.dev-Server/services"
	"github.com/CPU-commits/USACH.dev-Server/settings"
)

// Services
var (
	usersService      = services.NewUserService()
	authService       = services.NewAuthService()
	repoService       = services.NewRepositoryService()
	systemFileService = services.NewSystemFileService()
	linkService       = services.NewLinkService()
	discussionService = services.NewDiscussionService()
	commentService    = services.NewCommentService()
)

// Settings
var settingsData = settings.GetSettings()
