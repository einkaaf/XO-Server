package domain

import "errors"

var (
    ErrNotFound        = errors.New("not found")
    ErrInvalidInput    = errors.New("invalid input")
    ErrUnauthorized    = errors.New("unauthorized")
    ErrForbidden       = errors.New("forbidden")
    ErrGameNotActive   = errors.New("game not active")
    ErrNotYourTurn     = errors.New("not your turn")
    ErrPositionTaken   = errors.New("position taken")
    ErrInvalidPosition = errors.New("invalid position")
    ErrAlreadyInQueue  = errors.New("already in queue")
    ErrDrawNotOffered  = errors.New("draw not offered")
)
