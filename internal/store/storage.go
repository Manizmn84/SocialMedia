package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// QueryTimeoutDuration bounds each store operation, including an entire delete transaction.
const QueryTimeoutDuration = 5 * time.Second

var (
	ErrNotFound     = errors.New("resource not found")
	ErrEditConflict = errors.New("post has changed or been deleted; reload and try again")
)

type Storage struct {
	Posts interface {
		Create(context.Context, *Post) error
		GetByID(context.Context, uint) (Post, error)
		Update(context.Context, *Post) error
		Delete(context.Context, int64) error
	}

	Users interface {
		Create(context.Context, *User) error
	}

	Comments interface {
		Create(ctx context.Context, comment *Comment) error
		GetByPostID(ctx context.Context, id uint) ([]Comment, error)
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts:    &PostsStore{db: db},
		Users:    &UsersStore{db: db},
		Comments: &CommentStore{db: db},
	}
}
