package repo

import (
	"github.com/luyb177/silent-sign-backend/internal/repo/verify"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Repositories struct {
	Verify verify.Repository
}

// NewRepositories creates Repositories with both Redis and MySQL.
func NewRepositories(redisClient *redis.Client, db *gorm.DB) *Repositories {
	return &Repositories{
		Verify: verify.NewVerifyRepo(redisClient),
	}
}
