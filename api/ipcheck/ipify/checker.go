package ipify

import (
	"github.com/rdegges/go-ipify"
	"justanother.org/labdns/api/ipcheck"
)

type _ipify struct{}

func New() (ipcheck.Checker, error) {
	return new(_ipify), nil
}

func (_ipify) GetIP() (string, error) {
	return ipify.GetIp()
}
