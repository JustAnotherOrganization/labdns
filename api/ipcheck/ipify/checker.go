package ipify

import (
	"github.com/justanotherorganization/labdns/api/ipcheck"
	"github.com/rdegges/go-ipify"
)

type _ipify struct{}

// New returns a ipcheck.Checker
func New() (ipcheck.Checker, error) {
	return new(_ipify), nil
}

// GetIP returns the IP address as seen by ipify.
func (_ipify) GetIP() (string, error) {
	return ipify.GetIp()
}
