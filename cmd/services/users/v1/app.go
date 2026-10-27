package v1

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/AdventurerAmer/shortner/apps/web"
	"github.com/AdventurerAmer/shortner/async/goorch"
	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/health"
	"github.com/AdventurerAmer/shortner/infra"
	"github.com/AdventurerAmer/shortner/internal/brokers"
	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/services/users"
	"github.com/AdventurerAmer/shortner/internal/password"
	"github.com/AdventurerAmer/shortner/internal/repos/tokensrepo"
	"github.com/AdventurerAmer/shortner/internal/repos/usersrepo"
	"github.com/AdventurerAmer/shortner/internal/tokens"
	"github.com/AdventurerAmer/shortner/logging"
	"github.com/AdventurerAmer/shortner/postgresutils"
)

func Run() int {
	cfg, err := config.Load()
	if err != nil {
		logger := logging.New()
		logger.Error("failed to load config", "error", err)
		return 1
	}

	serviceCfg := &cfg.Services.Users
	logger := cfg.NewLogger().With(slog.String("service", serviceCfg.Name))

	var (
		postgresCtx infra.Postgres
		redisCtx    infra.Redis
	)

	inf, err := infra.New()
	if err != nil {
		logger.Error("'infra.New()' failed", "error", err)
		return 1
	}
	inf.BindRedis(cfg.Infra.Redis, &redisCtx)
	inf.BindPostgres(cfg.Infra.Postgres, &postgresCtx)

	if err := inf.Start(context.Background()); err != nil {
		logger.Error("infrastructure connection failed", "error", err)
		return 1
	}
	defer inf.Shutdown(context.Background())

	userCreatedWriter := infra.NewKafkaWriter(cfg.Infra.Kafka, domain.UserCreatedTopic)
	defer func() {
		err := userCreatedWriter.Close()
		if err != nil {
			logger.Error("failed to close kafka writer", "error", err)
		}
	}()

	userVerificationWriter := infra.NewKafkaWriter(cfg.Infra.Kafka, domain.UserVerificationTopic)
	defer func() {
		err := userCreatedWriter.Close()
		if err != nil {
			logger.Error("failed to close kafka writer", "error", err)
		}
	}()

	userPasswordResetWriter := infra.NewKafkaWriter(cfg.Infra.Kafka, domain.UserVerificationTopic)
	defer func() {
		err := userCreatedWriter.Close()
		if err != nil {
			logger.Error("failed to close kafka writer", "error", err)
		}
	}()

	userCreatedProducer := brokers.NewKafkaProducer(userCreatedWriter)
	userVerificationProducer := brokers.NewKafkaProducer(userVerificationWriter)
	userPasswordResetProducer := brokers.NewKafkaProducer(userPasswordResetWriter)

	// redisCache := caches.NewRedis(redisCtx.Client)
	usersRepo := usersrepo.NewPostgres(postgresCtx.Database)
	tokensrepo := tokensrepo.NewPostgres(postgresCtx.Database)

	srvCfg := users.Config{
		Transactor:                postgresutils.NewTransactor(postgresCtx.Database),
		UsersRepo:                 usersRepo,
		PasswordManager:           password.NewArgon2(),
		UserCreatedProducer:       userCreatedProducer,
		UserVerficiationProducer:  userVerificationProducer,
		UserPasswordResetProducer: userPasswordResetProducer,
		VerificationTokenManager: tokens.NewManager(
			tokens.NewHMAC(cfg.Tokens.UserVerification.Secret),
			tokensrepo, domain.TokenTypeVerification, cfg.Tokens.UserVerification.ExpiresAfter),
		PasswordResetTokenManager: tokens.NewManager(
			tokens.NewHMAC(cfg.Tokens.UserPasswordReset.Secret),
			tokensrepo, domain.TokenTypePasswordReset, cfg.Tokens.UserPasswordReset.ExpiresAfter),
	}
	service := users.New(srvCfg)

	orch := goorch.New(context.Background())
	defer orch.CancelAndWait()

	readiness := func(ctx context.Context, checks health.Checks) error {
		if err := postgresCtx.Database.PingContext(ctx); err != nil {
			checks["postgres"] = err.Error()
			return fmt.Errorf("'postgresCtx.Database.PingContext' failed: %w", err)
		}

		checks["postgres"] = "up"

		if _, err := redisCtx.Client.Ping(ctx).Result(); err != nil {
			checks["redis"] = err.Error()
			return fmt.Errorf("'redisCtx.Client.Ping' failed: %w", err)
		}

		checks["redis"] = "up"

		if err := infra.PingKafka(ctx, cfg.Infra.Kafka); err != nil {
			checks["kafka"] = err.Error()
			return fmt.Errorf("'infra.PingKafka' failed: %w", err)
		}

		checks["kafka"] = "up"

		return nil
	}

	handlers := newHandlers(service)

	mux := web.NewMux(logger)

	mux.Use(web.Trace)
	mux.Use(web.Measure)
	mux.Use(web.Logging)
	mux.Use(web.Recover(cfg.Env))
	mux.Use(web.Timeout(serviceCfg.DefaultTimeout))

	mux.Post("/v1/users/register", handlers.register)
	mux.Get("/v1/users/me", handlers.get)
	mux.Put("/v1/users/me", handlers.update)
	mux.Delete("/v1/users/me", handlers.delete)

	mux.Post("/v1/users/verify", handlers.sendVerification)
	mux.Get("/v1/users/verify", handlers.verify)

	mux.Post("/v1/users/change-password", handlers.changePassword)
	mux.Post("/v1/users/forgot-password", handlers.forgotPassword)
	mux.Get("/v1/users/reset-password", handlers.resetPassword)

	app := web.New(serviceCfg, cfg, logger)
	if err := app.Run(mux, readiness); err != nil {
		logger.Error("'app.Run' failed", "error", err)
		return 1
	}

	return 0
}
