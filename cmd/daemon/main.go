package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/ingcr3at1on/x/sigctx"
	flag "github.com/spf13/pflag"
	"justanother.org/labdns/api/dns"
	"justanother.org/labdns/api/dns/cloudflare"
	"justanother.org/labdns/api/dns/digitalocean"
	"justanother.org/labdns/api/ipcheck"
	"justanother.org/labdns/api/ipcheck/ipify"
)

var (
	currentIP   string
	accessToken *string
	domain      *string
	subname     *string
	provider    *string
	email       *string
)

func init() {
	accessToken = flag.StringP(`token`, `t`, ``, `Provider access token`)
	domain = flag.StringP(`domain`, `d`, ``, `Provider managed domain`)
	subname = flag.StringP(`subname`, `s`, ``, `Domain subname`)
	provider = flag.StringP(`provider`, `p`, ``, `Provider name`)
	email = flag.StringP(`email`, `e`, ``, `Provider user email`)

	flag.Parse()
}

func main() {
	if err := sigctx.StartWith(func(ctx context.Context) error {
		if err := checkAndUpdate(ctx); err != nil {
			return err
		}

		t := time.NewTicker(24 * time.Hour)
		for {
			<-t.C
			if err := checkAndUpdate(ctx); err != nil {
				return err
			}
		}
	}); err != nil {
		log.Fatal(err)
	}
}

func checkAndUpdate(ctx context.Context) error {
	c, err := getChecker()
	if err != nil {
		return err
	}

	ip, err := doCheck(c)
	if err != nil {
		return err
	}

	// If the returned IP matches our current "known" IP, don't bother
	// calling update at the DNS provider.
	// TODO: consider checking if the provider has the correct IP and only
	// setting it in the case it doesn't match as an alternative.
	if ip == currentIP {
		return nil
	}
	currentIP = ip

	p, err := getProvider(ctx)
	if err != nil {
		return err
	}

	return doUpdate(ctx, p, currentIP)
}

func getChecker() (ipcheck.Checker, error) {
	return ipify.New()
}

func doCheck(c ipcheck.Checker) (string, error) {
	return c.GetIP(context.Background())
}

func getProvider(ctx context.Context) (dns.Provider, error) {
	switch *provider {
	case `cloudflare`:
		return cloudflare.New(&cloudflare.Config{
			Domain: *domain,
			APIKey: *accessToken,
			Email:  *email,
			// FIXME: expose `proxy` value
		})
	case `digitalocean`, `do`:
		return digitalocean.New(ctx, &digitalocean.Config{
			Domain:      *domain,
			AccessToken: *accessToken,
		})
	default:
		return nil, errors.New("unrecognized provider string")
	}
}

func doUpdate(ctx context.Context, p dns.Provider, ip string) error {
	return p.UpdateRecord(ctx, *subname, ip)
}
