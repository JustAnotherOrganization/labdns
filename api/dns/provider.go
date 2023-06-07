package dns

import "context"

type Provider interface {
	UpdateRecord(ctx context.Context, name, ip string) error
}
