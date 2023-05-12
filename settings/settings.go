package settings

import (
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
)

var lock = &sync.Mutex{}
var singleSettingsInstace *settings

type settings struct {
	JWT_SECRET_KEY      string
	JWT_SECRET_REFRESH  string
	MONGO_DB            string
	MONGO_ROOT_USERNAME string
	MONGO_ROOT_PASSWORD string
	MONGO_HOST          string
	MONGO_CONNECTION    string
	AWS_BUCKET          string
	AWS_REGION          string
	CLIENT_URL          string
	SMTP_HOST           string
	SMTP_PORT           int
	SMTP_USER           string
	SMTP_PASSWORD       string
	GO_ENV              string
	MEDIA_FOLDER        string
	REDIS_URI           string
	REDIS_PASS          string
	REDIS_DB            int
}

func newSettings() *settings {
	smtpPort, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		panic("SMTP_PORT Must be a int")
	}
	redisDB, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		panic("REDIS_DB Must be a int")
	}
	return &settings{
		JWT_SECRET_KEY:      os.Getenv("JWT_SECRET_KEY"),
		MONGO_DB:            os.Getenv("MONGO_DB"),
		MONGO_ROOT_USERNAME: os.Getenv("MONGO_ROOT_USERNAME"),
		MONGO_ROOT_PASSWORD: os.Getenv("MONGO_ROOT_PASSWORD"),
		MONGO_HOST:          os.Getenv("MONGO_HOST"),
		MONGO_CONNECTION:    os.Getenv("MONGO_CONNECTION"),
		AWS_BUCKET:          os.Getenv("AWS_BUCKET"),
		AWS_REGION:          os.Getenv("AWS_REGION"),
		CLIENT_URL:          os.Getenv("CLIENT_URL"),
		SMTP_HOST:           os.Getenv("SMTP_HOST"),
		SMTP_PORT:           smtpPort,
		SMTP_USER:           os.Getenv("SMTP_USER"),
		SMTP_PASSWORD:       os.Getenv("SMTP_PASSWORD"),
		GO_ENV:              os.Getenv("GO_ENV"),
		MEDIA_FOLDER:        os.Getenv("MEDIA_FOLDER"),
		REDIS_URI:           os.Getenv("REDIS_URI"),
		REDIS_PASS:          os.Getenv("REDIS_PASS"),
		REDIS_DB:            redisDB,
	}
}

func init() {
	if os.Getenv("GO_ENV") != "prod" {
		if err := godotenv.Load(); err != nil {
			log.Println("No .env file found")
		}
	}
}

func GetSettings() *settings {
	if singleSettingsInstace == nil {
		lock.Lock()
		defer lock.Unlock()
		singleSettingsInstace = newSettings()
	}
	return singleSettingsInstace
}
