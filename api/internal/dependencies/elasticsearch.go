package dependencies

import (
	"context"
	"log"

	"github.com/elastic/go-elasticsearch/v9"
)

type ES struct {
	*elasticsearch.TypedClient
}

func NewES(connStr string) *ES {
	es, err := elasticsearch.NewTyped(
		elasticsearch.WithAddresses(connStr),
	)
	if err != nil {
		log.Panic(err)
	}

	return &ES{es}
}

func (e *ES) Close() {
	e.TypedClient.Close(context.Background())
}