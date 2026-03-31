package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/nats-io/nats.go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"petverse/server/internal/app"
	"petverse/server/internal/config"
	"petverse/server/internal/handler"
	"petverse/server/internal/pkg/ai"
	"petverse/server/internal/pkg/events"
	"petverse/server/internal/pkg/jwt"
	"petverse/server/internal/pkg/push"
	"petverse/server/internal/pkg/socialauth"
	"petverse/server/internal/pkg/timeseries"
	"petverse/server/internal/pkg/upload"
	"petverse/server/internal/repository"
	"petverse/server/internal/service"
	"petverse/server/internal/ws"
)

func main() {
	configPath := flag.String("config", "./config/config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		panic(fmt.Errorf("load config: %w", err))
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		panic(fmt.Errorf("open database: %w", err))
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic(fmt.Errorf("get sql db: %w", err))
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		panic(fmt.Errorf("ping database: %w", err))
	}

	timeseriesDB := db
	if cfg.Timeseries.Enabled {
		if cfg.Timeseries.DSN() == cfg.Database.DSN() {
			logger.Warn("timeseries_matches_primary_database", slog.String("database", cfg.Database.Name))
		} else {
			resolvedTimeseriesDB, err := gorm.Open(postgres.Open(cfg.Timeseries.DSN()), &gorm.Config{})
			if err != nil {
				panic(fmt.Errorf("open timeseries database: %w", err))
			}

			timeseriesSQLDB, err := resolvedTimeseriesDB.DB()
			if err != nil {
				panic(fmt.Errorf("get timeseries sql db: %w", err))
			}
			defer timeseriesSQLDB.Close()

			if err := timeseriesSQLDB.Ping(); err != nil {
				panic(fmt.Errorf("ping timeseries database: %w", err))
			}
			if err := timeseries.Bootstrap(resolvedTimeseriesDB); err != nil {
				panic(fmt.Errorf("bootstrap timeseries database: %w", err))
			}

			timeseriesDB = resolvedTimeseriesDB
		}
	}

	uploader, err := buildUploader(cfg.ObjectStorage)
	if err != nil {
		panic(fmt.Errorf("build uploader: %w", err))
	}

	var eventPublisher events.Publisher
	if cfg.NATS.Enabled {
		conn, err := nats.Connect(cfg.NATS.URL)
		if err != nil {
			panic(fmt.Errorf("connect nats: %w", err))
		}
		defer conn.Drain()

		eventPublisher = events.NewNATSPublisher(conn)
	}

	tokenManager := jwt.NewManager(cfg.JWT.Secret, cfg.JWT.Issuer, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)

	userRepo := repository.NewUserRepository(db)
	petRepo := repository.NewPetRepository(db)
	healthRepo := repository.NewHealthRepository(db)
	deviceRepo := repository.NewDeviceRepositoryWithTimeseries(db, timeseriesDB)
	serviceProviderRepo := repository.NewServiceProviderRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	communityRepo := repository.NewCommunityRepository(db)
	trainingRepo := repository.NewTrainingRepository(db)
	shopRepo := repository.NewShopRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	pushTokenRepo := repository.NewPushTokenRepository(db)
	aiGenerator := ai.NewTextGenerator(ai.TextGeneratorConfig{
		Provider:    cfg.AI.Provider,
		BaseURL:     cfg.AI.BaseURL,
		APIKey:      cfg.AI.APIKey,
		Model:       cfg.AI.Model,
		Temperature: cfg.AI.Temperature,
		Timeout:     cfg.AI.Timeout,
	})
	assistant := ai.NewAssistant(aiGenerator)
	pushDispatcher := push.NewDispatcher(push.Config{
		Provider:    cfg.Push.Provider,
		ExpoURL:     cfg.Push.ExpoURL,
		AccessToken: cfg.Push.AccessToken,
		Timeout:     cfg.Push.Timeout,
	})

	authService := service.NewAuthService(
		userRepo,
		tokenManager,
		service.WithAppleVerifier(buildAppleVerifier(cfg.SocialAuth)),
		service.WithGoogleVerifier(buildGoogleVerifier(cfg.SocialAuth)),
	)
	userService := service.NewUserService(userRepo)
	petService := service.NewPetService(petRepo)
	wsHub := ws.NewHub()
	healthAI := ai.NewHealthAI()
	healthService := service.NewHealthServiceWithOptions(petRepo, healthRepo, deviceRepo, healthAI, service.WithHealthAssistant(assistant))
	deviceService := service.NewDeviceService(deviceRepo, petRepo, wsHub, service.WithDeviceEvents(eventPublisher))
	serviceMarketService := service.NewServiceMarketService(serviceProviderRepo)
	bookingService := service.NewBookingService(bookingRepo, serviceProviderRepo, petRepo)
	communityService := service.NewCommunityService(communityRepo, petRepo, service.WithCommunityAssistant(assistant))
	notificationService := service.NewNotificationService(
		notificationRepo,
		wsHub,
		service.WithPushNotifications(pushTokenRepo, pushDispatcher),
		service.WithNotificationEvents(eventPublisher),
	)
	trainingService := service.NewTrainingService(
		trainingRepo,
		petRepo,
		notificationService,
		service.WithTrainingAssistant(assistant),
		service.WithTrainingEvents(eventPublisher),
	)
	shopService := service.NewShopService(shopRepo, petRepo)

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService, uploader)
	petHandler := handler.NewPetHandler(petService, uploader)
	healthHandler := handler.NewHealthHandler(healthService)
	deviceHandler := handler.NewDeviceHandler(deviceService, wsHub)
	serviceMarketHandler := handler.NewServiceMarketHandler(serviceMarketService)
	bookingHandler := handler.NewBookingHandler(bookingService)
	communityHandler := handler.NewCommunityHandler(communityService)
	trainingHandler := handler.NewTrainingHandler(trainingService)
	shopHandler := handler.NewShopHandler(shopService)
	notificationHandler := handler.NewNotificationHandler(notificationService, wsHub)
	uploadHandler := handler.NewUploadHandler(uploader)

	router := app.NewRouter(
		logger,
		tokenManager,
		authHandler,
		userHandler,
		petHandler,
		healthHandler,
		deviceHandler,
		serviceMarketHandler,
		bookingHandler,
		communityHandler,
		trainingHandler,
		shopHandler,
		notificationHandler,
		uploadHandler,
		cfg.ObjectStorage.LocalDir,
	)

	server := &http.Server{
		Addr:         cfg.HTTP.Addr(),
		Handler:      router,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	logger.Info("server_starting", slog.String("addr", cfg.HTTP.Addr()), slog.String("env", cfg.App.Env))

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(fmt.Errorf("listen and serve: %w", err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server_shutdown_failed", slog.String("error", err.Error()))
		return
	}

	logger.Info("server_stopped")
}

func buildUploader(cfg config.ObjectStorageConfig) (upload.Store, error) {
	provider := strings.ToLower(strings.TrimSpace(cfg.Provider))
	switch provider {
	case "", "local":
		localDir := strings.TrimSpace(cfg.LocalDir)
		if localDir == "" {
			localDir = "./uploads"
		}
		return upload.NewLocalStore(localDir), nil
	case "minio":
		return upload.NewMinIOStore(upload.MinIOConfig{
			Endpoint:      cfg.Endpoint,
			AccessKey:     cfg.AccessKey,
			SecretKey:     cfg.SecretKey,
			UseSSL:        cfg.UseSSL,
			Bucket:        cfg.Bucket,
			PublicBaseURL: cfg.PublicBaseURL,
		})
	default:
		return nil, fmt.Errorf("unsupported object storage provider: %s", cfg.Provider)
	}
}

func buildAppleVerifier(cfg config.SocialAuthConfig) socialauth.Verifier {
	if len(cfg.AppleAudiences) == 0 {
		return nil
	}
	return socialauth.NewOIDCVerifier(socialauth.OIDCVerifierConfig{
		JWKSURL:          socialauth.AppleJWKSURL,
		AllowedIssuers:   []string{"https://appleid.apple.com"},
		AllowedAudiences: cfg.AppleAudiences,
		HTTPTimeout:      cfg.HTTPTimeout,
		CacheTTL:         cfg.CacheTTL,
	})
}

func buildGoogleVerifier(cfg config.SocialAuthConfig) socialauth.Verifier {
	if len(cfg.GoogleClientIDs) == 0 {
		return nil
	}
	return socialauth.NewOIDCVerifier(socialauth.OIDCVerifierConfig{
		JWKSURL:          socialauth.GoogleJWKSURL,
		AllowedIssuers:   []string{"https://accounts.google.com", "accounts.google.com"},
		AllowedAudiences: cfg.GoogleClientIDs,
		HTTPTimeout:      cfg.HTTPTimeout,
		CacheTTL:         cfg.CacheTTL,
	})
}
