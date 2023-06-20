package ipcheck

import "context"

type Checker interface {
	GetIP(ctx context.Context) (string, error)
}
