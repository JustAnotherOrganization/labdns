package digitalocean

import (
	"context"
	"errors"
	"sync"

	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
	"justanother.org/labdns/api/dns"
)

// FIXME: consider monitoring DO responses for API backoff.
// I don't actually remember if this expects the name to be the full name or partial, but I think it's partial..
// if not it should be updated to match the cloudflare functionality lol...

// Config is the provider configuration.
type Config struct {
	// Domain is the name of the domain to update.
	Domain string
	// AccessToken is a digital ocean Personal Access Token (PAT).
	AccessToken string
}

// Token returns an oauth2.Token.
func (c *Config) Token() (*oauth2.Token, error) {
	return &oauth2.Token{
		AccessToken: c.AccessToken,
	}, nil
}

func (c *Config) validate() error {
	if c == nil {
		return errors.New("provider config cannot be nil")
	}

	if c.AccessToken == `` {
		return errors.New("provider access token cannot be empty")
	}

	if c.Domain == `` {
		return errors.New("domain cannot be empty")
	}

	return nil
}

type _digitalocean struct {
	*Config
	godo.DomainsService

	mux       sync.Mutex
	recordIDs map[string]int
}

// New returns a new dns.Provider.
func New(ctx context.Context, cfg *Config) (dns.Provider, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	client := godo.NewClient(oauth2.NewClient(ctx, cfg))

	return &_digitalocean{
		Config:         cfg,
		DomainsService: client.Domains,
		recordIDs:      make(map[string]int),
	}, nil
}

// expects mux to already be locked!
func (do *_digitalocean) lookupRecordID(ctx context.Context, name string) (int, error) {
	rs, _, err := do.Records(ctx, do.Domain, nil)
	if err != nil {
		return 0, err
	}

	for _, r := range rs {
		if r.Name == name {
			return r.ID, nil
		}
	}

	return 0, nil
}

// UpdateRecord updates the given DNS record.
func (do *_digitalocean) UpdateRecord(ctx context.Context, name, ip string) error {
	if name == `` {
		return errors.New(`domain subname cannot be empty`)
	}

	if ip == `` {
		return errors.New(`ip address cannot be empty`)
	}

	do.mux.Lock()
	defer do.mux.Unlock()

	id, ok := do.recordIDs[name]
	if !ok {
		var err error // don't shadow id
		id, err = do.lookupRecordID(ctx, name)
		if err != nil {
			return err
		}
	}

	_, _, err := do.DomainsService.EditRecord(ctx, do.Domain, id, &godo.DomainRecordEditRequest{
		Type: `A`,
		Name: name,
		Data: ip,
	})
	if err != nil {
		return err
	}

	return nil
}
