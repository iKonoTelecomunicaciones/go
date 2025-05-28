// Copyright (c) 2024 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package federation

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"go.mau.fi/util/exgjson"
	"go.mau.fi/util/jsontime"

	"maunium.net/go/mautrix/crypto/canonicaljson"
	"maunium.net/go/mautrix/id"
)

// SigningKey is a Matrix federation signing key pair.
type SigningKey struct {
	ID   id.KeyID
	Pub  id.SigningKey
	Priv ed25519.PrivateKey
}

// SynapseString returns a string representation of the private key compatible with Synapse's .signing.key file format.
//
// The output of this function can be parsed back into a [SigningKey] using the [ParseSynapseKey] function.
func (sk *SigningKey) SynapseString() string {
	alg, id := sk.ID.Parse()
	return fmt.Sprintf("%s %s %s", alg, id, base64.RawStdEncoding.EncodeToString(sk.Priv.Seed()))
}

// ParseSynapseKey parses a Synapse-compatible private key string into a SigningKey.
func ParseSynapseKey(key string) (*SigningKey, error) {
	parts := strings.Split(key, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid key format (expected 3 space-separated parts, got %d)", len(parts))
	} else if parts[0] != string(id.KeyAlgorithmEd25519) {
		return nil, fmt.Errorf("unsupported key algorithm %s (only ed25519 is supported)", parts[0])
	}
	seed, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}
	priv := ed25519.NewKeyFromSeed(seed)
	pub := base64.RawStdEncoding.EncodeToString(priv.Public().(ed25519.PublicKey))
	return &SigningKey{
		ID:   id.NewKeyID(id.KeyAlgorithmEd25519, parts[1]),
		Pub:  id.SigningKey(pub),
		Priv: priv,
	}, nil
}

// GenerateSigningKey generates a new random signing key.
func GenerateSigningKey() *SigningKey {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}
	return &SigningKey{
		ID:   id.NewKeyID(id.KeyAlgorithmEd25519, base64.RawURLEncoding.EncodeToString(pub[:4])),
		Pub:  id.SigningKey(base64.RawStdEncoding.EncodeToString(pub)),
		Priv: priv,
	}
}

// ServerKeyResponse is the response body for the `GET /_matrix/key/v2/server` endpoint.
// It's also used inside the query endpoint response structs.
type ServerKeyResponse struct {
	ServerName    string                         `json:"server_name"`
	VerifyKeys    map[id.KeyID]ServerVerifyKey   `json:"verify_keys"`
	OldVerifyKeys map[id.KeyID]OldVerifyKey      `json:"old_verify_keys,omitempty"`
	Signatures    map[string]map[id.KeyID]string `json:"signatures,omitempty"`
	ValidUntilTS  jsontime.UnixMilli             `json:"valid_until_ts"`

	Extra map[string]any `json:"-"`
}

func (skr *ServerKeyResponse) HasKey(keyID id.KeyID) bool {
	if skr == nil {
		return false
	} else if _, ok := skr.VerifyKeys[keyID]; ok {
		return true
	}
	return false
}

func (skr *ServerKeyResponse) VerifySelfSignature() bool {
	for keyID, key := range skr.VerifyKeys {
		if !VerifyJSON(skr.ServerName, keyID, key.Key, skr) {
			return false
		}
	}
	return true
}

func VerifyJSON(serverName string, keyID id.KeyID, key id.SigningKey, data any) bool {
	var err error
	message, ok := data.(json.RawMessage)
	if !ok {
		message, err = json.Marshal(data)
		if err != nil {
			return false
		}
	}
	sigVal := gjson.GetBytes(message, exgjson.Path("signatures", serverName, string(keyID)))
	if sigVal.Type != gjson.String {
		return false
	}
	message, err = sjson.DeleteBytes(message, "signatures")
	if err != nil {
		return false
	}
	message, err = sjson.DeleteBytes(message, "unsigned")
	if err != nil {
		return false
	}
	return VerifyJSONRaw(key, sigVal.Str, message)
}

func VerifyJSONRaw(key id.SigningKey, sig string, message json.RawMessage) bool {
	sigBytes, err := base64.RawURLEncoding.DecodeString(sig)
	if err != nil {
		return false
	}
	keyBytes, err := base64.RawStdEncoding.DecodeString(string(key))
	if err != nil {
		return false
	}
	message = canonicaljson.CanonicalJSONAssumeValid(message)
	return ed25519.Verify(keyBytes, message, sigBytes)
}

type marshalableSKR ServerKeyResponse

func (skr *ServerKeyResponse) MarshalJSON() ([]byte, error) {
	if skr.Extra == nil {
		return json.Marshal((*marshalableSKR)(skr))
	}
	marshalable := maps.Clone(skr.Extra)
	marshalable["server_name"] = skr.ServerName
	marshalable["verify_keys"] = skr.VerifyKeys
	marshalable["old_verify_keys"] = skr.OldVerifyKeys
	marshalable["signatures"] = skr.Signatures
	marshalable["valid_until_ts"] = skr.ValidUntilTS
	return json.Marshal(skr.Extra)
}

func (skr *ServerKeyResponse) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, (*marshalableSKR)(skr))
	if err != nil {
		return err
	}
	var extra map[string]any
	err = json.Unmarshal(data, &extra)
	if err != nil {
		return err
	}
	delete(extra, "server_name")
	delete(extra, "verify_keys")
	delete(extra, "old_verify_keys")
	delete(extra, "signatures")
	delete(extra, "valid_until_ts")
	if len(extra) > 0 {
		skr.Extra = extra
	} else {
		skr.Extra = nil
	}
	return nil
}

type ServerVerifyKey struct {
	Key id.SigningKey `json:"key"`
}

func (svk *ServerVerifyKey) Decode() (ed25519.PublicKey, error) {
	return base64.RawStdEncoding.DecodeString(string(svk.Key))
}

type OldVerifyKey struct {
	Key       id.SigningKey      `json:"key"`
	ExpiredTS jsontime.UnixMilli `json:"expired_ts"`
}

func (sk *SigningKey) SignJSON(data any) (string, error) {
	marshaled, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	marshaled, err = sjson.DeleteBytes(marshaled, "signatures")
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(sk.SignRawJSON(marshaled)), nil
}

func (sk *SigningKey) SignRawJSON(data json.RawMessage) []byte {
	return ed25519.Sign(sk.Priv, canonicaljson.CanonicalJSONAssumeValid(data))
}

// GenerateKeyResponse generates a key response signed by this key with the given server name and optionally some old verify keys.
func (sk *SigningKey) GenerateKeyResponse(serverName string, oldVerifyKeys map[id.KeyID]OldVerifyKey) *ServerKeyResponse {
	skr := &ServerKeyResponse{
		ServerName:    serverName,
		OldVerifyKeys: oldVerifyKeys,
		ValidUntilTS:  jsontime.UM(time.Now().Add(24 * time.Hour)),
		VerifyKeys: map[id.KeyID]ServerVerifyKey{
			sk.ID: {Key: sk.Pub},
		},
	}
	signature, err := sk.SignJSON(skr)
	if err != nil {
		panic(err)
	}
	skr.Signatures = map[string]map[id.KeyID]string{
		serverName: {
			sk.ID: signature,
		},
	}
	return skr
}
