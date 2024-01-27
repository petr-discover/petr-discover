package internal

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/petr-discover/config"
)

func ConnectSQLDB(dbDriver string) (*sql.DB, error) {
	urlDB := config.PostgresDBConfig()
	db, err := sql.Open(dbDriver, urlDB)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	log.Println("Database Connection Success!")
	return db, nil
}

func ConnectNeo4jDB(ctx context.Context) (neo4j.SessionWithContext, error) {
	uri := config.Neo4jDBConfig()
	username := config.Neo4jUser
	password := config.Neo4jPass
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(username, password, ""))

	if err != nil {
		return nil, err
	}

	defer driver.Close(ctx)

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})

	return session, nil
}
