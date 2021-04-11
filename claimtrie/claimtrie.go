package claimtrie

import (
	"fmt"
	"path/filepath"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

// ClaimTrie implements a Merkle Trie supporting linear history of commits.
type ClaimTrie struct {
	cm *commitMgr
	nm *nodeMgr

	// Merkle Trie of the ClaimTrie.
	trie *merkleTrie

	cleanup func() error
}

var (
	defaultHomeDir = btcutil.AppDataDir("lbrycrd.go", false)
	defaultDataDir = filepath.Join(defaultHomeDir, "data")
)

// New returns a ClaimTrie.
func New() (*ClaimTrie, error) {
	path := filepath.Join(defaultDataDir, "trie.db")
	dbTrie, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "can't open %s", path)
	}

	path = filepath.Join(defaultDataDir, "nm.db")
	dbNodeMgr, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "can't open %s", path)
	}

	path = filepath.Join(defaultDataDir, "commit.db")
	dbCommit, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "can't open %s", path)
	}

	cm := newCommitMgr(dbCommit)
	if err := cm.load(); err != nil {
		return nil, errors.Wrapf(err, "cm.Load()")
	}
	fmt.Printf("%d of commits loaded. Head: %d\n", len(cm.commits), cm.head.Height)

	nm := newNodeMgr(dbNodeMgr)
	nm.Load(cm.head.Height)
	fmt.Printf("%d of nodes loaded.\n", nm.size())

	tr := newMerkleTrie(nm, dbTrie)
	tr.SetRoot(cm.head.MerkleRoot)
	fmt.Printf("ClaimTrie Root: %s.\n", tr.MerkleHash())

	ct := &ClaimTrie{
		cm:   cm,
		nm:   nm,
		trie: tr,

		cleanup: func() error {
			if err := nm.save(); err != nil {
				return errors.Wrapf(err, "nm.Save()")
			}
			if err := cm.save(); err != nil {
				return errors.Wrapf(err, "cm.Save()")
			}
			if err := dbTrie.Close(); err != nil {
				return errors.Wrapf(err, "dbTrie.Close()")
			}
			if err := dbNodeMgr.Close(); err != nil {
				return errors.Wrapf(err, "dbClose()")
			}
			if err := dbCommit.Close(); err != nil {
				return errors.Wrapf(err, "dbCommit.Close()")
			}
			return nil
		},
	}
	return ct, nil
}

// Close saves ClaimTrie state to database.
func (ct *ClaimTrie) Close() error {
	return ct.cleanup()
}

// Height returns the highest height of blocks commited to the ClaimTrie.
func (ct *ClaimTrie) Height() Height {
	return ct.cm.head.Height
}

// AddClaim adds a Claim to the ClaimTrie.
func (ct *ClaimTrie) AddClaim(name string, op wire.OutPoint, amt Amount, val []byte) error {
	c := newChange(cmdAddClaim).setOP(op).setAmt(amt).setValue(val)
	return ct.modify(name, c)
}

// SpendClaim spend a Claim in the ClaimTrie.
func (ct *ClaimTrie) SpendClaim(name string, op wire.OutPoint) error {
	c := newChange(cmdSpendClaim).setOP(op)
	return ct.modify(name, c)
}

// UpdateClaim updates a Claim in the ClaimTrie.
func (ct *ClaimTrie) UpdateClaim(name string, op wire.OutPoint, amt Amount, id ClaimID, val []byte) error {
	c := newChange(cmdUpdateClaim).setOP(op).setAmt(amt).setID(id).setValue(val)
	return ct.modify(name, c)
}

// AddSupport adds a Support to the ClaimTrie.
func (ct *ClaimTrie) AddSupport(name string, op wire.OutPoint, amt Amount, id ClaimID) error {
	c := newChange(cmdAddSupport).setOP(op).setAmt(amt).setID(id)
	return ct.modify(name, c)
}

// SpendSupport spend a support in the ClaimTrie.
func (ct *ClaimTrie) SpendSupport(name string, op wire.OutPoint) error {
	c := newChange(cmdSpendSupport).setOP(op)
	return ct.modify(name, c)
}

// MerkleHash returns the Merkle Hash of the ClaimTrie.
func (ct *ClaimTrie) MerkleHash() *chainhash.Hash {
	return ct.trie.MerkleHash()
}

// Commit commits the current changes into database.
func (ct *ClaimTrie) Commit(ht Height) {
	if ht < ct.Height() {
		return
	}
	for i := ct.Height() + 1; i <= ht; i++ {
		ct.nm.catchUp(i, ct.trie.Update)
	}
	h := ct.MerkleHash()
	ct.cm.commit(ht, h)
	ct.trie.SetRoot(h)
}

// Reset resets the tip commit to a previous height specified.
func (ct *ClaimTrie) Reset(ht Height) error {
	if ht > ct.Height() {
		return errInvalidHeight
	}
	ct.cm.reset(ht)
	ct.nm.reset(ht)
	ct.trie.SetRoot(ct.cm.head.MerkleRoot)
	return nil
}

// Node returns the node adjusted to specified height.
func (ct *ClaimTrie) Node(name string) *Node {
	return ct.nm.nodeAt(name, ct.Height())
}

// Size returns the number of nodes loaded into the cache.
func (ct *ClaimTrie) Size() int {
	return ct.nm.size()
}

// Visit visits every node in the cache with VisitFunc.
// If the VisitFunc returns true, the iteration ends immediately.
func (ct *ClaimTrie) Visit(v visitFunc) {
	ct.nm.visit(v)
}

func (ct *ClaimTrie) modify(name string, c *change) error {
	c.setHeight(ct.Height() + 1).setName(name)
	if err := ct.nm.modifyNode(name, c); err != nil {
		return err
	}
	ct.trie.Update([]byte(name))
	return nil
}
