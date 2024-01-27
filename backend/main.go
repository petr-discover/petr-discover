package main

import (
	"fmt"

	"github.com/petr-discover/config"
)

func main() {
	s := config.Neo4jDBConfig()
	fmt.Println(s)
}
