// blsproof generates a BLS keypair and the proof-of-possession that
// StakeHub.createValidator requires.
//
// ── Why this tool has to exist ──────────────────────────────────────────
//
// createValidator takes a `blsProof`, checked on-chain by the BLS precompile
// at address 0x66 against:
//
//     msgHash = keccak256(operatorAddress ++ voteAddress ++ chainId)
//
// Building that signature needs the BLS SECRET key, and geth's `bls` command
// deliberately refuses to export one — "exporting your key in unencrypted
// format is NOT supported". That is the right default for an operator tool,
// and it leaves no way to produce a proof from a geth-generated key.
//
// So the key is generated here, where both halves are available: the proof
// for registration, and the raw secret so the same key can be imported into
// geth's BLS wallet for finality voting.
//
// ── Deliberately the same library as the chain ──────────────────────────
//
// github.com/prysmaticlabs/prysm/v5/crypto/bls — the exact package Windpay's own
// core/vote/vote_signer.go uses. A different BLS implementation can produce a
// signature that verifies locally and is rejected by the precompile, which is
// an expensive thing to debug.
package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/prysmaticlabs/prysm/v5/crypto/bls"
)

type output struct {
	Operator    string `json:"operator"`
	ChainID     uint64 `json:"chainId"`
	VoteAddress string `json:"voteAddress"` // BLS public key, 48 bytes
	BlsProof    string `json:"blsProof"`    // BLS signature, 96 bytes
	SecretKey   string `json:"secretKey"`   // 32 bytes, for `geth bls account import`
	MsgHash     string `json:"msgHash"`
}

func main() {
	operator := flag.String("operator", "", "address that will call createValidator")
	chainID := flag.Uint64("chainid", 0, "chain id")
	secretHex := flag.String("secret", "", "existing 32-byte secret key hex; a new key is generated when empty")
	flag.Parse()

	if *operator == "" || *chainID == 0 || !common.IsHexAddress(*operator) {
		fmt.Fprintln(os.Stderr, "usage: blsproof -operator 0xADDRESS -chainid N [-secret HEX]")
		os.Exit(2)
	}

	var sk bls.SecretKey
	var err error
	if *secretHex != "" {
		raw, decErr := hex.DecodeString(strings.TrimPrefix(strings.TrimPrefix(*secretHex, "0x"), "0X"))
		if decErr != nil {
			fmt.Fprintf(os.Stderr, "bad secret hex: %v\n", decErr)
			os.Exit(1)
		}
		sk, err = bls.SecretKeyFromBytes(raw)
	} else {
		sk, err = bls.RandKey()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "key: %v\n", err)
		os.Exit(1)
	}

	voteAddress := sk.PublicKey().Marshal()

	// keccak256(abi.encodePacked(operatorAddress, voteAddress, block.chainid))
	//
	// encodePacked: the address is 20 bytes, the pubkey 48, and the chain id a
	// full 32-byte word because solidity widens it to uint256. Encoding the
	// chain id in fewer bytes yields a proof that simply fails to verify, with
	// nothing on-chain to say why.
	addr := common.HexToAddress(*operator)
	chainWord := make([]byte, 32)
	new(big.Int).SetUint64(*chainID).FillBytes(chainWord)

	packed := make([]byte, 0, 20+48+32)
	packed = append(packed, addr.Bytes()...)
	packed = append(packed, voteAddress...)
	packed = append(packed, chainWord...)
	msgHash := crypto.Keccak256(packed)

	signature := sk.Sign(msgHash)
	sig := signature.Marshal()

	if len(voteAddress) != 48 || len(sig) != 96 {
		fmt.Fprintf(os.Stderr, "unexpected lengths: pubkey=%d sig=%d\n", len(voteAddress), len(sig))
		os.Exit(1)
	}

	// Verified locally before being printed. A bad proof fails on-chain too,
	// and a reverted createValidator is a slower, more confusing way to learn
	// the same thing — Verify lives on the signature, not the public key.
	if !signature.Verify(sk.PublicKey(), msgHash) {
		fmt.Fprintln(os.Stderr, "self-verification failed — refusing to emit a bad proof")
		os.Exit(1)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(output{
		Operator:    addr.Hex(),
		ChainID:     *chainID,
		VoteAddress: "0x" + hex.EncodeToString(voteAddress),
		BlsProof:    "0x" + hex.EncodeToString(sig),
		SecretKey:   "0x" + hex.EncodeToString(sk.Marshal()),
		MsgHash:     "0x" + hex.EncodeToString(msgHash),
	})
}
