package jobs

import (
	"github.com/CPU-commits/USACH.dev-Server/models"
	"github.com/CPU-commits/USACH.dev-Server/services"
	"github.com/CPU-commits/USACH.dev-Server/settings"
)

// Models
var (
	usersModel      = models.NewUsersModel()
	usersTokenModel = models.NewUsersTokenModel()
	systemFileModel = models.NewSystemFileModel()
)

// Services
var (
	systemFileService = services.NewSystemFileService()
)

// Settings
var settingsData = settings.GetSettings()
