package pk_test

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/iKonoTelecomunicaciones/go/crypto/goolm/crypto"
	"github.com/iKonoTelecomunicaciones/go/crypto/goolm/pk"
)

func TestEncryptionDecryption(t *testing.T) {
	alicePrivate := []byte{
		0x77, 0x07, 0x6D, 0x0A, 0x73, 0x18, 0xA5, 0x7D,
		0x3C, 0x16, 0xC1, 0x72, 0x51, 0xB2, 0x66, 0x45,
		0xDF, 0x4C, 0x2F, 0x87, 0xEB, 0xC0, 0x99, 0x2A,
		0xB1, 0x77, 0xFB, 0xA5, 0x1D, 0xB9, 0x2C, 0x2A,
	}
	alicePublic := []byte("hSDwCYkwp1R0i33ctD73Wg2/Og0mOBr066SpjqqbTmo")
	bobPrivate := []byte{
		0x5D, 0xAB, 0x08, 0x7E, 0x62, 0x4A, 0x8A, 0x4B,
		0x79, 0xE1, 0x7F, 0x8B, 0x83, 0x80, 0x0E, 0xE6,
		0x6F, 0x3B, 0xB1, 0x29, 0x26, 0x18, 0xB6, 0xFD,
		0x1C, 0x2F, 0x8B, 0x27, 0xFF, 0x88, 0xE0, 0xEB,
	}
	bobPublic := []byte("3p7bfXt9wbTTW2HC7OQ1Nz+DQ8hbeGdNrfx+FG+IK08")
	decryption, err := pk.NewDecryptionFromPrivate(alicePrivate)
	assert.NoError(t, err)
	assert.EqualValues(t, alicePublic, decryption.PublicKey(), "public key not correct")
	assert.EqualValues(t, alicePrivate, decryption.PrivateKey(), "private key not correct")

	encryption, err := pk.NewEncryption(decryption.PublicKey())
	assert.NoError(t, err)
	plaintext := []byte("This is a test")

	ciphertext, mac, err := encryption.Encrypt(plaintext, bobPrivate)
	assert.NoError(t, err)

	decrypted, err := decryption.Decrypt(bobPublic, mac, ciphertext)
	assert.NoError(t, err)
	assert.EqualValues(t, plaintext, decrypted, "message not equal")
}

func TestSigning(t *testing.T) {
	seed := []byte{
		0x77, 0x07, 0x6D, 0x0A, 0x73, 0x18, 0xA5, 0x7D,
		0x3C, 0x16, 0xC1, 0x72, 0x51, 0xB2, 0x66, 0x45,
		0xDF, 0x4C, 0x2F, 0x87, 0xEB, 0xC0, 0x99, 0x2A,
		0xB1, 0x77, 0xFB, 0xA5, 0x1D, 0xB9, 0x2C, 0x2A,
	}
	message := []byte("We hold these truths to be self-evident, that all men are created equal, that they are endowed by their Creator with certain unalienable Rights, that among these are Life, Liberty and the pursuit of Happiness.")
	signing, _ := pk.NewSigningFromSeed(seed)
	signature, err := signing.Sign(message)
	assert.NoError(t, err)
	signatureDecoded, err := base64.RawStdEncoding.DecodeString(string(signature))
	assert.NoError(t, err)
	pubKeyEncoded := signing.PublicKey()
	pubKeyDecoded, err := base64.RawStdEncoding.DecodeString(string(pubKeyEncoded))
	assert.NoError(t, err)
	pubKey := crypto.Ed25519PublicKey(pubKeyDecoded)

	verified := pubKey.Verify(message, signatureDecoded)
	assert.True(t, verified, "signature did not verify")

	copy(signatureDecoded[0:], []byte("m"))
	verified = pubKey.Verify(message, signatureDecoded)
	assert.False(t, verified, "signature verified with wrong message")
}

func TestDecryptionPickling(t *testing.T) {
	alicePrivate := []byte{
		0x77, 0x07, 0x6D, 0x0A, 0x73, 0x18, 0xA5, 0x7D,
		0x3C, 0x16, 0xC1, 0x72, 0x51, 0xB2, 0x66, 0x45,
		0xDF, 0x4C, 0x2F, 0x87, 0xEB, 0xC0, 0x99, 0x2A,
		0xB1, 0x77, 0xFB, 0xA5, 0x1D, 0xB9, 0x2C, 0x2A,
	}
	alicePublic := []byte("hSDwCYkwp1R0i33ctD73Wg2/Og0mOBr066SpjqqbTmo")
	decryption, err := pk.NewDecryptionFromPrivate(alicePrivate)
	assert.NoError(t, err)
	assert.EqualValues(t, alicePublic, decryption.PublicKey(), "public key not correct")
	assert.EqualValues(t, alicePrivate, decryption.PrivateKey(), "private key not correct")
	pickleKey := []byte("secret_key")
	expectedPickle := []byte("qx37WTQrjZLz5tId/uBX9B3/okqAbV1ofl9UnHKno1eipByCpXleAAlAZoJgYnCDOQZDQWzo3luTSfkF9pU1mOILCbbouubs6TVeDyPfgGD9i86J8irHjA")
	pickled, err := decryption.Pickle(pickleKey)
	assert.NoError(t, err)
	assert.EqualValues(t, expectedPickle, pickled, "pickle not as expected")

	newDecription, err := pk.NewDecryption()
	assert.NoError(t, err)
	err = newDecription.Unpickle(pickled, pickleKey)
	assert.NoError(t, err)
	assert.EqualValues(t, alicePublic, newDecription.PublicKey(), "public key not correct")
	assert.EqualValues(t, alicePrivate, newDecription.PrivateKey(), "private key not correct")
}
