package repo

import (
	"github.com/luyb177/silent-sign-backend/internal/repo/comment"
	"github.com/luyb177/silent-sign-backend/internal/repo/image"
	"github.com/luyb177/silent-sign-backend/internal/repo/moment"
	"github.com/luyb177/silent-sign-backend/internal/repo/user"
	"github.com/luyb177/silent-sign-backend/internal/repo/verify"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Repositories struct {
	Verify  verify.Repository
	User    user.Repository
	Moment  moment.Repository
	Image   image.Repository
	Comment comment.Repository
	db      *gorm.DB
}

// NewRepositories creates Repositories with both Redis and MySQL.
func NewRepositories(redisClient *redis.Client, db *gorm.DB) *Repositories {
	return &Repositories{
		Verify:  verify.NewVerifyRepo(redisClient),
		User:    user.NewRepository(redisClient, db),
		Moment:  moment.NewRepository(db, redisClient),
		Image:   image.NewRepository(db, redisClient),
		Comment: comment.NewRepository(db, redisClient),
		db:      db,
	}
}

// Transaction 开启 MySQL 事务。Redis 操作不适合放入事务中，应在事务前后执行。
func (r *Repositories) Transaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}
