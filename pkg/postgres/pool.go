package postgres

import (
	"database/sql"
	"time"
)

// todo: decide where pool limits live (code or config), see the same note in application/auth/limits.go
const (
	MaxOpenConns    = 25
	MaxIdleConns    = 10
	ConnMaxLifetime = 30 * time.Minute
	ConnMaxIdleTime = 5 * time.Minute
)

func configurePool(db *sql.DB) {
	db.SetMaxOpenConns(MaxOpenConns)
	db.SetMaxIdleConns(MaxIdleConns)
	db.SetConnMaxLifetime(ConnMaxLifetime)
	db.SetConnMaxIdleTime(ConnMaxIdleTime)
}
