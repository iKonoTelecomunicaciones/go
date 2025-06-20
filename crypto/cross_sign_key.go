// Copyright (c) 2020 Nikos Filippakis
// Copyright (c) 2023 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package crypto

import (
	"context"
	"fmt"

	"go.mau.fi/util/jsonbytes"

	mautrix "github.com/iKonoTelecomunicaciones/go"
	"github.com/iKonoTelecomunicaciones/go/crypto/olm"
	"github.com/iKonoTelecomunicaciones/go/crypto/signatures"
	"github.com/iKonoTelecomunicaciones/go/id"
)

// CrossSigningKeysCache holds the three cross-signing keys for the current user.
type CrossSigningKeysCache struct {
	MasterKey      olm.PKSigning
	SelfSigningKey olm.PKSigning
	UserSigningKey olm.PKSigning
}

func (cskc *CrossSigningKeysCache) PublicKeys() *CrossSigningPublicKeysCache {
	return &CrossSigningPublicKeysCache{
		MasterKey:      cskc.MasterKey.PublicKey(),
		SelfSigningKey: cskc.SelfSigningKey.PublicKey(),
		UserSigningKey: cskc.UserSigningKey.PublicKey(),
	}
}

type CrossSigningSeeds struct {
	MasterKey      jsonbytes.UnpaddedURLBytes `json:"m.cross_signing.master"`
	SelfSigningKey jsonbytes.UnpaddedURLBytes `json:"m.cross_signing.self_signing"`
	UserSigningKey jsonbytes.UnpaddedURLBytes `json:"m.cross_signing.user_signing"`
}

func (mach *OlmMachine) ExportCrossSigningKeys() CrossSigningSeeds {
	return CrossSigningSeeds{
		MasterKey:      mach.CrossSigningKeys.MasterKey.Seed(),
		SelfSigningKey: mach.CrossSigningKeys.SelfSigningKey.Seed(),
		UserSigningKey: mach.CrossSigningKeys.UserSigningKey.Seed(),
	}
}

func (mach *OlmMachine) ImportCrossSigningKeys(keys CrossSigningSeeds) (err error) {
	var keysCache CrossSigningKeysCache
	if keysCache.MasterKey, err = olm.NewPKSigningFromSeed(keys.MasterKey); err != nil {
		return
	}
	if keysCache.SelfSigningKey, err = olm.NewPKSigningFromSeed(keys.SelfSigningKey); err != nil {
		return
	}
	if keysCache.UserSigningKey, err = olm.NewPKSigningFromSeed(keys.UserSigningKey); err != nil {
		return
	}

	mach.Log.Debug().
		Str("master", keysCache.MasterKey.PublicKey().String()).
		Str("self_signing", keysCache.SelfSigningKey.PublicKey().String()).
		Str("user_signing", keysCache.UserSigningKey.PublicKey().String()).
		Msg("Imported own cross-signing keys")

	mach.CrossSigningKeys = &keysCache
	mach.crossSigningPubkeys = keysCache.PublicKeys()
	return
}

// GenerateCrossSigningKeys generates new cross-signing keys.
func (mach *OlmMachine) GenerateCrossSigningKeys() (*CrossSigningKeysCache, error) {
	var keysCache CrossSigningKeysCache
	var err error
	if keysCache.MasterKey, err = olm.NewPKSigning(); err != nil {
		return nil, fmt.Errorf("failed to generate master key: %w", err)
	}
	if keysCache.SelfSigningKey, err = olm.NewPKSigning(); err != nil {
		return nil, fmt.Errorf("failed to generate self-signing key: %w", err)
	}
	if keysCache.UserSigningKey, err = olm.NewPKSigning(); err != nil {
		return nil, fmt.Errorf("failed to generate user-signing key: %w", err)
	}
	mach.Log.Debug().
		Str("master", keysCache.MasterKey.PublicKey().String()).
		Str("self_signing", keysCache.SelfSigningKey.PublicKey().String()).
		Str("user_signing", keysCache.UserSigningKey.PublicKey().String()).
		Msg("Generated cross-signing keys")
	return &keysCache, nil
}

// PublishCrossSigningKeys signs and uploads the public keys of the given cross-signing keys to the server.
func (mach *OlmMachine) PublishCrossSigningKeys(ctx context.Context, keys *CrossSigningKeysCache, uiaCallback mautrix.UIACallback) error {
	userID := mach.Client.UserID
	masterKeyID := id.NewKeyID(id.KeyAlgorithmEd25519, keys.MasterKey.PublicKey().String())
	masterKey := mautrix.CrossSigningKeys{
		UserID: userID,
		Usage:  []id.CrossSigningUsage{id.XSUsageMaster},
		Keys: map[id.KeyID]id.Ed25519{
			masterKeyID: keys.MasterKey.PublicKey(),
		},
	}
	masterSig, err := mach.account.SignJSON(masterKey)
	if err != nil {
		return fmt.Errorf("failed to sign master key: %w", err)
	}
	masterKey.Signatures = signatures.NewSingleSignature(userID, id.KeyAlgorithmEd25519, mach.Client.DeviceID.String(), masterSig)

	selfKey := mautrix.CrossSigningKeys{
		UserID: userID,
		Usage:  []id.CrossSigningUsage{id.XSUsageSelfSigning},
		Keys: map[id.KeyID]id.Ed25519{
			id.NewKeyID(id.KeyAlgorithmEd25519, keys.SelfSigningKey.PublicKey().String()): keys.SelfSigningKey.PublicKey(),
		},
	}
	selfSig, err := keys.MasterKey.SignJSON(selfKey)
	if err != nil {
		return fmt.Errorf("failed to sign self-signing key: %w", err)
	}
	selfKey.Signatures = signatures.NewSingleSignature(userID, id.KeyAlgorithmEd25519, keys.MasterKey.PublicKey().String(), selfSig)

	userKey := mautrix.CrossSigningKeys{
		UserID: userID,
		Usage:  []id.CrossSigningUsage{id.XSUsageUserSigning},
		Keys: map[id.KeyID]id.Ed25519{
			id.NewKeyID(id.KeyAlgorithmEd25519, keys.UserSigningKey.PublicKey().String()): keys.UserSigningKey.PublicKey(),
		},
	}
	userSig, err := keys.MasterKey.SignJSON(userKey)
	if err != nil {
		return fmt.Errorf("failed to sign user-signing key: %w", err)
	}
	userKey.Signatures = signatures.NewSingleSignature(userID, id.KeyAlgorithmEd25519, keys.MasterKey.PublicKey().String(), userSig)

	err = mach.Client.UploadCrossSigningKeys(ctx, &mautrix.UploadCrossSigningKeysReq{
		Master:      masterKey,
		SelfSigning: selfKey,
		UserSigning: userKey,
	}, uiaCallback)
	if err != nil {
		return err
	}

	if err := mach.CryptoStore.PutSignature(ctx, userID, keys.MasterKey.PublicKey(), userID, mach.account.SigningKey(), masterSig); err != nil {
		return fmt.Errorf("error storing signature of master key by device signing key in crypto store: %w", err)
	}
	if err := mach.CryptoStore.PutSignature(ctx, userID, keys.SelfSigningKey.PublicKey(), userID, keys.MasterKey.PublicKey(), selfSig); err != nil {
		return fmt.Errorf("error storing signature of self-signing key by master key in crypto store: %w", err)
	}
	if err := mach.CryptoStore.PutSignature(ctx, userID, keys.UserSigningKey.PublicKey(), userID, keys.MasterKey.PublicKey(), userSig); err != nil {
		return fmt.Errorf("error storing signature of user-signing key by master key in crypto store: %w", err)
	}

	mach.CrossSigningKeys = keys
	mach.crossSigningPubkeys = keys.PublicKeys()

	return nil
}
