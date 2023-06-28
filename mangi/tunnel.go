package mangi

import (
	"github.com/multiformats/go-multiaddr"
)

var Tunnel multiaddr.Protocol = multiaddr.Protocol{
	Name: "tunnel",
	Code: 1099,
	Size: -1,
	Transcoder: multiaddr.NewTranscoderFromFunctions(func(s string) ([]byte, error) {
		d := Demangle(s, '/')
		x, err := multiaddr.NewMultiaddr(d)
		if err != nil {
			return []byte{}, err
		}
		return x.Bytes(), nil
	}, func(b []byte) (string, error) {
		a, err := multiaddr.NewMultiaddrBytes(b)
		if err != nil {
			return "", err
		}
		return Mangle(a.String(), '/'), nil
	}, func(b []byte) error {
		_, err := multiaddr.NewMultiaddrBytes(b)
		return err
	}),
}

func init() {
	multiaddr.AddProtocol(Tunnel)
}
