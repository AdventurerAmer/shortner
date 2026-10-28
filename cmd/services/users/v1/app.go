package v1

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/AdventurerAmer/shortner/apps/web"
	"github.com/AdventurerAmer/shortner/async/goorch"
	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/health"
	"github.com/AdventurerAmer/shortner/infra"
	"github.com/AdventurerAmer/shortner/internal/brokers"
	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/internal/core/services/auth"
	"github.com/AdventurerAmer/shortner/internal/core/services/users"
	"github.com/AdventurerAmer/shortner/internal/repos/tokensrepo"
	"github.com/AdventurerAmer/shortner/internal/repos/usersrepo"
	"github.com/AdventurerAmer/shortner/internal/tokens"
	"github.com/AdventurerAmer/shortner/logging"
	"github.com/AdventurerAmer/shortner/password"
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

	authServiceCfg := auth.Config{
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
		JWTSecret:            cfg.Auth.Secret,
		JWTIssuer:            cfg.App.Name,
		JWTTokenExpiresAfter: cfg.Auth.TokenExpiresAfter,
	}
	authService := auth.New(authServiceCfg)

	usersServiceCfg := users.Config{
		UsersRepo: usersRepo,
	}

	usersService := users.New(usersServiceCfg)

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

	authMiddleware := newAuthMiddleware(authService)

	handlers := newHandlers(usersService)

	mux := web.NewMux(logger)

	mux.Use(web.Trace)
	mux.Use(web.Measure)
	mux.Use(web.Logging)
	mux.Use(web.Recover(cfg.Env))
	mux.Use(web.Timeout(serviceCfg.DefaultTimeout))
	mux.Use(authMiddleware)

	mux.Get("/api/v1/users/me", handlers.get)
	mux.Put("/api/v1/users/me", handlers.update)

	// mux.Post("/api/v1/users/register", handlers.register)
	// mux.Post("/api/v1/users/verify", handlers.sendVerification)
	// mux.Get("/api/v1/users/verify", handlers.verify)

	// mux.Post("/api/v1/users/change-password", handlers.changePassword)
	// mux.Post("/api/v1/users/forgot-password", handlers.forgotPassword)
	// mux.Get("/api/v1/users/reset-password", handlers.resetPassword)

	app := web.New(serviceCfg, cfg, logger)
	if err := app.Run(mux, readiness); err != nil {
		logger.Error("'app.Run' failed", "error", err)
		return 1
	}

	return 0
}

type userIdContextKey struct{}

func userIdFromContext(ctx context.Context) string {
	userId, _ := ctx.Value(userIdContextKey{}).(string)
	return userId
}

func newAuthMiddleware(auth ports.AuthService) web.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			prefix := "Bearer "
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, prefix) {
				web.WriteJSON(map[string]any{"message": "invalid authorization header"}, w)
				return
			}
			tokenStr := strings.TrimPrefix(authHeader, prefix)

			resp, err := auth.VerifyToken(r.Context(), ports.VerifyTokenRequest{Token: tokenStr})
			if err != nil {
				web.WriteJSON(map[string]any{"message": "invalid or expried token"}, w)
				return
			}

			userId := resp.UserId
			dctx := context.WithValue(r.Context(), userIdContextKey{}, userId)
			next(w, r.WithContext(dctx))
		}
	}
}
