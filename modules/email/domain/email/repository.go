package email

import "context"

type Repository interface {
	Create(ctx context.Context, email *Email) error
	Get(ctx context.Context, id string) (*Email, error)
	Update(ctx context.Context, id string, fn func(*Email) error) error
}
