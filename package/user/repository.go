package user

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ngoldack/dicetrace/package/core"
)

type UserRepository interface {
	SaveUser(ctx context.Context, usr *core.User) error
	GetUserByUsername(ctx context.Context, username string) (*core.User, error)
}

type PostgreSQLUserRepository struct {
	pool pgxpool.Conn
}
