package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type jwtService struct {
	secret    []byte
	issuer    string
	expiresIn time.Duration
}

type userClaims struct {
	UserID   int64    `json:"user_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

func newJWTService(cfg appConfig) *jwtService {
	return &jwtService{
		secret:    []byte(cfg.JWT.Secret),
		issuer:    cfg.JWT.Issuer,
		expiresIn: cfg.jwtExpiresIn(),
	}
}

func (s *jwtService) issueToken(user appPrincipal) (string, error) {
	if s == nil {
		return "", errors.New("jwt service is nil")
	}
	now := time.Now()
	claims := userClaims{
		UserID:   user.UserID,
		Username: user.Username,
		Roles:    user.Roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   strconv.FormatInt(user.UserID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.expiresIn)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *jwtService) parseToken(rawToken string) (appPrincipal, error) {
	if s == nil {
		return appPrincipal{}, errors.New("jwt service is nil")
	}
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return appPrincipal{}, errors.New("empty token")
	}

	claims := &userClaims{}
	token, err := jwt.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
		}
		return s.secret, nil
	})
	if err != nil {
		return appPrincipal{}, err
	}
	if token == nil || !token.Valid {
		return appPrincipal{}, errors.New("invalid token")
	}

	return appPrincipal{
		UserID:   claims.UserID,
		Username: claims.Username,
		Roles:    claims.Roles,
	}, nil
}
