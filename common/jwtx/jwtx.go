package jwtx

import (
	"errors"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

const defaultLeeway = 1 * time.Minute

type Claims struct {
	ClaimsParams
	jwtv5.RegisteredClaims
}

type ClaimsParams struct {
	UserID    uint64 `json:"user_id"`
	TokenType string `json:"token_type"`
}

type Handler interface {
	SetJWTToken(claims ClaimsParams) (string, error)
	ParseJWTToken(tokenString string) (*Claims, error)
}

type HandlerImpl struct {
	Secret      []byte
	TokenExpire time.Duration
	Leeway      time.Duration
}

func NewHandler(secret string, expire time.Duration) (Handler, error) {
	if secret == "" {
		return nil, errors.New("jwt secret cannot be empty")
	}

	if expire <= 0 {
		return nil, errors.New("jwt expire must be positive")
	}

	return &HandlerImpl{
		Secret:      []byte(secret),
		TokenExpire: expire,
		Leeway:      defaultLeeway,
	}, nil
}

func (h *HandlerImpl) SetJWTToken(claimsParams ClaimsParams) (string, error) {
	now := time.Now()

	claims := Claims{
		ClaimsParams: claimsParams,
		RegisteredClaims: jwtv5.RegisteredClaims{
			IssuedAt:  jwtv5.NewNumericDate(now),
			NotBefore: jwtv5.NewNumericDate(now),
			ExpiresAt: jwtv5.NewNumericDate(now.Add(h.TokenExpire)),
		},
	}

	token := jwtv5.NewWithClaims(
		jwtv5.SigningMethodHS256,
		claims,
	)

	return token.SignedString(h.Secret)
}

func (h *HandlerImpl) ParseJWTToken(tokenString string) (*Claims, error) {
	leeway := h.Leeway
	if leeway <= 0 {
		leeway = defaultLeeway
	}

	parser := jwtv5.NewParser(
		jwtv5.WithLeeway(leeway),
		jwtv5.WithExpirationRequired(),
		jwtv5.WithValidMethods(
			[]string{
				jwtv5.SigningMethodHS256.Alg(),
			},
		),
	)

	token, err := parser.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwtv5.Token) (any, error) {
			return h.Secret, nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
