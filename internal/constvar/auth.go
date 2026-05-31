package constvar

import "time"

type VerificationChannel int32

const (
	ChannelEmail VerificationChannel = 1
	ChannelPhone VerificationChannel = 2
)

type VerificationPurpose int32

const (
	PurposeRegistration  VerificationPurpose = 1
	PurposePasswordReset VerificationPurpose = 2
)

const VerifyCodeExpire = 5 * time.Minute
