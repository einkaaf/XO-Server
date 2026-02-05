package auth

import (
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"
    "xo-server/internal/domain"
)

type JWTProvider struct {
    secret []byte
    ttl    time.Duration
}

func NewJWTProvider(secret string, ttl time.Duration) *JWTProvider {
    return &JWTProvider{secret: []byte(secret), ttl: ttl}
}

type Claims struct {
    UserID   string `json:"uid"`
    Username string `json:"usr"`
    jwt.RegisteredClaims
}

func (p *JWTProvider) IssueToken(user *domain.User) (string, error) {
    now := time.Now().UTC()
    claims := Claims{
        UserID:   user.ID.String(),
        Username: user.Username,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(now.Add(p.ttl)),
            IssuedAt:  jwt.NewNumericDate(now),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(p.secret)
}

func (p *JWTProvider) ParseToken(tokenStr string) (*domain.User, error) {
    token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
        return p.secret, nil
    })
    if err != nil {
        return nil, err
    }

    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, domain.ErrUnauthorized
    }

    uid, err := uuid.Parse(claims.UserID)
    if err != nil {
        return nil, domain.ErrUnauthorized
    }

    return &domain.User{ID: uid, Username: claims.Username}, nil
}
