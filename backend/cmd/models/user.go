package models

import (
	"time"
)

type Members struct {
	ID        int64     `db:"id" dataType:"SERIAL PRIMARY KEY" constraint:"NOT NULL"`
	Username  string    `db:"username" dataType:"VARCHAR(50)" constraint:"NOT NULL UNIQUE"`
	Password  string    `db:"password" dataType:"VARCHAR(50)" constraint:"NOT NULL"`
	Email     string    `db:"email" dataType:"VARCHAR(50)" constraint:"NOT NULL UNIQUE"`
	CreatedAt time.Time `db:"created_at" dataType:"TIMESTAMP" constraint:"NOT NULL DEFAULT CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `db:"updated_at" dataType:"TIMESTAMP" constraint:"NOT NULL DEFAULT CURRENT_TIMESTAMP"`
}

type Session struct {
	ID        int64     `db:"id" dataType:"SERIAL PRIMARY KEY" constraint:"NOT NULL"`
	UserID    int64     `db:"user_id" dataType:"INT" constraint:"REFERENCES members(id) ON DELETE CASCADE"`
	Token     string    `db:"token" dataType:"VARCHAR(255)" constraint:"NOT NULL UNIQUE"`
	CreatedAt time.Time `db:"created_at" dataType:"TIMESTAMP" constraint:"NOT NULL DEFAULT CURRENT_TIMESTAMP"`
	ExpiresAt time.Time `db:"expires_at" dataType:"TIMESTAMP" constraint:"NOT NULL"`
}

type GoogleAuth struct {
	ID         int64     `db:"id" dataType:"SERIAL PRIMARY KEY" constraint:"REFERENCES members(id) ON DELETE CASCADE"`
	GoogleID   string    `db:"google_id" dataType:"VARCHAR(50)" constraint:"NOT NULL UNIQUE"`
	Gmail      string    `db:"gmail" dataType:"VARCHAR(100)" constraint:"NOT NULL UNIQUE"`
	GoogleName string    `db:"google_name" dataType:"VARCHAR(100)" constraint:""`
	LoggedAt   time.Time `db:"logged_at" dataType:"TIMESTAMP" constraint:"NOT NULL DEFAULT CURRENT_TIMESTAMP"`
}
