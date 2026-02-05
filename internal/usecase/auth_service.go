package usecase

import (
    "context"
    "strings"
    "time"

    "github.com/google/uuid"
    "golang.org/x/crypto/bcrypt"
    "xo-server/internal/domain"
)

type authService struct {
    users  UserRepository
    tokens TokenProvider
}

func NewAuthService(users UserRepository, tokens TokenProvider) AuthService {
    return &authService{users: users, tokens: tokens}
}

func (s *authService) Register(ctx context.Context, username, password string) (*domain.User, error) {
    username = strings.TrimSpace(username)
    if username == "" || len(password) < 6 {
        return nil, domain.ErrInvalidInput
    }

    _, err := s.users.GetUserByUsername(ctx, username)
    if err == nil {
        return nil, domain.ErrInvalidInput
    }

    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }

    user := &domain.User{
        ID:           uuid.New(),
        Username:     username,
        PasswordHash: string(hash),
        CreatedAt:    time.Now().UTC(),
    }

    if err := s.users.CreateUser(ctx, user); err != nil {
        return nil, err
    }

    return user, nil
}

func (s *authService) Login(ctx context.Context, username, password string) (string, *domain.User, error) {
    username = strings.TrimSpace(username)
    if username == "" || password == "" {
        return "", nil, domain.ErrInvalidInput
    }

    user, err := s.users.GetUserByUsername(ctx, username)
    if err != nil {
        return "", nil, domain.ErrUnauthorized
    }

    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
        return "", nil, domain.ErrUnauthorized
    }

    token, err := s.tokens.IssueToken(user)
    if err != nil {
        return "", nil, err
    }

    return token, user, nil
}
