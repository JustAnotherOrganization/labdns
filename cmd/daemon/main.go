package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ingcr3at1on/x/sigctx"
	"github.com/justanotherorganization/labdns/api/dns"
	do "github.com/justanotherorganization/labdns/api/dns/digitalocean"
	"github.com/justanotherorganization/labdns/api/ipcheck"
	"github.com/justanotherorganization/labdns/api/ipcheck/ipify"
	"github.com/justanotherorganization/labdns/cmd/internal"
	flag "github.com/spf13/pflag"
)

// FIXME: consider using the heap less...
var (
	currentIP   string
	accessToken *string
	domain      *string
	subname     *string
)

func init() {
	accessToken = flag.StringP(`token`, `t`, ``, `Provider access token`)
	domain = flag.StringP(`domain`, `d`, ``, `Provider managed domain`)
	subname = flag.StringP(`subname`, `s`, ``, `Domain subname`)

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
	internal.Fatal(sigctx.StartWithContext(ctx, func() error {
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
	// More providers would go here.
	return do.New(ctx, &do.Config{
		Domain:      *domain,
		AccessToken: *accessToken,
	})
}

func doUpdate(ctx context.Context, p dns.Provider, ip string) error {
	return p.UpdateRecord(ctx, *subname, ip)
}

type _logger struct{}

func (_logger) Log(v ...interface{}) error {
	_, err := fmt.Println(v...)
	return err
}
