package models

import (
	"github.com/CPU-commits/USACH.dev-Server/db"
	"github.com/CPU-commits/USACH.dev-Server/settings"
)

var settingsData = settings.GetSettings()

// Models
var (
	userModel = NewUsersModel()
)

// MongoDB
var DbConnect = db.NewConnection(
	settingsData.MONGO_HOST,
	settingsData.MONGO_DB,
)
