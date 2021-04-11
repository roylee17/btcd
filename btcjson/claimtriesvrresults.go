// Copyright (c) 2014-2017 The btcsuite developers
// Copyright (c) 2015-2017 The Decred developers
// Copyright (c) 2018-2018 The LBRY developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btcjson

import "github.com/btcsuite/btcd/claimtrie"

// EmptyResult models an empty JSON object.
type EmptyResult struct{}

// GetClaimsInTrieResult models the data from the GetClaimsInTrie command.
type GetClaimsInTrieResult []ClaimsInTrieEntry

// ClaimsInTrieEntry models the data from the ClaimsInTrie command.
type ClaimsInTrieEntry struct {
	Name   string               `json:"name"`
	Claims []ClaimsInTrieDetail `json:"claims"`
}

// ClaimsInTrieDetail models the Claim from the GetClaimsInTrie command.
type ClaimsInTrieDetail struct {
	ClaimID string           `json:"claimId"`
	TxID    string           `json:"txid"`
	N       uint32           `json:"n"`
	Amount  string           `json:"amount"`
	Height  claimtrie.Height `json:"height"`
	Value   string           `json:"value"`
}

// GetClaimTrieResult models the data from the GetClaimTrie command.
type GetClaimTrieResult []ClaimTrieEntry

// ClaimTrieEntry models the data from the GetClaimTrie command.
type ClaimTrieEntry struct {
	Name   string           `json:"name"`
	Hash   string           `json:"hash"`
	TxID   string           `json:"txid"`
	N      uint32           `json:"n"`
	Value  claimtrie.Amount `json:"value"`
	Height claimtrie.Height `json:"height"`
}

// GetValueForNameResult models the data from the GetValueForName command.
type GetValueForNameResult struct {
	Value           string           `json:"value"`
	ClaimID         string           `json:"claimId"`
	TxID            string           `json:"txid"`
	N               uint32           `json:"n"`
	Amount          claimtrie.Amount `json:"amount"`
	EffectiveAmount claimtrie.Amount `json:"effective amount"`
	Height          claimtrie.Height `json:"height"`
}

// GetClaimsForNameResult models the data from the GetClaimsForName command.
type GetClaimsForNameResult struct {
	LastTakeoverHeight claimtrie.Height `json:"nLastTakeoverheight"`
	Claims             []ClaimForName   `json:"claims"`
	UnmatchedSupports  []SupportOfClaim `json:"unmatched supports"`
}

// ClaimForName models the Claim from the GetClaimsForName command.
type ClaimForName struct {
	ClaimID         string           `json:"claimId"`
	TxID            string           `json:"txid"`
	N               uint32           `json:"n"`
	Height          claimtrie.Height `json:"nHeight"`
	ValidHeight     claimtrie.Height `json:"nValidAtHeight"`
	Amount          claimtrie.Amount `json:"nAmount"`
	EffectiveAmount claimtrie.Amount `json:"nEffectiveAmount"`
	Supports        []SupportOfClaim `json:"supports"`
}

// SupportOfClaim models the data of support from the GetClaimsForName command.
type SupportOfClaim struct {
	TxID        string           `json:"txid"`
	N           uint32           `json:"n"`
	Height      claimtrie.Height `json:"nHeight"`
	ValidHeight claimtrie.Height `json:"nValidAtHeight"`
	Amount      claimtrie.Amount `json:"nAmount"`
}

// GetNameProofResult models the data from the GetNameProof command.
type GetNameProofResult struct {
	Nodes              []NameProofNode  `json:"nodes"`
	TxHash             string           `json:"txhash"`
	N                  uint32           `json:"nOut"`
	LastTakeoverHeight claimtrie.Height `json:"last takeover height"`
}

// NameProofNode models the Node from the GetNameProof command.
type NameProofNode struct {
	Children  []NameProofNodeChild `json:"children"`
	ValueHash string               `json:"valueHash"`
}

// NameProofNodeChild models the Child of Node from the GetNameProof command.
type NameProofNodeChild struct {
	Character string `json:"character"`
	NodeHash  string `json:"nodeHash"`
}

// GetClaimsForTxResult models the data from the GetClaimsForTx command.
type GetClaimsForTxResult []ClaimsForTxEntry

// ClaimsForTxEntry models the data from the GetClaimsForTx command.
type ClaimsForTxEntry struct {
	N             uint32           `json:"nOut"`
	Type          string           `json:"claim type"`
	Name          string           `json:"name"`
	Value         string           `json:"value"`
	SupportedID   string           `json:"supported txid"`
	SupportedNOut uint32           `json:"supported nout"`
	Depth         claimtrie.Height `json:"depth"`
	InClaimTrie   bool             `json:"in claim trie"`
	IsControlling bool             `json:"is controlling"`
	InSupportMap  bool             `json:"in support map"`
	InQueue       bool             `json:"in queue"`
	BlocksToValid claimtrie.Height `json:"blocks to valid"`
}

// GetClaimByIDResult models the data from the GetClaimByID command.
type GetClaimByIDResult struct {
	Name        string             `json:"name"`
	Value       string             `json:"value"`
	ClaimID     string             `json:"claimId"`
	TxID        string             `json:"txid"`
	N           uint32             `json:"n"`
	Amount      claimtrie.Amount   `json:"amount"`
	EffAmount   claimtrie.Amount   `json:"effective amount"`
	Supports    []ClaimByIDSupport `json:"supports"`
	Height      claimtrie.Height   `json:"height"`
	ValidHeight claimtrie.Height   `json:"valid at height"`
}

// ClaimByIDSupport models the data of support from the GetClaimByID command.
type ClaimByIDSupport struct {
	TxID        string           `json:"txid"`
	N           uint32           `json:"n"`
	Height      claimtrie.Height `json:"height"`
	ValidHeight claimtrie.Height `json:"valid at height"`
	Amount      claimtrie.Amount `json:"amount"`
}
