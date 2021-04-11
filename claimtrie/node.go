package claimtrie

import (
	"crypto/sha256"
	"encoding/binary"
	"math"
	"strconv"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/pkg/errors"
)

type Node struct {
	Name      string    // Name where the Node blongs.
	Height    Height    // Height is the current height.
	BestClaim *Claim    // BestClaim is the best claim at the current height.
	Tookover  Height    // Tookover is the height at when the current BestClaim Tookover.
	Claims    claimList // Claims returns the Claims at the current height.
	Supports  claimList // Supports returns the Supports at the current height.

	removed claimList // refer to updateClaim.
}

// NewNode returns a new Node.
func NewNode(name string) *Node {
	return &Node{Name: name}
}

// addClaim adds a Claim to the Node.
func (n *Node) addClaim(op wire.OutPoint, amt Amount, val []byte) error {
	if Find(ByOP(op), n.Claims, n.Supports) != nil {
		return errDuplicate
	}
	accepted := n.Height + 1
	c := newClaim(op, amt).setID(NewID(op)).setAccepted(accepted).setValue(val)
	c.setActiveAt(accepted + calculateDelay(accepted, n.Tookover))
	if !IsActiveAt(n.BestClaim, accepted) {
		c.setActiveAt(accepted)
		n.BestClaim, n.Tookover = c, accepted
	}
	n.Claims = append(n.Claims, c)
	return nil
}

// spendClaim spends a Claim in the Node.
func (n *Node) spendClaim(op wire.OutPoint) error {
	var c *Claim
	if n.Claims, c = remove(n.Claims, ByOP(op)); c == nil {
		return errNotFound
	}
	n.removed = append(n.removed, c)
	return nil
}

// updateClaim updates a Claim in the Node.
// A claim update is composed of two separate commands (2 & 3 below).
//
//   (1) blk  500: Add Claim (opA, amtA, NewID(opA)
//     ...
//   (2) blk 1000: Spend Claim (opA, idA)
//   (3) blk 1000: Update Claim (opB, amtB, idA)
//
// For each block, all the spent claims are kept in n.removed until committed.
// The paired (spend, update) commands has to happen in the same trasaction.
func (n *Node) updateClaim(op wire.OutPoint, amt Amount, id ClaimID, val []byte) error {
	if Find(ByOP(op), n.Claims, n.Supports) != nil {
		return errDuplicate
	}
	var c *Claim
	if n.removed, c = remove(n.removed, ByID(id)); c == nil {
		return errors.Wrapf(errNotFound, "remove(n.removed, byID(%s)", id)
	}

	accepted := n.Height + 1
	c.setOutPoint(op).setAmt(amt).setAccepted(accepted).setValue(val)
	c.setActiveAt(accepted + calculateDelay(accepted, n.Tookover))
	if n.BestClaim != nil && n.BestClaim.ID == id {
		c.setActiveAt(n.Tookover)
	}
	n.Claims = append(n.Claims, c)
	return nil
}

// addSupport adds a Support to the Node.
func (n *Node) addSupport(op wire.OutPoint, amt Amount, id ClaimID) error {
	if Find(ByOP(op), n.Claims, n.Supports) != nil {
		return errDuplicate
	}
	// Accepted by rules. No effects on bidding result though.
	// It may be spent later.
	if Find(ByID(id), n.Claims, n.removed) == nil {
		// fmt.Printf("INFO: can't find suooported claim ID: %s for %s\n", id, n.name)
	}

	accepted := n.Height + 1
	s := newClaim(op, amt).setID(id).setAccepted(accepted)
	s.setActiveAt(accepted + calculateDelay(accepted, n.Tookover))
	if n.BestClaim != nil && n.BestClaim.ID == id {
		s.setActiveAt(accepted)
	}
	n.Supports = append(n.Supports, s)
	return nil
}

// spendSupport spends a support in the Node.
func (n *Node) spendSupport(op wire.OutPoint) error {
	var s *Claim
	if n.Supports, s = remove(n.Supports, ByOP(op)); s != nil {
		return nil
	}
	return errNotFound
}

// adjustTo increments current height until it reaches the specific height.
func (n *Node) adjustTo(ht Height) *Node {
	if ht <= n.Height {
		return n
	}
	for n.Height < ht {
		n.Height++
		n.bid()
		next := n.nextUpdate()
		if next > ht || next == n.Height {
			n.Height = ht
			break
		}
		n.Height = next
		n.bid()
	}
	n.bid()
	return n
}

// nextUpdate returns the height at which pending updates should happen.
// When no pending updates exist, current height is returned.
func (n *Node) nextUpdate() Height {
	next := Height(math.MaxInt32)
	min := func(l claimList) Height {
		for _, v := range l {
			exp := v.expireAt()
			if n.Height >= exp {
				continue
			}
			if v.ActiveAt > n.Height && v.ActiveAt < next {
				next = v.ActiveAt
			}
			if exp > n.Height && exp < next {
				next = exp
			}
		}
		return next
	}
	min(n.Claims)
	min(n.Supports)
	if next == Height(math.MaxInt32) {
		next = n.Height
	}
	return next
}

// Hash calculates the Hash value based on the OutPoint and when it tookover.
func (n *Node) Hash() *chainhash.Hash {
	if n.BestClaim == nil {
		return nil
	}
	return calculateNodeHash(n.BestClaim.OutPoint, n.Tookover)
}

func (n *Node) bid() {
	for {
		if n.BestClaim == nil || n.Height >= n.BestClaim.expireAt() {
			n.BestClaim, n.Tookover = nil, n.Height
			updateActiveHeights(n, n.Claims, n.Supports)
		}
		updateEffectiveAmounts(n.Height, n.Claims, n.Supports)
		c := findCandiadte(n.Height, n.Claims)
		if equal(n.BestClaim, c) {
			break
		}
		n.BestClaim, n.Tookover = c, n.Height
		updateActiveHeights(n, n.Claims, n.Supports)
	}
	n.removed = nil
}

func updateEffectiveAmounts(ht Height, claims, supports claimList) {
	for _, c := range claims {
		c.EffectiveAmount = 0
		if !IsActiveAt(c, ht) {
			continue
		}
		c.EffectiveAmount = c.Amount
		for _, s := range supports {
			if !IsActiveAt(s, ht) || s.ID != c.ID {
				continue
			}
			c.EffectiveAmount += s.Amount
		}
	}
}

func updateActiveHeights(n *Node, lists ...claimList) {
	for _, l := range lists {
		for _, v := range l {
			if v.ActiveAt < n.Height {
				continue
			}
			v.ActiveAt = v.Accepted + calculateDelay(n.Height, n.Tookover)
			if v.ActiveAt < n.Height {
				v.ActiveAt = n.Height
			}
		}
	}
}

func findCandiadte(ht Height, claims claimList) *Claim {
	var c *Claim
	for _, v := range claims {
		switch {
		case !IsActiveAt(v, ht):
			continue
		case c == nil:
			c = v
		case v.EffectiveAmount > c.EffectiveAmount:
			c = v
		case v.EffectiveAmount < c.EffectiveAmount:
			continue
		case v.Accepted < c.Accepted:
			c = v
		case v.Accepted > c.Accepted:
			continue
		case outPointLess(c.OutPoint, v.OutPoint):
			c = v
		}
	}
	return c
}

func calculateDelay(curr, tookover Height) Height {
	delay := (curr - tookover) / paramActiveDelayFactor
	if delay > paramMaxActiveDelay {
		return paramMaxActiveDelay
	}
	return delay
}

func calculateNodeHash(op wire.OutPoint, tookover Height) *chainhash.Hash {
	txHash := chainhash.DoubleHashH(op.Hash[:])

	nOut := []byte(strconv.Itoa(int(op.Index)))
	nOutHash := chainhash.DoubleHashH(nOut)

	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(tookover))
	heightHash := chainhash.DoubleHashH(buf)

	h := make([]byte, 0, sha256.Size*3)
	h = append(h, txHash[:]...)
	h = append(h, nOutHash[:]...)
	h = append(h, heightHash[:]...)

	hh := chainhash.DoubleHashH(h)
	return &hh
}
