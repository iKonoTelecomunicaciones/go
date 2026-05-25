//go:build goolm

package crypto

import (
	"github.com/iKonoTelecomunicaciones/go/crypto/goolm"
)

func init() {
	goolm.Register()
}
