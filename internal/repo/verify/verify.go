package verify

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// verify:code:{channel}:{purpose}:{target_hash}
	CodeKey = "verify:code:%d:%d:%s"
)

type Meta struct {
	Target  string
	Channel int32
	Purpose int32
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
