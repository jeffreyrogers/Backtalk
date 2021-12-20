package globals

import (
	"context"
	"database/sql"
	"github.com/jeffreyrogers/backtalk/internal/sqlc"
)

var AuthKey []byte
var DB *sql.DB
var Queries *sqlc.Queries
var Ctx context.Context
