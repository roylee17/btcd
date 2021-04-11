package claimtrie

import (
	"bytes"

	"github.com/btcsuite/btcd/wire"
)

type (
	Amount int64 // Amount defines the amount in LBC.
	Height int32 // Height defines the height of a block.
)

// newClaim returns a Claim (or Support) initialized with specified op and amt.
func newClaim(op wire.OutPoint, amt Amount) *Claim {
	return &Claim{OutPoint: op, Amount: amt}
}

// Claim defines a structure of a Claim (or Support).
type Claim struct {
	OutPoint wire.OutPoint
	ID       ClaimID
	Amount   Amount
	Accepted Height
	Value    []byte

	EffectiveAmount Amount
	ActiveAt        Height
}

func (c *Claim) setOutPoint(op wire.OutPoint) *Claim { c.OutPoint = op; return c }
func (c *Claim) setID(id ClaimID) *Claim             { c.ID = id; return c }
func (c *Claim) setAmt(amt Amount) *Claim            { c.Amount = amt; return c }
func (c *Claim) setAccepted(ht Height) *Claim        { c.Accepted = ht; return c }
func (c *Claim) setActiveAt(ht Height) *Claim        { c.ActiveAt = ht; return c }
func (c *Claim) setValue(val []byte) *Claim          { c.Value = val; return c }

func (c *Claim) expireAt() Height {
	if c.Accepted+paramOriginalClaimExpirationTime > paramExtendedClaimExpirationForkHeight {
		return c.Accepted + paramExtendedClaimExpirationTime
	}
	return c.Accepted + paramOriginalClaimExpirationTime
}

func IsActiveAt(c *Claim, ht Height) bool {
	return c != nil && c.ActiveAt <= ht && c.expireAt() > ht
}

func equal(a, b *Claim) bool {
	if a != nil && b != nil {
		return a.OutPoint == b.OutPoint
	}
	return a == nil && b == nil
}

func outPointLess(a, b wire.OutPoint) bool {
	switch cmp := bytes.Compare(a.Hash[:], b.Hash[:]); {
	case cmp > 0:
		return true
	case cmp < 0:
		return false
	default:
		return a.Index < b.Index
	}
}
