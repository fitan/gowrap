package nest

import "context"

type Nest interface {
	SayNest(ctx context.Context, req string) (string, error)
}