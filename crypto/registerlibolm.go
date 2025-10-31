//go:build !goolm

package crypto

import "github.com/iKonoTelecomunicaciones/go/crypto/libolm"

func init() {
	libolm.Register()
}
