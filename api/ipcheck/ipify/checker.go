package ipify

import (
	"context"

	"github.com/rdegges/go-ipify"
	"justanother.org/labdns/api/ipcheck"
)

type _ipify struct{}

func New() (ipcheck.Checker, error) {
	return new(_ipify), nil
}

func (_ipify) GetIP(ctx context.Context) (string, error) {
	return ipify.GetIp()
}
