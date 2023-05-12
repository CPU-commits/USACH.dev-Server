package controllers

import (
	"github.com/CPU-commits/USACH.dev-Server/services"
	"github.com/CPU-commits/USACH.dev-Server/settings"
)

var (
	usersService      = services.NewUserService()
	authService       = services.NewAuthService()
	repoService       = services.NewRepositoryService()
	systemFileService = services.NewSystemFileService()
	linkService       = services.NewLinkService()
	discussionService = services.NewDiscussionService()
)

var settingsData = settings.GetSettings()
