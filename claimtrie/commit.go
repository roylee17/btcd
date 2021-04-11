package claimtrie

import (
	"bytes"
	"encoding/gob"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

func newCommit(head *commit, height Height, h *chainhash.Hash) *commit {
	return &commit{
		MerkleRoot: h,
		Height:     height,
	}
}

type commit struct {
	MerkleRoot *chainhash.Hash
	Height     Height
}

type commitMgr struct {
	db      *leveldb.DB
	commits []*commit
	head    *commit
}

func newCommitMgr(db *leveldb.DB) *commitMgr {
	head := newCommit(nil, 0, emptyTrieHash)
	cm := commitMgr{
		db:   db,
		head: head,
	}
	cm.commits = append(cm.commits, head)
	return &cm
}

func (cm *commitMgr) commit(ht Height, merkle *chainhash.Hash) {
	if ht == 0 {
		return
	}
	c := newCommit(cm.head, ht, merkle)
	cm.commits = append(cm.commits, c)
	cm.head = c
}

func (cm *commitMgr) reset(ht Height) {
	for i := len(cm.commits) - 1; i >= 0; i-- {
		c := cm.commits[i]
		if c.Height <= ht {
			cm.head = c
			cm.commits = cm.commits[:i+1]
			break
		}
	}
	if cm.head.Height == ht {
		return
	}
	cm.commit(ht, cm.head.MerkleRoot)
}

func (cm *commitMgr) save() error {
	exported := struct {
		Commits []*commit
		Head    *commit
	}{
		Commits: cm.commits,
		Head:    cm.head,
	}

	buf := bytes.NewBuffer(nil)
	if err := gob.NewEncoder(buf).Encode(exported); err != nil {
		return errors.Wrapf(err, "gob.Encode()", err)
	}
	if err := cm.db.Put([]byte("CommitMgr"), buf.Bytes(), nil); err != nil {
		return errors.Wrapf(err, "db.Put(CommitMgr)")
	}
	return nil
}

func (cm *commitMgr) load() error {
	exported := struct {
		Commits []*commit
		Head    *commit
	}{}

	data, err := cm.db.Get([]byte("CommitMgr"), nil)
	if err == leveldb.ErrNotFound {
		return nil
	}
	if err != nil {
		return errors.Wrapf(err, "db.Get(CommitMgr)")
	}
	if err := gob.NewDecoder(bytes.NewBuffer(data)).Decode(&exported); err != nil {
		return errors.Wrapf(err, "gob.Encode()", err)
	}
	cm.commits = exported.Commits
	cm.head = exported.Head
	return nil
}
