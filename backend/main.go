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

	database.DBMain.LionMigrate(&models.Members{})
	database.DBMain.LionMigrate(&models.Session{})
	database.DBMain.LionMigrate(&models.GoogleAuth{})

	defer func() {
		if err = database.DBMain.Close(); err != nil {
			panic(err)
		}
		log.Println("Disconnected from SQL Database")
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	database.Neo4jSession, err = database.NewNeo4jDB(ctx)

	defer func() {
		if err = database.Neo4jSession.Close(ctx); err != nil {
			panic(err)
		}
		log.Println("Disconnected from Neo4j Database")
	}()

	r := routes.NewRouter(":8080")
	http.ListenAndServe(":8080", r)
}
