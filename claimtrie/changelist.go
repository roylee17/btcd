package claimtrie

import (
	"bytes"
	"encoding/gob"

	"github.com/btcsuite/btcd/wire"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

type changeList struct {
	db      *leveldb.DB
	name    string
	changes []*change
	err     error
}

func newChangeList(db *leveldb.DB, name string) *changeList {
	return &changeList{db: db, name: name}
}

func (cl *changeList) load() *changeList {
	if cl.err == nil {
		cl.changes, cl.err = loadChanges(cl.db, cl.name)
	}
	return cl
}

func (cl *changeList) save() *changeList {
	if cl.err == nil {
		cl.err = saveChanges(cl.db, cl.name, cl.changes)
	}
	return cl
}

// append appends a Change to the Changes in the list.
func (cl *changeList) append(chg *change) *changeList {
	cl.changes = append(cl.changes, chg)
	return cl
}

// truncate truncates Changes that has Heiht larger than ht.
func (cl *changeList) truncate(ht Height) *changeList {
	for i, chg := range cl.changes {
		if chg.height > ht {
			cl.changes = cl.changes[:i]
			break
		}
	}
	return cl
}

func loadChanges(db *leveldb.DB, name string) ([]*change, error) {
	data, err := db.Get([]byte(name), nil)
	if err == leveldb.ErrNotFound {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrapf(err, "db.Get(%s)", name)
	}
	var chgs []*change
	if err = gob.NewDecoder(bytes.NewBuffer(data)).Decode(&chgs); err != nil {
		return nil, errors.Wrapf(err, "gob.Decode(&blk)")
	}
	return chgs, nil
}

func saveChanges(db *leveldb.DB, name string, chgs []*change) error {
	buf := bytes.NewBuffer(nil)
	if err := gob.NewEncoder(buf).Encode(&chgs); err != nil {
		return errors.Wrapf(err, "gob.Decode(&blk)")
	}
	return errors.Wrapf(db.Put([]byte(name), buf.Bytes(), nil), "db.put(%s, buf)", name)
}

// command defines the type of Change.
type command int

const (
	cmdAddClaim command = 1 << iota
	cmdSpendClaim
	cmdUpdateClaim
	cmdAddSupport
	cmdSpendSupport
)

// change represent a record of changes to the node of Name at Height.
type change struct {
	height Height
	cmd    command
	name   string
	op     wire.OutPoint
	amount Amount
	id     ClaimID
	value  []byte
}

func newChange(cmd command) *change {
	return &change{cmd: cmd}
}

func (c *change) setName(name string) *change    { c.name = name; return c }
func (c *change) setHeight(h Height) *change     { c.height = h; return c }
func (c *change) setOP(op wire.OutPoint) *change { c.op = op; return c }
func (c *change) setAmt(amt Amount) *change      { c.amount = amt; return c }
func (c *change) setID(id ClaimID) *change       { c.id = id; return c }
func (c *change) setValue(v []byte) *change      { c.value = v; return c }
