package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/elastic/go-elasticsearch/v9"
	"github.com/go-chi/chi/v5"
	"github.com/stevesajeev1/gatorplanner-backend/internal/config"
	"github.com/stevesajeev1/gatorplanner-backend/internal/logger"
)

func Run() {
	logger := logger.New()
	config := config.LoadConfig()

	es, err := elasticsearch.NewTyped(
		elasticsearch.WithAddresses(config.ElasticsearchURL),
	)
	if err != nil {
		logger.Fatal().Msg("Failed to connect to Elasticsearch")
	}
	defer es.Close(context.Background())

	r := chi.NewRouter()

	humaConfig := huma.DefaultConfig("GatorPlanner API", "1.0.0")
	humaConfig.DocsRenderer = huma.DocsRendererScalar
	humaConfig.CreateHooks = nil

	api := humachi.New(r, humaConfig)

	// Repositories Setup
	classesDBRepo := repository.NewClassesDBRepo(db)
	classesESRepo := repository.NewClassesESRepo(es)

	// Routes registration
	// classesService := classes.NewService()
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
		logger.Fatal().Msg("Failed to start server.")
	}
}
