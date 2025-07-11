package libolm

// #cgo LDFLAGS: -lolm -lstdc++
// #include <olm/olm.h>
import "C"

import (
	"crypto/rand"
	"encoding/base64"
	"unsafe"

	"github.com/iKonoTelecomunicaciones/go/crypto/olm"
	"github.com/iKonoTelecomunicaciones/go/id"
)

// OutboundGroupSession stores an outbound encrypted messaging session
// for a group.
type OutboundGroupSession struct {
	int *C.OlmOutboundGroupSession
	mem []byte
}

func init() {
	olm.InitNewOutboundGroupSessionFromPickled = func(pickled, key []byte) (olm.OutboundGroupSession, error) {
		if len(pickled) == 0 {
			return nil, olm.EmptyInput
		}
		s := NewBlankOutboundGroupSession()
		return s, s.Unpickle(pickled, key)
	}
	olm.InitNewOutboundGroupSession = func() (olm.OutboundGroupSession, error) { return NewOutboundGroupSession() }
	olm.InitNewBlankOutboundGroupSession = func() olm.OutboundGroupSession { return NewBlankOutboundGroupSession() }
}

// Ensure that [OutboundGroupSession] implements [olm.OutboundGroupSession].
var _ olm.OutboundGroupSession = (*OutboundGroupSession)(nil)

func NewOutboundGroupSession() (*OutboundGroupSession, error) {
	s := NewBlankOutboundGroupSession()
	random := make([]byte, s.createRandomLen()+1)
	_, err := rand.Read(random)
	if err != nil {
		return nil, err
	}
	r := C.olm_init_outbound_group_session(
		(*C.OlmOutboundGroupSession)(s.int),
		(*C.uint8_t)(&random[0]),
		C.size_t(len(random)))
	if r == errorVal() {
		return nil, s.lastError()
	}
	return s, nil
}

// outboundGroupSessionSize is the size of an outbound group session object in
// bytes.
func outboundGroupSessionSize() uint {
	return uint(C.olm_outbound_group_session_size())
}

// NewBlankOutboundGroupSession initialises an empty [OutboundGroupSession].
func NewBlankOutboundGroupSession() *OutboundGroupSession {
	memory := make([]byte, outboundGroupSessionSize())
	return &OutboundGroupSession{
		int: C.olm_outbound_group_session(unsafe.Pointer(&memory[0])),
		mem: memory,
	}
}

// lastError returns an error describing the most recent error to happen to an
// outbound group session.
func (s *OutboundGroupSession) lastError() error {
	return convertError(C.GoString(C.olm_outbound_group_session_last_error((*C.OlmOutboundGroupSession)(s.int))))
}

// Clear clears the memory used to back this OutboundGroupSession.
func (s *OutboundGroupSession) Clear() error {
	r := C.olm_clear_outbound_group_session((*C.OlmOutboundGroupSession)(s.int))
	if r == errorVal() {
		return s.lastError()
	} else {
		return nil
	}
}

// pickleLen returns the number of bytes needed to store an outbound group
// session.
func (s *OutboundGroupSession) pickleLen() uint {
	return uint(C.olm_pickle_outbound_group_session_length((*C.OlmOutboundGroupSession)(s.int)))
}

// Pickle returns an OutboundGroupSession as a base64 string.  Encrypts the
// OutboundGroupSession using the supplied key.
func (s *OutboundGroupSession) Pickle(key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, olm.NoKeyProvided
	}
	pickled := make([]byte, s.pickleLen())
	r := C.olm_pickle_outbound_group_session(
		(*C.OlmOutboundGroupSession)(s.int),
		unsafe.Pointer(&key[0]),
		C.size_t(len(key)),
		unsafe.Pointer(&pickled[0]),
		C.size_t(len(pickled)))
	if r == errorVal() {
		return nil, s.lastError()
	}
	return pickled[:r], nil
}

func (s *OutboundGroupSession) Unpickle(pickled, key []byte) error {
	if len(key) == 0 {
		return olm.NoKeyProvided
	}
	r := C.olm_unpickle_outbound_group_session(
		(*C.OlmOutboundGroupSession)(s.int),
		unsafe.Pointer(&key[0]),
		C.size_t(len(key)),
		unsafe.Pointer(&pickled[0]),
		C.size_t(len(pickled)))
	if r == errorVal() {
		return s.lastError()
	}
	return nil
}

// Deprecated
func (s *OutboundGroupSession) GobEncode() ([]byte, error) {
	pickled, err := s.Pickle(pickleKey)
	if err != nil {
		return nil, err
	}
	length := base64.RawStdEncoding.DecodedLen(len(pickled))
	rawPickled := make([]byte, length)
	_, err = base64.RawStdEncoding.Decode(rawPickled, pickled)
	return rawPickled, err
}

// Deprecated
func (s *OutboundGroupSession) GobDecode(rawPickled []byte) error {
	if s == nil || s.int == nil {
		*s = *NewBlankOutboundGroupSession()
	}
	length := base64.RawStdEncoding.EncodedLen(len(rawPickled))
	pickled := make([]byte, length)
	base64.RawStdEncoding.Encode(pickled, rawPickled)
	return s.Unpickle(pickled, pickleKey)
}

// Deprecated
func (s *OutboundGroupSession) MarshalJSON() ([]byte, error) {
	pickled, err := s.Pickle(pickleKey)
	if err != nil {
		return nil, err
	}
	quotes := make([]byte, len(pickled)+2)
	quotes[0] = '"'
	quotes[len(quotes)-1] = '"'
	copy(quotes[1:len(quotes)-1], pickled)
	return quotes, nil
}

// Deprecated
func (s *OutboundGroupSession) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || data[0] != '"' || data[len(data)-1] != '"' {
		return olm.InputNotJSONString
	}
	if s == nil || s.int == nil {
		*s = *NewBlankOutboundGroupSession()
	}
	return s.Unpickle(data[1:len(data)-1], pickleKey)
}

// createRandomLen returns the number of random bytes needed to create an
// Account.
func (s *OutboundGroupSession) createRandomLen() uint {
	return uint(C.olm_init_outbound_group_session_random_length((*C.OlmOutboundGroupSession)(s.int)))
}

// encryptMsgLen returns the size of the next message in bytes for the given
// number of plain-text bytes.
func (s *OutboundGroupSession) encryptMsgLen(plainTextLen int) uint {
	return uint(C.olm_group_encrypt_message_length((*C.OlmOutboundGroupSession)(s.int), C.size_t(plainTextLen)))
}

// Encrypt encrypts a message using the Session.  Returns the encrypted message
// as base64.
func (s *OutboundGroupSession) Encrypt(plaintext []byte) ([]byte, error) {
	if len(plaintext) == 0 {
		return nil, olm.EmptyInput
	}
	message := make([]byte, s.encryptMsgLen(len(plaintext)))
	r := C.olm_group_encrypt(
		(*C.OlmOutboundGroupSession)(s.int),
		(*C.uint8_t)(&plaintext[0]),
		C.size_t(len(plaintext)),
		(*C.uint8_t)(&message[0]),
		C.size_t(len(message)))
	if r == errorVal() {
		return nil, s.lastError()
	}
	return message[:r], nil
}

// sessionIdLen returns the number of bytes needed to store a session ID.
func (s *OutboundGroupSession) sessionIdLen() uint {
	return uint(C.olm_outbound_group_session_id_length((*C.OlmOutboundGroupSession)(s.int)))
}

// ID returns a base64-encoded identifier for this session.
func (s *OutboundGroupSession) ID() id.SessionID {
	sessionID := make([]byte, s.sessionIdLen())
	r := C.olm_outbound_group_session_id(
		(*C.OlmOutboundGroupSession)(s.int),
		(*C.uint8_t)(&sessionID[0]),
		C.size_t(len(sessionID)))
	if r == errorVal() {
		panic(s.lastError())
	}
	return id.SessionID(sessionID[:r])
}

// MessageIndex returns the message index for this session.  Each message is
// sent with an increasing index; this returns the index for the next message.
func (s *OutboundGroupSession) MessageIndex() uint {
	return uint(C.olm_outbound_group_session_message_index((*C.OlmOutboundGroupSession)(s.int)))
}

// sessionKeyLen returns the number of bytes needed to store a session key.
func (s *OutboundGroupSession) sessionKeyLen() uint {
	return uint(C.olm_outbound_group_session_key_length((*C.OlmOutboundGroupSession)(s.int)))
}

// Key returns the base64-encoded current ratchet key for this session.
func (s *OutboundGroupSession) Key() string {
	sessionKey := make([]byte, s.sessionKeyLen())
	r := C.olm_outbound_group_session_key(
		(*C.OlmOutboundGroupSession)(s.int),
		(*C.uint8_t)(&sessionKey[0]),
		C.size_t(len(sessionKey)))
	if r == errorVal() {
		panic(s.lastError())
	}
	return string(sessionKey[:r])
}
