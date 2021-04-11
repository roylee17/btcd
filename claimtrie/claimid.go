package claimtrie

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"

	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

// NewID returns a Claim ID caclculated from Ripemd160(Sha256(OUTPOINT).
func NewID(op wire.OutPoint) ClaimID {
	w := bytes.NewBuffer(op.Hash[:])
	if err := binary.Write(w, binary.BigEndian, op.Index); err != nil {
		panic(err)
	}
	var id ClaimID
	copy(id[:], btcutil.Hash160(w.Bytes()))
	return id
}

// NewIDFromString returns a Claim ID from a string.
func NewIDFromString(s string) (ClaimID, error) {
	var id ClaimID
	if len(s) != 40 {
		return id, errInvalidID
	}
	_, err := hex.Decode(id[:], []byte(s))
	for i, j := 0, len(id)-1; i < j; i, j = i+1, j-1 {
		id[i], id[j] = id[j], id[i]
	}
	return id, err
}

// ClaimID represents a Claim's ClaimID.
type ClaimID [20]byte

func (id ClaimID) String() string {
	for i, j := 0, len(id)-1; i < j; i, j = i+1, j-1 {
		id[i], id[j] = id[j], id[i]
	}
	return hex.EncodeToString(id[:])
}
