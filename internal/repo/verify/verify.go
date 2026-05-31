package verify

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/luyb177/silent-sign-backend/internal/constvar"
	"github.com/redis/go-redis/v9"
)

const (
	// verify:code:{channel}:{purpose}:{target_hash}
	CodeKey = "verify:code:%d:%d:%s"
)

type Meta struct {
	Target  string
	Channel constvar.VerificationChannel
	Purpose constvar.VerificationPurpose
}

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

func (r *repo) SetCode(ctx context.Context, meta *Meta, code string, expire time.Duration) error {
	key := verifyCodeKey(meta)

	return r.client.Set(ctx, key, code, expire).Err()
}

func (r *repo) VerifyCode(ctx context.Context, meta *Meta, code string) (bool, error) {
	key := verifyCodeKey(meta)

	val, err := verifyScript.Run(
		ctx,
		r.client,
		[]string{key},
		code,
	).Result()

	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	matched, ok := val.(int64)
	if !ok {
		return false, fmt.Errorf("unexpected redis result type: %T", val)
	}

	return matched == 1, nil
}

func verifyCodeKey(meta *Meta) string {
	// target 做 hash，避免邮箱/手机号直接暴露在 key 中
	sum := sha256.Sum256([]byte(meta.Target))
	targetHash := hex.EncodeToString(sum[:])

	return fmt.Sprintf(CodeKey, meta.Channel, meta.Purpose, targetHash)
}
