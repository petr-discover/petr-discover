package database

import (
	"context"
	"log"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/petr-discover/internal"
)

func NewNeo4jDB(ctx context.Context) (neo4j.DriverWithContext, error) {
	driver, err := internal.ConnectNeo4jDB(ctx)
	if err != nil {
		return nil, err
	}

	log.Println("Connected to Neo4j Database")

	return driver, nil
}
