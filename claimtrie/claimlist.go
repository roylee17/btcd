package claimtrie

import "github.com/btcsuite/btcd/wire"

type claimList []*Claim

type comparator func(c *Claim) bool

func ByOP(op wire.OutPoint) comparator {
	return func(c *Claim) bool {
		return c.OutPoint == op
	}
}

func ByID(id ClaimID) comparator {
	return func(c *Claim) bool {
		return c.ID == id
	}
}

func remove(l claimList, cmp comparator) (claimList, *Claim) {
	last := len(l) - 1
	for i, v := range l {
		if !cmp(v) {
			continue
		}
		removed := l[i]
		l[i] = l[last]
		l[last] = nil
		return l[:last], removed
	}
	return l, nil
}

func Find(cmp comparator, lists ...claimList) *Claim {
	for _, l := range lists {
		for _, v := range l {
			if cmp(v) {
				return v
			}
		}
	}
	return nil
}
