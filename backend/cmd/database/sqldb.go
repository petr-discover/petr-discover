package database

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/petr-discover/internal"
)

type DB struct {
	*sql.DB
}

func (d *DB) LionMigrate(dbModel interface{}) {
	t := reflect.TypeOf(dbModel)
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
		log.Println("Invalid model type. Migration failed")
		return
	}

	tableName := t.Elem().Name()
	var argsSQL []string

	for i := 0; i < t.Elem().NumField(); i++ {
		field := t.Elem().Field(i)
		tags := field.Tag
		columnName := tags.Get("db")
		dataType := tags.Get("dataType")
		constraint := tags.Get("constraint")
		clause := strings.TrimSpace(columnName + " " + dataType + " " + constraint)
		argsSQL = append(argsSQL, clause)
	}
	sqlArg := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", tableName, strings.Join(argsSQL, ", "))

	_, err := d.Exec(sqlArg)
	if err != nil {
		log.Printf("Table creation failed: %v", err)
		return
	}
	log.Printf("Successfully Migrated Table: %s", tableName)
}

func NewSQLDB(dbDriver string) (*DB, error) {
	db, err := internal.ConnectSQLDB(dbDriver)
	if err != nil {
		return nil, err
	}
	log.Println("Connected to SQL Database")
	return &DB{db}, nil
}
