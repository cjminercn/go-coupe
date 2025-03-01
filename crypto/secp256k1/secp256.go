// Copyright 2015 The go-coupe Authors
// This file is part of the go-coupe library.
//
// The go-coupe library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-coupe library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-coupe library. If not, see <http://www.gnu.org/licenses/>.

package secp256k1

// TODO: set USE_SCALAR_4X64 depending on platform?

/*
#cgo CFLAGS: -I./libsecp256k1
#cgo darwin CFLAGS: -I/usr/local/include
#cgo freebsd CFLAGS: -I/usr/local/include
#cgo linux,arm CFLAGS: -I/usr/local/arm/include
#cgo LDFLAGS: -lgmp
#cgo darwin LDFLAGS: -L/usr/local/lib
#cgo freebsd LDFLAGS: -L/usr/local/lib
#cgo linux,arm LDFLAGS: -L/usr/local/arm/lib
#define USE_NUM_GMP
#define USE_FIELD_10X26
#define USE_FIELD_INV_BUILTIN
#define USE_SCALAR_8X32
#define USE_SCALAR_INV_BUILTIN
#define NDEBUG
#include "./libsecp256k1/src/secp256k1.c"
#include "./libsecp256k1/src/modules/recovery/main_impl.h"

typedef void (*callbackFunc) (const char* msg, void* data);
extern void secp256k1GoPanicIllegal(const char* msg, void* data);
extern void secp256k1GoPanicError(const char* msg, void* data);
*/
import "C"

import (
	"errors"
	"unsafe"

	"github.com/cjminercn/go-coupe/crypto/randentropy"
)

//#define USE_FIELD_5X64

/*
   TODO:
   > store private keys in buffer and shuffle (deters persistance on swap disc)
   > byte permutation (changing)
   > xor with chaning random block (to deter scanning memory for 0x63) (stream cipher?)
   > on disk: store keys in wallets
*/

// holds ptr to secp256k1_context_struct (see secp256k1/include/secp256k1.h)
var context *C.secp256k1_context

func init() {
	// around 20 ms on a modern CPU.
	context = C.secp256k1_context_create(3) // SECP256K1_START_SIGN | SECP256K1_START_VERIFY
	C.secp256k1_context_set_illegal_callback(context, C.callbackFunc(C.secp256k1GoPanicIllegal), nil)
	C.secp256k1_context_set_error_callback(context, C.callbackFunc(C.secp256k1GoPanicError), nil)
}

var (
	ErrInvalidMsgLen       = errors.New("invalid message length for signature recovery")
	ErrInvalidSignatureLen = errors.New("invalid signature length")
	ErrInvalidRecoveryID   = errors.New("invalid signature recovery id")
)

func GenerateKeyPair() ([]byte, []byte) {
	var seckey []byte = randentropy.GetEntropyCSPRNG(32)
	var seckey_ptr *C.uchar = (*C.uchar)(unsafe.Pointer(&seckey[0]))

	var pubkey64 []byte = make([]byte, 64) // secp256k1_pubkey
	var pubkey65 []byte = make([]byte, 65) // 65 byte uncompressed pubkey
	pubkey64_ptr := (*C.secp256k1_pubkey)(unsafe.Pointer(&pubkey64[0]))
	pubkey65_ptr := (*C.uchar)(unsafe.Pointer(&pubkey65[0]))

	ret := C.secp256k1_ec_pubkey_create(
		context,
		pubkey64_ptr,
		seckey_ptr,
	)

	if ret != C.int(1) {
		return GenerateKeyPair() // invalid secret, try again
	}

	var output_len C.size_t

	C.secp256k1_ec_pubkey_serialize( // always returns 1
		context,
		pubkey65_ptr,
		&output_len,
		pubkey64_ptr,
		0, // SECP256K1_EC_COMPRESSED
	)

	return pubkey65, seckey
}

func GeneratePubKey(seckey []byte) ([]byte, error) {
	if err := VerifySeckeyValidity(seckey); err != nil {
		return nil, err
	}

	var pubkey []byte = make([]byte, 64)
	var pubkey_ptr *C.secp256k1_pubkey = (*C.secp256k1_pubkey)(unsafe.Pointer(&pubkey[0]))

	var seckey_ptr *C.uchar = (*C.uchar)(unsafe.Pointer(&seckey[0]))

	ret := C.secp256k1_ec_pubkey_create(
		context,
		pubkey_ptr,
		seckey_ptr,
	)

	if ret != C.int(1) {
		return nil, errors.New("Unable to generate pubkey from seckey")
	}

	return pubkey, nil
}

func Sign(msg []byte, seckey []byte) ([]byte, error) {
	msg_ptr := (*C.uchar)(unsafe.Pointer(&msg[0]))
	seckey_ptr := (*C.uchar)(unsafe.Pointer(&seckey[0]))

	sig := make([]byte, 65)
	sig_ptr := (*C.secp256k1_ecdsa_recoverable_signature)(unsafe.Pointer(&sig[0]))

	nonce := randentropy.GetEntropyCSPRNG(32)
	ndata_ptr := unsafe.Pointer(&nonce[0])

	noncefp_ptr := &(*C.secp256k1_nonce_function_default)

	if C.secp256k1_ec_seckey_verify(context, seckey_ptr) != C.int(1) {
		return nil, errors.New("Invalid secret key")
	}

	ret := C.secp256k1_ecdsa_sign_recoverable(
		context,
		sig_ptr,
		msg_ptr,
		seckey_ptr,
		noncefp_ptr,
		ndata_ptr,
	)

	if ret == C.int(0) {
		return Sign(msg, seckey) //invalid secret, try again
	}

	sig_serialized := make([]byte, 65)
	sig_serialized_ptr := (*C.uchar)(unsafe.Pointer(&sig_serialized[0]))
	var recid C.int

	C.secp256k1_ecdsa_recoverable_signature_serialize_compact(
		context,
		sig_serialized_ptr, // 64 byte compact signature
		&recid,
		sig_ptr, // 65 byte "recoverable" signature
	)

	sig_serialized[64] = byte(int(recid)) // add back recid to get 65 bytes sig

	return sig_serialized, nil

}

func VerifySeckeyValidity(seckey []byte) error {
	if len(seckey) != 32 {
		return errors.New("priv key is not 32 bytes")
	}
	var seckey_ptr *C.uchar = (*C.uchar)(unsafe.Pointer(&seckey[0]))
	ret := C.secp256k1_ec_seckey_verify(context, seckey_ptr)
	if int(ret) != 1 {
		return errors.New("invalid seckey")
	}
	return nil
}

// RecoverPubkey returns the the public key of the signer.
// msg must be the 32-byte hash of the message to be signed.
// sig must be a 65-byte compact ECDSA signature containing the
// recovery id as the last element.
func RecoverPubkey(msg []byte, sig []byte) ([]byte, error) {
	if len(msg) != 32 {
		return nil, ErrInvalidMsgLen
	}
	if err := checkSignature(sig); err != nil {
		return nil, err
	}

	msg_ptr := (*C.uchar)(unsafe.Pointer(&msg[0]))
	sig_ptr := (*C.uchar)(unsafe.Pointer(&sig[0]))
	pubkey := make([]byte, 64)
	/*
		this slice is used for both the recoverable signature and the
		resulting serialized pubkey (both types in libsecp256k1 are 65
		bytes). this saves one allocation of 65 bytes, which is nice as
		pubkey recovery is one bottleneck during load in cjminercn
	*/
	bytes65 := make([]byte, 65)
	pubkey_ptr := (*C.secp256k1_pubkey)(unsafe.Pointer(&pubkey[0]))
	recoverable_sig_ptr := (*C.secp256k1_ecdsa_recoverable_signature)(unsafe.Pointer(&bytes65[0]))
	recid := C.int(sig[64])

	ret := C.secp256k1_ecdsa_recoverable_signature_parse_compact(
		context,
		recoverable_sig_ptr,
		sig_ptr,
		recid)
	if ret == C.int(0) {
		return nil, errors.New("Failed to parse signature")
	}

	ret = C.secp256k1_ecdsa_recover(
		context,
		pubkey_ptr,
		recoverable_sig_ptr,
		msg_ptr,
	)
	if ret == C.int(0) {
		return nil, errors.New("Failed to recover public key")
	}

	serialized_pubkey_ptr := (*C.uchar)(unsafe.Pointer(&bytes65[0]))
	var output_len C.size_t
	C.secp256k1_ec_pubkey_serialize( // always returns 1
		context,
		serialized_pubkey_ptr,
		&output_len,
		pubkey_ptr,
		0, // SECP256K1_EC_COMPRESSED
	)
	return bytes65, nil
}

func checkSignature(sig []byte) error {
	if len(sig) != 65 {
		return ErrInvalidSignatureLen
	}
	if sig[64] >= 4 {
		return ErrInvalidRecoveryID
	}
	return nil
}
