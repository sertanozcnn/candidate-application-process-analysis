package application

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"unicode/utf8"
)

var (
	ErrInvalidApplication  = errors.New("invalid application")
	ErrApplicationExists   = errors.New("application already exists")
	ErrApplicationNotFound = errors.New("application not found")
)

var allowedPositions = map[string]struct{}{
	"frontend":  {},
	"backend":   {},
	"fullstack": {},
}

type store interface {
	Create(ctx context.Context, input CreateInput) (Application, error)
	List(ctx context.Context, limit, offset int) ([]Application, error)
	FindByID(ctx context.Context, id string) (Application, error)
}

type Service struct {
	store store
}

func NewService(store store) Service {
	return Service{store: store}
}

func (s Service) Create(ctx context.Context, input CreateInput) (Application, error) {
	input, err := normalizeAndValidate(input)
	if err != nil {
		return Application{}, err
	}
	return s.store.Create(ctx, input)
}

func (s Service) List(ctx context.Context, limit, offset int) ([]Application, error) {
	if limit < 1 || limit > 100 || offset < 0 {
		return nil, ErrInvalidApplication
	}
	return s.store.List(ctx, limit, offset)
}

func (s Service) FindByID(ctx context.Context, id string) (Application, error) {
	if strings.TrimSpace(id) == "" {
		return Application{}, ErrApplicationNotFound
	}
	return s.store.FindByID(ctx, id)
}

func normalizeAndValidate(input CreateInput) (CreateInput, error) {
	input.FullName = strings.Join(strings.Fields(input.FullName), " ")
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.PositionCode = strings.ToLower(strings.TrimSpace(input.PositionCode))
	input.Experience = strings.TrimSpace(input.Experience)

	if runeLength(input.FullName) < 2 || runeLength(input.FullName) > 120 {
		return CreateInput{}, ErrInvalidApplication
	}
	if len(input.Email) > 320 {
		return CreateInput{}, ErrInvalidApplication
	}
	parsed, err := mail.ParseAddress(input.Email)
	if err != nil || parsed.Address != input.Email {
		return CreateInput{}, ErrInvalidApplication
	}
	if _, ok := allowedPositions[input.PositionCode]; !ok {
		return CreateInput{}, ErrInvalidApplication
	}
	if runeLength(input.Experience) < 20 || runeLength(input.Experience) > 5000 {
		return CreateInput{}, ErrInvalidApplication
	}
	return input, nil
}

func runeLength(value string) int {
	return utf8.RuneCountInString(value)
}
