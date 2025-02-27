package server

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/roundfeather/configuration-server/internal/config"
	"github.com/roundfeather/configuration-server/internal/controller"
	"github.com/roundfeather/configuration-server/internal/utils"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"strings"

	log "github.com/sirupsen/logrus"
)

func initConfig() (config.AppConfig, error) {
	var k = koanf.New(".")
	err := k.Load(file.Provider("properties.yml"), yaml.Parser())
	if err != nil {
		return config.AppConfig{}, err
	}
	err = k.Load(env.Provider("REPO_", ".", func(s string) string {
		return strings.Replace(strings.ToLower(s), "_", ".", -1)
	}), nil)
	if err != nil {
		return config.AppConfig{}, err
	}
	err = k.Load(env.Provider("APP_", ".", func(s string) string {
		return strings.Replace(strings.ToLower(s), "_", ".", -1)
	}), nil)
	if err != nil {
		return config.AppConfig{}, err
	}

	var cfg config.AppConfig
	err = k.Unmarshal("", &cfg)
	if err != nil {
		return config.AppConfig{}, err
	}

	return cfg, nil
}

func initTracer(ctx context.Context) (trace.TracerProvider, func(context.Context) error) {
	exporter, err := otlptrace.New(
		ctx,
		otlptracehttp.NewClient(
			otlptracehttp.WithInsecure(),
		),
	)
	if err != nil {
		fmt.Println(err)
	}

	resources, err := resource.New(
		context.Background(),
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithOS(),
		resource.WithContainer(),
		resource.WithHost(),
	)
	if err != nil {
		fmt.Println(err)
	}

	otel.SetTracerProvider(
		sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(resources),
		),
	)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))

	return otel.GetTracerProvider(), exporter.Shutdown
}

func hashMiddleware(hashValue string) func(echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("hash", hashValue)
			return next(c)
		}
	}
}

func cfgMiddleware(cfg config.Cfg) func(echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("cfg", cfg)
			return next(c)
		}
	}
}

func setupRepo(cfg config.RepoConfig) (string, error) {
	repo, err := git.PlainClone("config-repo", false, &git.CloneOptions{
		URL: cfg.Url,
		Auth: &http.BasicAuth{
			Username: cfg.Account,
			Password: cfg.Password,
		},
		SingleBranch:  false,
		ReferenceName: plumbing.NewBranchReferenceName(cfg.Branch),
	})
	if err != nil {
		return "", err
	}

	head, err := repo.Head()
	if err != nil {
		return "", err
	}
	return head.Hash().String(), nil
}

func Run() error {
	ctx := context.Background()

	cfg, err := initConfig()
	if err != nil {
		fmt.Println(err)
		return err
	}
	hash, err := setupRepo(cfg.Repo)
	if err != nil {
		fmt.Println(err)
		return err
	}

	tracerProvider, cleanup := initTracer(ctx)
	defer cleanup(ctx)

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(hashMiddleware(hash))
	e.Use(cfgMiddleware(cfg.App))
	e.Use(otelecho.Middleware("configuration-server", otelecho.WithTracerProvider(tracerProvider)))
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI: true, LogURIPath: true, LogStatus: true, LogMethod: true, LogLatency: true,
		LogQueryParams: []string{"service"},
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			lf := utils.GetLogFields(c)
			log.
				WithFields(lf).
				WithField("uri", v.URIPath).
				WithField("status", v.Status).
				WithField("method", v.Method).
				WithField("latency", v.Latency).
				WithField("query", v.QueryParams).
				Info()
			return nil
		},
	}))
	e.Use(middleware.BodyDump(func(c echo.Context, reqBody []byte, resBody []byte) {
		lf := utils.GetLogFields(c)
		log.WithFields(lf).Debug(string(resBody))
	}))

	e.GET("/v1/configuration", controller.GetV1Configuration)
	e.GET("/healthcheck/live", controller.Live)
	e.GET("/healthcheck/ready", controller.Ready)

	er := e.Start(":" + cfg.Server.Port)
	e.Logger.Fatal(er)
	return er
}
