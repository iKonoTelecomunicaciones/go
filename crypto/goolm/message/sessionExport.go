package message

import (
	"encoding/binary"

	"codeberg.org/DerLukas/goolm"
	"codeberg.org/DerLukas/goolm/crypto"
	"github.com/pkg/errors"
)

const (
	sessionExportVersion = 0x01
)

// MegolmSessionExport represents a message in the session export format.
type MegolmSessionExport struct {
	Counter     uint32                  `json:"counter"`
	RatchetData [128]byte               `json:"data"`
	PublicKey   crypto.Ed25519PublicKey `json:"kPub"`
}

// Encode returns the encoded message in the correct format.
func (s MegolmSessionExport) Encode() []byte {
	output := make([]byte, 165)
	output[0] = sessionExportVersion
	binary.BigEndian.PutUint32(output[1:], s.Counter)
	copy(output[5:], s.RatchetData[:])
	copy(output[133:], s.PublicKey)
	return output
}

// Decode populates the struct with the data encoded in input.
func (s *MegolmSessionExport) Decode(input []byte) error {
	if len(input) != 165 {
		return errors.Wrap(goolm.ErrBadInput, "decrypt")
	}
	if input[0] != sessionExportVersion {
		return errors.Wrap(goolm.ErrBadVersion, "decrypt")
	}
	s.Counter = binary.BigEndian.Uint32(input[1:5])
	copy(s.RatchetData[:], input[5:133])
	s.PublicKey = input[133:]
	return nil
}
