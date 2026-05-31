package verify

import (
	_ "embed"

	"github.com/redis/go-redis/v9"
)

//go:embed lua/verify.lua
var verifyLua string

var verifyScript = redis.NewScript(verifyLua)
