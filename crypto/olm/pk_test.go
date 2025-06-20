// Copyright (c) 2024 Sumner Evans
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Only run this test if goolm is disabled (that is, libolm is used).

package olm_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iKonoTelecomunicaciones/go/crypto/goolm/pk"
	"github.com/iKonoTelecomunicaciones/go/crypto/libolm"
)

func FuzzSign(f *testing.F) {
	seed := []byte("Quohboh3ka3ooghequier9lee8Bahwoh")
	goolmPkSigning, err := pk.NewSigningFromSeed(seed)
	require.NoError(f, err)

	libolmPkSigning, err := libolm.NewPKSigningFromSeed(seed)
	require.NoError(f, err)

	f.Add([]byte("message"))

	f.Fuzz(func(t *testing.T, message []byte) {
		// libolm breaks with empty messages, so don't perform differential
		// fuzzing on that.
		if len(message) == 0 {
			return
		}

		libolmResult, libolmErr := libolmPkSigning.Sign(message)
		goolmResult, goolmErr := goolmPkSigning.Sign(message)

		assert.Equal(t, goolmErr, libolmErr)
		assert.Equal(t, goolmResult, libolmResult)
	})
}
