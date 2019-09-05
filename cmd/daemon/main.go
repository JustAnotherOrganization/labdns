package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ingcr3at1on/x/sigctx"
	flag "github.com/spf13/pflag"
	"justanother.org/labdns/api/dns"
	"justanother.org/labdns/api/dns/cloudflare"
	do "justanother.org/labdns/api/dns/digitalocean"
	"justanother.org/labdns/api/ipcheck"
	"justanother.org/labdns/api/ipcheck/ipify"
	"justanother.org/labdns/cmd/internal"
)

// FIXME: consider using the heap less...
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

	// TODO: setLogger using flags
	setLogger(new(_logger))
}

func setLogger(logger internal.Logger) {
	sigctx.SetLogger(logger)
	internal.SetLogger(logger)
}

func main() {
	ctx := sigctx.FromContext(context.Background())
	internal.Fatal(sigctx.StartWithContext(ctx, func(ctx context.Context) error {
		return func() error {
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
		}()
	}))
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
	// More ip checkers would go here.
	return ipify.New()
}

func doCheck(c ipcheck.Checker) (string, error) {
	return c.GetIP()
}

func getProvider(ctx context.Context) (dns.Provider, error) {
	switch *provider {
	case `cloudflare`:
		return cloudflare.New(ctx, &cloudflare.Config{
			Domain: *domain,
			APIKey: *accessToken,
			Email:  *email,
		})
	case `digitalocean`, `do`:
		return do.New(ctx, &do.Config{
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

type _logger struct{}

func (_logger) Log(v ...interface{}) error {
	_, err := fmt.Println(v...)
	return err
}
