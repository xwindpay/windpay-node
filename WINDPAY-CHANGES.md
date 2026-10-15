# Windpay changes to this node source

Forked from bnb-chain/bsc at commit `c5533ab5b7244dc474add10740834417a2c605d7`
(2026-08-13). GPL-3.0 requires a modified version to state what was changed.
This file is that statement.

## 1. BreatheBlockInterval — params/protocol_params.go

    - BreatheBlockInterval uint64 = 24 * 3600
    + BreatheBlockInterval uint64 = 600

Upstream recalculates the validator set once per UTC day, which makes every
staking question on a testnet take a day to answer. Shortened to 10 minutes so a
set update can be observed directly. Only the trigger interval changes — the
election logic, staking rules and slashing thresholds are untouched.

**A mainnet build must restore `24 * 3600`.** Images built from this source are
tagged `:testnet-fastbreathe` so the difference cannot be overlooked.

## 2. cmd/blsproof — new command

Produces a BLS proof-of-possession for a validator's vote address in the form
`StakeHub.createValidator` expects.

## 3. BSC/BNB renamed to Windpay/WIND

The chain is Windpay, so the source says Windpay. 580 Windpay identifiers now
exist where BSC/BNB names were.

### Files and directories renamed

    eth/protocols/bsc/                      -> eth/protocols/windpay/
    eth/handler_bsc.go                      -> eth/handler_windpay.go
    eth/handler_bsc_test.go                 -> eth/handler_windpay_test.go
    core/filtermaps/checkpoints_bsc_mainnet.json -> checkpoints_windpay_mainnet.json
    core/filtermaps/checkpoints_bsc_chapel.json  -> checkpoints_windpay_chapel.json
    tests/truffle/config/config-bsc-rpc.toml     -> config-windpay-rpc.toml
    tests/truffle/scripts/bsc-rpc.sh             -> windpay-rpc.sh
    tests/truffle/scripts/bsc-validator.sh       -> windpay-validator.sh

### Deleted

    CHANGELOG.md    BNB Chain's own release history, not this project's.

## What was deliberately NOT renamed, and why

### The wire protocol name

    eth/protocols/windpay/protocol.go:  const ProtocolName = "bsc"

This string is exchanged between peers to agree they speak the same protocol. A
node built with a different value cannot peer with the running validators, which
would halt the chain. The package was renamed; this one string was not.

### StakeHub ABI identifiers

    consensus/parlia/abi.go:  PooledBNB, minSelfDelegationBNB, minDelegationBNBChange

These are function and field names in the deployed contract's ABI. Renaming them
changes the call signatures the node sends, and every read from StakeHub would
fail. The same names are kept for the same reason in the contracts repository.

### Upstream URLs

226 references to `github.com/bnb-chain/...` are left intact. They point at real
upstream repositories and are required attribution under GPL-3.0.

An earlier pass of this rename mangled them — `bsc-genesis-contract` became
`windpay-genesis-contract`, a repository that does not exist. All 155 were
restored. Protect whole URLs, not just the organisation name.

### Words that merely contain the letters "bsc"

1853 occurrences, none of them BSC references:

    Subscribe, Subscription, unsubscribe   su-bscr-ibe
    subscoping                             su-bsc-oping
    BlobSchedule, BlobScheduleConfig       Blob-Sc-hedule
    dbscheme                               d-bsc-heme
    vbscript                               v-bsc-ript
    bsCount                                bs-Count
    BscScan                                a real external product

An earlier pass collapsed `subscri` and `Subscri` to a single case, turning
`Unsubscribe` into `UnSubscribe`. The build caught it; the tree was restored from
git and the rename redone with case-separate protection.

### Files holding encoded data, not text

    core/genesis_alloc.go              encoded genesis allocation blobs
    crypto/secp256k1/.../wycheproof/   crypto test vectors
    crypto/keccak/testdata/            compressed test data
    graphql/.../*.min.js, *.min.css    minified third-party bundles
    signer/fourbyte/4byte.json         third-party signature database

Editing these would corrupt them. `core/genesis_alloc.go` needed its two
variable names changed so the code still compiled; only the identifiers were
touched, and the encoded payload was verified byte-identical afterwards
(md5 `57d1624f031b2bcaa3bb1ab6ea7cef7b` before and after).

## Verification

Done in a `golang:1.25-alpine` container, never on the validator host.

| Check | Result |
|---|---|
| Baseline build before any change | pass |
| `go build ./...` after rename | pass |
| `go run build/ci.go install -static ./cmd/geth` | pass, 88 MB binary |
| `geth version` runs | pass |
| Genesis payload md5 unchanged | pass |
| `ProtocolName` still `"bsc"` | pass |
| `Unsubscribe` not corrupted | pass |
| Chain kept producing throughout | pass, all 4 nodes in sync |

The running validators were never touched. They execute the Docker image
`windpay/node:testnet-fastbreathe`, which was built before this rename. Nothing
in this directory is read by a running node.

## Rebuilding the image from this source

    docker build -f docker-images/node.Dockerfile \
      -t windpay/node:testnet-fastbreathe2 source/

Do not retag over the image the validators currently run until the new binary
has been tested on a node that is not sealing.
