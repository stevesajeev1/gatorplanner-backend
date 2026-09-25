package dependencies

import (
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Optimizer struct {
	*grpc.ClientConn
}

func NewOptimizer(connStr string) *Optimizer {
	optimizer, err := grpc.NewClient(
		connStr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Panic(err)
	}

	return &Optimizer{optimizer}
}

func (o *Optimizer) Close() {
	_ = o.ClientConn.Close()
}
