package verify

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Repository interface {
	// SetCode 设置验证码并设置过期时间
	SetCode(ctx context.Context, meta *Meta, code string, expire time.Duration) error

	// VerifyCode 校验验证码，匹配成功则删除并返回 true；不匹配则保留并返回 false
	VerifyCode(ctx context.Context, meta *Meta, code string) (bool, error)
}

type repo struct {
	client *redis.Client
}

func NewVerifyRepo(client *redis.Client) Repository {
	return &repo{
		client: client,
	}
}
