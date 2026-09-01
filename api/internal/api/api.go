package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stevesajeev1/gatorplanner-backend/internal/config"
	"github.com/stevesajeev1/gatorplanner-backend/internal/database/repository/elasticsearch"
	"github.com/stevesajeev1/gatorplanner-backend/internal/dependencies"
	"github.com/stevesajeev1/gatorplanner-backend/internal/domains/classes"
	"github.com/stevesajeev1/gatorplanner-backend/internal/logger"
)

func Run() {
	logger := logger.New()
	config := config.LoadConfig()

	// Initialize dependencies
	db := dependencies.NewDB(config.DatabaseURL)
	defer db.Close()

	es := dependencies.NewES(config.ElasticsearchURL)
	defer es.Close()

	// Router
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	humaConfig := huma.DefaultConfig("GatorPlanner API", "1.0.0")
	humaConfig.DocsRenderer = huma.DocsRendererScalar
	humaConfig.CreateHooks = nil

	api := humachi.New(r, humaConfig)

	// Repositories Setup
	classesESRepo := elasticsearch.NewClassesESRepository(es)

	// Routes registration
	classesService := classes.NewService(classesESRepo, logger)
	classesHandler := classes.NewHandler(classesService, logger)
	classes.RegisterRoutes(classesHandler, huma.NewGroup(api, "/term/{termID}/classes"))

	huma.Register(api, huma.Operation{
		OperationID: "ping",
		Method:      http.MethodGet,
		Summary:     "Ping",
		Description: "Health Check",
		Tags:        []string{"Misc"},
		Path:        "/ping",
	}, func(ctx context.Context, input *struct{}) (*struct{ Body string }, error) {
		return &struct{ Body string }{
			Body: "pong",
		}, nil
	})

	logger.Info().Msgf("API listening on port %s", config.Port)
	if err := http.ListenAndServe(":"+config.Port, r); err != nil {
		logger.Panic().Msg("Failed to start server.")
	}
}
