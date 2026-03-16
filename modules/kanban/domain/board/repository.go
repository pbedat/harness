package board

import "context"

type Repository interface {
	Create(ctx context.Context, board *Board) error
	Get(ctx context.Context, id string) (*Board, error)
	Update(ctx context.Context, id string, board *Board) error
}
