package database

import (
	"context"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var DBMain *DB

var Neo4jDriver neo4j.DriverWithContext

var Neo4jCtx context.Context
