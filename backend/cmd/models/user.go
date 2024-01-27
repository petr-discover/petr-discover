package models

import (
	"time"
)

type Member struct {
	ID        int64     `db:"id" dataType:"SERIAL PRIMARY KEY" constraint:"NOT NULL"`
	Username  string    `db:"username" dataType:"VARCHAR(50)" constraint:"NOT NULL UNIQUE"`
	Password  string    `db:"password" dataType:"VARCHAR(255)" constraint:"NOT NULL"`
	Email     string    `db:"email" dataType:"VARCHAR(50)" constraint:"NOT NULL UNIQUE"`
	CreatedAt time.Time `db:"created_at" dataType:"TIMESTAMP" constraint:"NOT NULL DEFAULT CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `db:"updated_at" dataType:"TIMESTAMP" constraint:"NOT NULL DEFAULT CURRENT_TIMESTAMP"`
}
