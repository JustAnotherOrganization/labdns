package cloudflare

import (
	"context"
	"errors"
	"sync"

	"github.com/cloudflare/cloudflare-go"
	"justanother.org/labdns/api/dns"
)

type Config struct {
	Domain  string
	Email   string
	APIKey  string
	Proxied bool
}

func (c *Config) validate() error {
	if c == nil {
		return errors.New("provider config cannot be nil")
	}

	return nil
}

type _cloudflare struct {
	*Config
	*cloudflare.API

	zoneID string

	mux       sync.Mutex
	recordIDs map[string]string
}

func New(cfg *Config) (dns.Provider, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	api, err := cloudflare.New(cfg.APIKey, cfg.Email)
	if err != nil {
		return nil, err
	}

	zone, err := api.ZoneIDByName(cfg.Domain)
	if err != nil {
		return nil, err
	}

	return &_cloudflare{
		Config:    cfg,
		API:       api,
		zoneID:    zone,
		recordIDs: make(map[string]string),
	}, nil
}

func (p *_cloudflare) lookupRecordID(name string) (string, error) {
	rs, err := p.DNSRecords(p.zoneID, cloudflare.DNSRecord{})
	if err != nil {
		return ``, err
	}

	for _, r := range rs {
		if r.Name == name {
			return r.ID, nil
		}
	}

	return ``, nil
}

func (p *_cloudflare) UpdateRecord(ctx context.Context, name, ip string) error {
	if name == `` {
		return errors.New(`domain subname cannot be empty`)
	}

	if ip == `` {
		return errors.New(`ip address cannot be empty`)
	}

	p.mux.Lock()
	defer p.mux.Unlock()

	_name := name + `.` + p.Domain

	id, ok := p.recordIDs[name]
	if !ok {
		var err error // don't shadow id
		id, err = p.lookupRecordID(_name)
		if err != nil {
			return err
		}
	}

	err := p.UpdateDNSRecord(p.zoneID, id, cloudflare.DNSRecord{
		Type:    `A`,
		Name:    _name,
		Content: ip,
		Proxied: p.Proxied,
	})
	if err != nil {
		return err
	}

	return nil
}
