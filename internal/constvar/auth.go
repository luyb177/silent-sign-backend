package constvar

import "time"

// Context keys
type contextKey string

const (
	IPLocationKey contextKey = "ip_location"
	AuthUserKey   contextKey = "auth_user"
)

const (
	ChannelEmail int32 = 1
	ChannelPhone int32 = 2
)

const (
	PurposeRegistration  int32 = 1
	PurposePasswordReset int32 = 2
)

const VerifyCodeExpire = 5 * time.Minute
