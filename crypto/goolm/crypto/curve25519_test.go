package crypto_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/iKonoTelecomunicaciones/go/crypto/goolm/crypto"
	"github.com/iKonoTelecomunicaciones/go/crypto/goolm/libolmpickle"
)

const curve25519KeyPairPickleLength = crypto.Curve25519PublicKeyLength + // Public Key
	crypto.Curve25519PrivateKeyLength // Private Key

func TestCurve25519(t *testing.T) {
	firstKeypair, err := crypto.Curve25519GenerateKey()
	assert.NoError(t, err)
	secondKeypair, err := crypto.Curve25519GenerateKey()
	assert.NoError(t, err)
	sharedSecretFromFirst, err := firstKeypair.SharedSecret(secondKeypair.PublicKey)
	assert.NoError(t, err)
	sharedSecretFromSecond, err := secondKeypair.SharedSecret(firstKeypair.PublicKey)
	assert.NoError(t, err)
	assert.Equal(t, sharedSecretFromFirst, sharedSecretFromSecond, "shared secret not equal")
	fromPrivate, err := crypto.Curve25519GenerateFromPrivate(firstKeypair.PrivateKey)
	assert.NoError(t, err)
	assert.Equal(t, fromPrivate, firstKeypair)
}

func TestCurve25519Case1(t *testing.T) {
	alicePrivate := []byte{
		0x77, 0x07, 0x6D, 0x0A, 0x73, 0x18, 0xA5, 0x7D,
		0x3C, 0x16, 0xC1, 0x72, 0x51, 0xB2, 0x66, 0x45,
		0xDF, 0x4C, 0x2F, 0x87, 0xEB, 0xC0, 0x99, 0x2A,
		0xB1, 0x77, 0xFB, 0xA5, 0x1D, 0xB9, 0x2C, 0x2A,
	}
	alicePublic := []byte{
		0x85, 0x20, 0xF0, 0x09, 0x89, 0x30, 0xA7, 0x54,
		0x74, 0x8B, 0x7D, 0xDC, 0xB4, 0x3E, 0xF7, 0x5A,
		0x0D, 0xBF, 0x3A, 0x0D, 0x26, 0x38, 0x1A, 0xF4,
		0xEB, 0xA4, 0xA9, 0x8E, 0xAA, 0x9B, 0x4E, 0x6A,
	}
	bobPrivate := []byte{
		0x5D, 0xAB, 0x08, 0x7E, 0x62, 0x4A, 0x8A, 0x4B,
		0x79, 0xE1, 0x7F, 0x8B, 0x83, 0x80, 0x0E, 0xE6,
		0x6F, 0x3B, 0xB1, 0x29, 0x26, 0x18, 0xB6, 0xFD,
		0x1C, 0x2F, 0x8B, 0x27, 0xFF, 0x88, 0xE0, 0xEB,
	}
	bobPublic := []byte{
		0xDE, 0x9E, 0xDB, 0x7D, 0x7B, 0x7D, 0xC1, 0xB4,
		0xD3, 0x5B, 0x61, 0xC2, 0xEC, 0xE4, 0x35, 0x37,
		0x3F, 0x83, 0x43, 0xC8, 0x5B, 0x78, 0x67, 0x4D,
		0xAD, 0xFC, 0x7E, 0x14, 0x6F, 0x88, 0x2B, 0x4F,
	}
	expectedAgreement := []byte{
		0x4A, 0x5D, 0x9D, 0x5B, 0xA4, 0xCE, 0x2D, 0xE1,
		0x72, 0x8E, 0x3B, 0xF4, 0x80, 0x35, 0x0F, 0x25,
		0xE0, 0x7E, 0x21, 0xC9, 0x47, 0xD1, 0x9E, 0x33,
		0x76, 0xF0, 0x9B, 0x3C, 0x1E, 0x16, 0x17, 0x42,
	}
	aliceKeyPair := crypto.Curve25519KeyPair{
		PrivateKey: alicePrivate,
		PublicKey:  alicePublic,
	}
	bobKeyPair := crypto.Curve25519KeyPair{
		PrivateKey: bobPrivate,
		PublicKey:  bobPublic,
	}
	agreementFromAlice, err := aliceKeyPair.SharedSecret(bobKeyPair.PublicKey)
	assert.NoError(t, err)
	assert.Equal(t, expectedAgreement, agreementFromAlice, "expected agreement does not match agreement from Alice's view")
	agreementFromBob, err := bobKeyPair.SharedSecret(aliceKeyPair.PublicKey)
	assert.NoError(t, err)
	assert.Equal(t, expectedAgreement, agreementFromBob, "expected agreement does not match agreement from Bob's view")
}

func TestCurve25519Pickle(t *testing.T) {
	//create keypair
	keyPair, err := crypto.Curve25519GenerateKey()
	assert.NoError(t, err)

	encoder := libolmpickle.NewEncoder()
	keyPair.PickleLibOlm(encoder)
	assert.Len(t, encoder.Bytes(), curve25519KeyPairPickleLength)

	unpickledKeyPair := crypto.Curve25519KeyPair{}
	err = unpickledKeyPair.UnpickleLibOlm(libolmpickle.NewDecoder(encoder.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, keyPair, unpickledKeyPair)
}

func TestCurve25519PicklePubKeyOnly(t *testing.T) {
	//create keypair
	keyPair, err := crypto.Curve25519GenerateKey()
	assert.NoError(t, err)

	//Remove privateKey
	keyPair.PrivateKey = nil

	encoder := libolmpickle.NewEncoder()
	keyPair.PickleLibOlm(encoder)
	assert.Len(t, encoder.Bytes(), curve25519KeyPairPickleLength)

	unpickledKeyPair := crypto.Curve25519KeyPair{}
	err = unpickledKeyPair.UnpickleLibOlm(libolmpickle.NewDecoder(encoder.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, keyPair, unpickledKeyPair)
}

func TestCurve25519PicklePrivKeyOnly(t *testing.T) {
	//create keypair
	keyPair, err := crypto.Curve25519GenerateKey()
	assert.NoError(t, err)
	//Remove public
	keyPair.PublicKey = nil
	encoder := libolmpickle.NewEncoder()
	keyPair.PickleLibOlm(encoder)
	assert.Len(t, encoder.Bytes(), curve25519KeyPairPickleLength)
	unpickledKeyPair := crypto.Curve25519KeyPair{}
	err = unpickledKeyPair.UnpickleLibOlm(libolmpickle.NewDecoder(encoder.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, keyPair, unpickledKeyPair)
}
