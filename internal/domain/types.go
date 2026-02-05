package domain

import (
    "errors"
    "strings"
    "time"

    "github.com/google/uuid"
)

type GameStatus string

const (
    GameWaiting    GameStatus = "waiting"
    GameInProgress GameStatus = "in_progress"
    GameFinished   GameStatus = "finished"
    GameDrawOffer  GameStatus = "draw_offered"
)

type User struct {
    ID           uuid.UUID
    Username     string
    PasswordHash string
    CreatedAt    time.Time
}

type Game struct {
    ID            uuid.UUID
    PlayerX       uuid.UUID
    PlayerO       uuid.UUID
    Board         [9]rune
    NextTurn      string
    Status        GameStatus
    WinnerUserID  *uuid.UUID
    DrawOfferedBy *uuid.UUID
    CreatedAt     time.Time
    UpdatedAt     time.Time
}

func NewEmptyBoard() [9]rune {
    return [9]rune{'.', '.', '.', '.', '.', '.', '.', '.', '.'}
}

func BoardToString(b [9]rune) string {
    var sb strings.Builder
    sb.Grow(9)
    for _, r := range b {
        if r == 0 {
            sb.WriteRune('.')
            continue
        }
        sb.WriteRune(r)
    }
    return sb.String()
}

func StringToBoard(s string) ([9]rune, error) {
    if len(s) != 9 {
        return [9]rune{}, errors.New("board string must be 9 chars")
    }
    var b [9]rune
    for i, ch := range s {
        if ch == 'X' || ch == 'O' || ch == '.' {
            b[i] = ch
        } else {
            return [9]rune{}, errors.New("invalid board char")
        }
    }
    return b, nil
}

type GameMove struct {
    ID        int64
    GameID    uuid.UUID
    UserID    uuid.UUID
    Position  int
    Symbol    string
    CreatedAt time.Time
}

type GameMessage struct {
    ID        int64
    GameID    uuid.UUID
    UserID    uuid.UUID
    Message   string
    CreatedAt time.Time
}
