package app

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"petverse/server/internal/handler"
	"petverse/server/internal/middleware"
	petjwt "petverse/server/internal/pkg/jwt"
)

func NewRouter(
	logger *slog.Logger,
	tokenManager *petjwt.Manager,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	petHandler *handler.PetHandler,
	healthHandler *handler.HealthHandler,
	deviceHandler *handler.DeviceHandler,
	serviceMarketHandler *handler.ServiceMarketHandler,
	bookingHandler *handler.BookingHandler,
	communityHandler *handler.CommunityHandler,
	trainingHandler *handler.TrainingHandler,
	shopHandler *handler.ShopHandler,
	notificationHandler *handler.NotificationHandler,
	uploadHandler *handler.UploadHandler,
	uploadDir string,
) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger(logger))
	router.Use(middleware.CORS())

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.Static("/uploads", uploadDir)

	apiV1 := router.Group("/api/v1")

	authGroup := apiV1.Group("/auth")
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)
	authGroup.POST("/refresh", authHandler.Refresh)
	authGroup.POST("/login/wechat", authHandler.LoginWechat)
	authGroup.POST("/login/apple", authHandler.LoginApple)
	authGroup.POST("/login/google", authHandler.LoginGoogle)

	protected := apiV1.Group("/")
	protected.Use(middleware.Auth(tokenManager))

	users := protected.Group("/users")
	users.GET("/me", userHandler.Me)
	users.PUT("/me", userHandler.UpdateMe)
	users.PUT("/me/avatar", userHandler.UploadAvatar)
	users.PUT("/me/location", userHandler.UpdateLocation)

	pets := protected.Group("/pets")
	pets.GET("", petHandler.List)
	pets.POST("", petHandler.Create)
	pets.GET("/:id", petHandler.Get)
	pets.PUT("/:id", petHandler.Update)
	pets.DELETE("/:id", petHandler.Delete)
	pets.PUT("/:id/avatar", petHandler.UploadAvatar)

	health := protected.Group("/pets/:id/health")
	health.GET("", healthHandler.List)
	health.POST("", healthHandler.Create)
	health.GET("/summary", healthHandler.Summary)
	health.GET("/alerts", healthHandler.Alerts)
	health.POST("/ask-ai", healthHandler.AskAI)
	health.GET("/:recordId", healthHandler.Get)
	health.PUT("/:recordId", healthHandler.Update)
	health.DELETE("/:recordId", healthHandler.Delete)

	training := protected.Group("/pets/:id/training")
	training.GET("", trainingHandler.List)
	training.POST("", trainingHandler.Create)
	training.POST("/generate", trainingHandler.Generate)

	devices := protected.Group("/devices")
	devices.GET("", deviceHandler.List)
	devices.POST("", deviceHandler.Create)
	devices.GET("/:id", deviceHandler.Get)
	devices.PUT("/:id", deviceHandler.Update)
	devices.DELETE("/:id", deviceHandler.Delete)
	devices.POST("/:id/command", deviceHandler.Command)
	devices.GET("/:id/data", deviceHandler.Data)
	devices.GET("/:id/status", deviceHandler.Status)
	devices.GET("/:id/stream", deviceHandler.Stream)

	services := protected.Group("/services")
	services.GET("", serviceMarketHandler.List)
	services.GET("/:id", serviceMarketHandler.Get)
	services.GET("/:id/reviews", serviceMarketHandler.Reviews)
	services.GET("/:id/availability", serviceMarketHandler.Availability)

	bookings := protected.Group("/bookings")
	bookings.POST("", bookingHandler.Create)
	bookings.GET("", bookingHandler.List)
	bookings.GET("/:id", bookingHandler.Get)
	bookings.PUT("/:id/cancel", bookingHandler.Cancel)
	bookings.PUT("/:id/review", bookingHandler.Review)

	posts := protected.Group("/posts")
	posts.GET("", communityHandler.ListPosts)
	posts.POST("", communityHandler.CreatePost)
	posts.GET("/:id", communityHandler.GetPost)
	posts.PUT("/:id", communityHandler.UpdatePost)
	posts.DELETE("/:id", communityHandler.DeletePost)
	posts.POST("/:id/like", communityHandler.ToggleLike)
	posts.GET("/:id/comments", communityHandler.ListComments)
	posts.POST("/:id/comments", communityHandler.CreateComment)

	protected.DELETE("/comments/:id", communityHandler.DeleteComment)
	protected.POST("/community/ask-ai", communityHandler.AskAI)

	trainingPlans := protected.Group("/training")
	trainingPlans.GET("/:id", trainingHandler.Get)
	trainingPlans.PUT("/:id", trainingHandler.Update)
	trainingPlans.DELETE("/:id", trainingHandler.Delete)

	shop := protected.Group("/shop")
	shop.GET("/products", shopHandler.ListProducts)
	shop.GET("/products/:id", shopHandler.GetProduct)
	shop.GET("/recommendations/:petId", shopHandler.Recommendations)

	notifications := protected.Group("/notifications")
	notifications.GET("", notificationHandler.List)
	notifications.PUT("/read-all", notificationHandler.ReadAll)
	notifications.PUT("/:id/read", notificationHandler.Read)
	notifications.POST("/push-token", notificationHandler.RegisterPushToken)
	notifications.DELETE("/push-token", notificationHandler.UnregisterPushToken)
	protected.GET("/ws", notificationHandler.Stream)

	uploads := protected.Group("/upload")
	uploads.POST("/image", uploadHandler.UploadImage)
	uploads.POST("/file", uploadHandler.UploadFile)

	return router
}
