package secp256k1

import (
	"fmt"

	"golang.org/x/crypto/sha3"
)

// Keccak256 computes the Keccak256 hash of the input.
func Keccak256(data []byte) []byte {
	hash := sha3.NewLegacyKeccak256()
	hash.Write(data)
	return hash.Sum(nil)
}

// Keccak256Hash computes the Keccak256 hash of multiple inputs.
func Keccak256Hash(data ...[]byte) []byte {
	hash := sha3.NewLegacyKeccak256()
	for _, d := range data {
		hash.Write(d)
	}
	return hash.Sum(nil)
}

// HashEthereumMessage hashes a message with the Ethereum signed message prefix.
// This matches the behavior of eth_sign and personal_sign.
func HashEthereumMessage(data []byte) []byte {
	prefix := fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(data))
	return Keccak256(append([]byte(prefix), data...))
}
