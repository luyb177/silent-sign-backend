package jwtx

import (
	"errors"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

const defaultLeeway = 1 * time.Minute

// TokenType 常量，用于区分访问令牌和刷新令牌
const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

type Claims struct {
	ClaimsParams
	jwtv5.RegisteredClaims
}

type ClaimsParams struct {
	UserID    uint64 `json:"user_id"`
	TokenType string `json:"token_type"`
}

type Handler interface {
	SetAccessToken(claims ClaimsParams) (string, error)
	SetRefreshToken(claims ClaimsParams) (string, error)
	ParseJWTToken(tokenString string) (*Claims, error)
}

type HandlerImpl struct {
	Secret        []byte
	AccessExpire  time.Duration
	RefreshExpire time.Duration
	Leeway        time.Duration
}

func NewHandler(secret string, accessExpire, refreshExpire time.Duration) (Handler, error) {
	if secret == "" {
		return nil, errors.New("jwt secret cannot be empty")
	}

	if accessExpire <= 0 {
		return nil, errors.New("jwt access expire must be positive")
	}

	if refreshExpire <= 0 {
		return nil, errors.New("jwt refresh expire must be positive")
	}

	return &HandlerImpl{
		Secret:        []byte(secret),
		AccessExpire:  accessExpire,
		RefreshExpire: refreshExpire,
		Leeway:        defaultLeeway,
	}, nil
}

func (h *HandlerImpl) SetAccessToken(claimsParams ClaimsParams) (string, error) {
	claimsParams.TokenType = TokenTypeAccess
	return h.signToken(claimsParams, h.AccessExpire)
}

func (h *HandlerImpl) SetRefreshToken(claimsParams ClaimsParams) (string, error) {
	claimsParams.TokenType = TokenTypeRefresh
	return h.signToken(claimsParams, h.RefreshExpire)
}

func (h *HandlerImpl) signToken(claimsParams ClaimsParams, expire time.Duration) (string, error) {
	now := time.Now()

	claims := Claims{
		ClaimsParams: claimsParams,
		RegisteredClaims: jwtv5.RegisteredClaims{
			IssuedAt:  jwtv5.NewNumericDate(now),
			NotBefore: jwtv5.NewNumericDate(now),
			ExpiresAt: jwtv5.NewNumericDate(now.Add(expire)),
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
