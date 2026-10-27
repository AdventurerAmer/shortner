package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"time"

	"github.com/AdventurerAmer/shortner/apps/worker"
	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/errs"
	"github.com/AdventurerAmer/shortner/infra"
	"github.com/AdventurerAmer/shortner/internal/brokers"
	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/internal/repos/usersrepo"
	"github.com/AdventurerAmer/shortner/logging"
	"github.com/AdventurerAmer/shortner/mailer"
	"github.com/AdventurerAmer/shortner/telemetry"
)

func Run(templates *template.Template) int {
	cfg, err := config.Load()
	if err != nil {
		logger := logging.New()
		logger.Error("failed to load config", "error", err)
		return 1
	}

	workerCfg := &cfg.Workers.Clicks
	logger := cfg.NewLogger().With(slog.String("worker", workerCfg.Name))

	inf, err := infra.New()
	if err != nil {
		logger.Error("failed to create to infra", "error", err)
		return 1
	}
	var (
		postgresCtx infra.Postgres
		redisCtx    infra.Redis
	)
	inf.BindPostgres(cfg.Infra.Postgres, &postgresCtx)
	inf.BindRedis(cfg.Infra.Redis, &redisCtx)

	if err := inf.Start(context.Background()); err != nil {
		logger.Error("failed to connect to infra", "error", err)
		return 1
	}
	defer inf.Shutdown(context.Background())

	userCreatedReader := infra.NewKafkaReader(cfg.Infra.Kafka, workerCfg.Name, domain.UserCreatedTopic)
	defer func() {
		if err := userCreatedReader.Close(); err != nil {
			logger.Error("'reader.Close' failed", "error", err)
		}
	}()

	consumer := brokers.NewKafkaConsumer(userCreatedReader)

	app := worker.New(&cfg.Workers.Email, cfg, consumer, logger)

	usersRepo := usersrepo.NewPostgres(postgresCtx.Database)
	h := &handler{
		cfg:       cfg,
		usersRepo: usersRepo,
		templates: templates,
		mailer:    mailer.New(&cfg.Mailer),
	}

	handlerFunc := func(ctx context.Context, msg ports.ConsumerMessage) error {
		dctx, span := telemetry.NewSpan(ctx, "Email Worker: Handle",
			telemetry.StrAttr("key", msg.Key))
		defer span.End()

		logger := logging.Get(dctx)

		logger.Info("recived event", "status", "started", "key", msg.Key)
		defer logger.Info("recived event", "status", "ended", "key", msg.Key)

		var e domain.UserCreatedEvent
		if err := json.Unmarshal(msg.Data, &e); err != nil {
			return fmt.Errorf("'json.Unmarshal' failed: %w", err)
		}

		if err := h.verifyUser(ctx, e.UserId, e.VerificationToken, e.VerificationTokenExpiresAt); err != nil {
			return fmt.Errorf("'h.verifyUser' failed: %w", err)
		}

		return nil
	}
	if err := app.Run(handlerFunc); err != nil {
		logger.Error("'app.Run' failed", "error", err)
		return 1
	}

	return 0
}

type handler struct {
	cfg       *config.Config
	usersRepo ports.UsersRepository
	templates *template.Template
	mailer    *mailer.Mailer
}

func (h *handler) verifyUser(ctx context.Context, userId, token string, expiresAt time.Time) error {
	var user *domain.User

	for range 4 {
		var err error
		user, err = h.usersRepo.Get(ctx, userId)
		if err != nil {
			if errs.IsNotFound(err) {
				// we are sleeping here because we are still waiting on the database to commit the user
				time.Sleep(500 * time.Millisecond)
				continue
			} else {
				return fmt.Errorf("'usersRepo.Get' failed: %w", err)
			}
		} else {
			break
		}
	}

	if user.IsVerified {
		return nil
	}

	now := time.Now().UTC()
	if now.After(expiresAt) {
		return nil
	}

	addr := h.cfg.Services.Users.Addr()

	data := struct {
		Name      string
		VerifyURL string
		ExpiresIn time.Duration
	}{
		Name:      user.DisplayName,
		VerifyURL: fmt.Sprintf("http://%s/api/v1/users/verify?token=%s", addr, token),
		ExpiresIn: expiresAt.Sub(now),
	}

	// TODO: have a buffer bool here...
	var b bytes.Buffer
	if err := h.templates.ExecuteTemplate(&b, "verify.html", data); err != nil {
		return fmt.Errorf("'templates.ExecuteTemplate' failed: %w", err)
	}

	message := mailer.Message{
		To:      user.Email,
		Subject: "Verify your account",
		Body:    b.String(),
	}
	if err := h.mailer.Send(message); err != nil {
		return fmt.Errorf("'mailer.Send' failed: %w", err)
	}

	return nil
}
