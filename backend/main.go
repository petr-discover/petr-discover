package main

import (
	"context"
	"log"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/petr-discover/cmd/database"
	"github.com/petr-discover/cmd/models"
	"github.com/petr-discover/cmd/routes"
)

var err error

func main() {
	database.DBMain, err = database.NewSQLDB("pgx")
	if err != nil {
		log.Fatal(err)
	}

	database.DBMain.LionMigrate(&models.Member{})

	defer func() {
		if err = database.DBMain.Close(); err != nil {
			panic(err)
		}
		log.Println("Disconnected from SQL Database")
	}()
	var cancel context.CancelFunc

	database.Neo4jCtx, cancel = context.WithCancel(context.Background())
	defer cancel()

	database.Neo4jDriver, err = database.NewNeo4jDB(database.Neo4jCtx)

	defer func() {
		if err = database.Neo4jDriver.Close(database.Neo4jCtx); err != nil {
			panic(err)
		}
		log.Println("Disconnected from Neo4j Database")
	}()

	// ctx := context.Background()
	// client, err := storage.NewClient(ctx, option.WithCredentialsFile("auth.json"))
	// if err != nil {
	// 	log.Fatalf("Failed to create client: %v", err)
	// }
	// defer client.Close()
	r := routes.NewRouter(":8080")
	http.ListenAndServe(":8080", r)
}
