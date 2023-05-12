package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/CPU-commits/USACH.dev-Server/controllers"
	"github.com/CPU-commits/USACH.dev-Server/middlewares"
	"github.com/CPU-commits/USACH.dev-Server/res"
	"github.com/CPU-commits/USACH.dev-Server/settings"
	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/secure"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	"github.com/swaggo/swag/example/basic/docs"

	// swagger embed files
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// gin-swagger middleware

func keyFunc(c *gin.Context) string {
	return c.ClientIP()
}

func ErrorHandler(c *gin.Context, info ratelimit.Info) {
	c.JSON(http.StatusTooManyRequests, &res.Response{
		Message: "Too many requests. Try again in" + time.Until(info.ResetTime).String(),
	})
}

var settingsData = settings.GetSettings()

func Init() {
	router := gin.New()
	// Proxies
	router.SetTrustedProxies([]string{"localhost"})
	// Zap logger
	// Create folder if not exists
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		err := os.Mkdir("logs", os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
	// Log file
	logEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	fileCore := zapcore.NewCore(logEncoder, zapcore.AddSync(&lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     7,
	}), zap.InfoLevel)
	// Log console
	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig())
	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zap.InfoLevel)
	// Combine cores for multi-output logging
	teeCore := zapcore.NewTee(fileCore, consoleCore)
	zapLogger := zap.New(teeCore)

	router.Use(ginzap.GinzapWithConfig(zapLogger, &ginzap.Config{
		TimeFormat: time.RFC3339,
		UTC:        true,
		SkipPaths:  []string{"/api/v1/swagger"},
	}))
	router.Use(ginzap.RecoveryWithZap(zapLogger, true))

	router.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Server Internal Error: %s", err))
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, res.Response{
			Message: "Server Internal Error",
		})
	}))
	// Docs
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Version = "v1"
	docs.SwaggerInfo.Host = "localhost:8080"
	// CORS
	if settingsData.GO_ENV == "prod" {
		httpOrigin := "http://" + settingsData.CLIENT_URL
		httpsOrigin := "https://" + settingsData.CLIENT_URL
		router.Use(cors.New(cors.Config{
			AllowOrigins:     []string{httpOrigin, httpsOrigin},
			AllowMethods:     []string{"GET", "OPTIONS", "PUT", "DELETE", "POST"},
			AllowCredentials: true,
			AllowHeaders:     []string{"*"},
			AllowWebSockets:  false,
			MaxAge:           12 * time.Hour,
		}))
	} else {
		router.Use(cors.New(cors.Config{
			AllowAllOrigins: true,
			AllowHeaders:    []string{"*"},
			AllowMethods:    []string{"GET", "OPTIONS", "PUT", "DELETE", "POST"},
		}))
	}
	// Secure
	if settingsData.GO_ENV == "prod" {
		sslUrl := "ssl." + settingsData.CLIENT_URL
		secureConfig := secure.Config{
			SSLHost:              sslUrl,
			STSSeconds:           315360000,
			STSIncludeSubdomains: true,
			FrameDeny:            true,
			ContentTypeNosniff:   true,
			BrowserXssFilter:     true,
			IENoOpen:             true,
			ReferrerPolicy:       "strict-origin-when-cross-origin",
			SSLProxyHeaders: map[string]string{
				"X-Fowarded-Proto": "https",
			},
		}
		router.Use(secure.New(secureConfig))
	}
	/*if settingsData.NODE_ENV == "prod" {
		secureConfig.AllowedHosts = []string{
			settingsData.CLIENT_URL,
			sslUrl,
		}
	}*/
	// Rate limit
	store := ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
		Rate:  time.Second,
		Limit: 7,
	})
	mw := ratelimit.RateLimiter(store, &ratelimit.Options{
		ErrorHandler: ErrorHandler,
		KeyFunc:      keyFunc,
	})
	router.Use(mw)
	// Routes
	auth := router.Group(
		"/api/v1/auth",
	)
	user := router.Group(
		"/api/v1/user",
	)
	repo := router.Group(
		"/api/v1/repository",
	)
	dis := router.Group(
		"/api/v1/discussion",
	)
	{
		// Init controllers
		authController := new(controllers.AuthController)
		discussionController := new(controllers.DiscussionController)
		userController := new(controllers.UserController)
		repoController := new(controllers.RepositoryController)
		systemFileController := new(controllers.SystemFileController)
		// Define routes
		// Authentication
		auth.POST(
			"",
			authController.CreateUser,
		)
		auth.GET(
			"/confirm",
			authController.ConfirmUser,
		)
		auth.POST(
			"/login",
			authController.Login,
		)
		auth.POST(
			"/refresh",
			authController.RefreshToken,
		)
		// User
		user.GET(
			":idUser",
			userController.GetUser,
		)
		// Repository
		repo.GET(
			"",
			middlewares.JWTMiddleware(true),
			repoController.GetRepositories,
		)
		repo.GET(
			":username",
			middlewares.JWTMiddleware(true),
			repoController.GetUserRepositories,
		)
		repo.GET(
			":username/:repository",
			middlewares.JWTMiddleware(true),
			middlewares.RepoAccess(),
			middlewares.SetUserID(),
			repoController.GetRepository,
		)
		repo.GET(
			":username/:repository/:folder",
			middlewares.JWTMiddleware(true),
			middlewares.RepoAccess(),
			middlewares.SetUserID(),
			systemFileController.GetFolder,
		)
		repo.GET(
			"download/:repository",
			middlewares.JWTMiddleware(true),
			middlewares.RepoAccess(),
			repoController.DownloadRepository,
		)
		repo.POST(
			"",
			middlewares.JWTMiddleware(false),
			repoController.UploadRepository,
		)
		repo.POST(
			"like/:repository",
			middlewares.JWTMiddleware(false),
			repoController.ToggleLike,
		)
		repo.PUT(
			":repository",
			middlewares.JWTMiddleware(false),
			repoController.UpdateRepository,
		)
		repo.PUT(
			"element/:idRepository",
			middlewares.JWTMiddleware(false),
			systemFileController.NewRepoElement,
		)
		repo.PUT(
			"link/:repository",
			middlewares.JWTMiddleware(false),
			repoController.AddLink,
		)
		repo.DELETE(
			":repository",
			middlewares.JWTMiddleware(false),
			repoController.DeleteRepository,
		)
		repo.DELETE(
			":repository/:element",
			middlewares.JWTMiddleware(false),
			systemFileController.DeleteElement,
		)
		repo.DELETE(
			":repository/link/:link",
			middlewares.JWTMiddleware(false),
			repoController.DeleteLink,
		)
		// Discussion
		dis.POST(
			"",
			middlewares.JWTMiddleware(false),
			discussionController.UploadDiscussion,
		)
	}
	// Route docs
	router.GET("/api/v1/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	// Route healthz
	router.GET("/api/v1/healthz", func(ctx *gin.Context) {
		ctx.JSON(http.StatusNoContent, nil)
	})
	// No route
	router.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(404, res.Response{
			Message: "Not found",
		})
	})
	// Init server
	if err := router.Run(); err != nil {
		log.Fatalf("Error init server")
	}
}
