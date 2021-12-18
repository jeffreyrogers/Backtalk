package models

import (
	"context"
	"database/sql"
	"github.com/jeffreyrogers/backtalk/internal/sqlc"
)

var DB *sql.DB
var Queries *sqlc.Queries
var Ctx context.Context
