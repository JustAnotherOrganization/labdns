package dns

import "context"

// Provider represents a DNS provider.
type Provider interface {
	UpdateRecord(ctx context.Context, name, ip string) error
}
