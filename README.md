# BadgerDB

[![Go Reference](https://pkg.go.dev/badge/github.com/dgraph-io/badger/v4.svg)](https://pkg.go.dev/github.com/dgraph-io/badger/v4)
[![Go Report Card](https://goreportcard.com/badge/github.com/dgraph-io/badger/v4)](https://goreportcard.com/report/github.com/dgraph-io/badger/v4)
[![Sourcegraph](https://sourcegraph.com/github.com/hypermodeinc/badger/-/badge.svg)](https://sourcegraph.com/github.com/hypermodeinc/badger?badge)
[![ci-badger-tests](https://github.com/hypermodeinc/badger/actions/workflows/ci-badger-tests.yml/badge.svg)](https://github.com/hypermodeinc/badger/actions/workflows/ci-badger-tests.yml)
[![ci-badger-bank-tests](https://github.com/hypermodeinc/badger/actions/workflows/ci-badger-bank-tests.yml/badge.svg)](https://github.com/hypermodeinc/badger/actions/workflows/ci-badger-bank-tests.yml)
[![ci-badger-bank-tests-nightly](https://github.com/hypermodeinc/badger/actions/workflows/ci-badger-bank-tests-nightly.yml/badge.svg)](https://github.com/hypermodeinc/badger/actions/workflows/ci-badger-bank-tests-nightly.yml)

![Badger mascot](images/diggy-shadow.png)

BadgerDB is an embeddable, persistent and fast key-value (KV) database written in pure Go. It is the
underlying database for [Dgraph](https://github.com/dgraph-io/dgraph), a fast, distributed graph
database. It's meant to be a performant alternative to non-Go-based key-value stores like RocksDB.

## Project Status

Badger is stable and is being used to serve data sets worth hundreds of terabytes. Badger supports
concurrent ACID transactions with serializable snapshot isolation (SSI) guarantees. A Jepsen-style
bank test runs nightly for 8h, with `--race` flag and ensures the maintenance of transactional
guarantees. Badger has also been tested to work with filesystem level anomalies, to ensure
persistence and consistency. Badger is being used by a number of projects which includes Dgraph,
Jaeger Tracing, UsenetExpress, and many more.

The list of projects using Badger can be found [here](#projects-using-badger).

Please consult the [Changelog] for more detailed information on releases.

Note: Badger is built with go 1.23 and we refrain from bumping this version to minimize downstream
effects of those using Badger in applications built with older versions of Go.

[Changelog]: https://github.com/hypermodeinc/badger/blob/main/CHANGELOG.md

## Table of Contents

- [BadgerDB](#badgerdb)
  - [Project Status](#project-status)
  - [Table of Contents](#table-of-contents)
  - [Getting Started](#getting-started)
    - [Installing](#installing)
      - [Installing Badger Command Line Tool](#installing-badger-command-line-tool)
      - [Choosing a version](#choosing-a-version)
  - [Badger Documentation](#badger-documentation)
  - [Resources](#resources)
    - [Blog Posts](#blog-posts)
  - [Design](#design)
    - [Comparisons](#comparisons)
    - [Benchmarks](#benchmarks)
  - [Projects Using Badger](#projects-using-badger)
  - [Contributing](#contributing)
  - [Contact](#contact)

## Getting Started

### Installing

To start using Badger, install Go 1.23 or above. Badger v3 and above needs go modules. From your
project, run the following command

```sh
go get github.com/dgraph-io/badger/v4
```

This will retrieve the library.

#### Installing Badger Command Line Tool

Badger provides a CLI tool which can perform certain operations like offline backup/restore. To
install the Badger CLI, retrieve the repository and checkout the desired version. Then run

```sh
cd badger
go install .
```

This will install the badger command line utility into your $GOBIN path.

## Badger Documentation

Badger Documentation is available at https://docs.hypermode.com/badger

## Resources

### Blog Posts

1. [Introducing Badger: A fast key-value store written natively in Go](https://hypermode.com/blog/badger/)
2. [Make Badger crash resilient with ALICE](https://hypermode.com/blog/alice/)
3. [Badger vs LMDB vs BoltDB: Benchmarking key-value databases in Go](https://hypermode.com/blog/badger-lmdb-boltdb/)
4. [Concurrent ACID Transactions in Badger](https://hypermode.com/blog/badger-txn/)

## Design

Badger was written with these design goals in mind:

- Write a key-value database in pure Go.
- Use latest research to build the fastest KV database for data sets spanning terabytes.
- Optimize for SSDs.

Badger’s design is based on a paper titled _[WiscKey: Separating Keys from Values in SSD-conscious
Storage][wisckey]_.

[wisckey]: https://www.usenix.org/system/files/conference/fast16/fast16-papers-lu.pdf

### Comparisons

| Feature                       | Badger                                     | RocksDB                       | BoltDB    |
| ----------------------------- | ------------------------------------------ | ----------------------------- | --------- |
| Design                        | LSM tree with value log                    | LSM tree only                 | B+ tree   |
| High Read throughput          | Yes                                        | No                            | Yes       |
| High Write throughput         | Yes                                        | Yes                           | No        |
| Designed for SSDs             | Yes (with latest research <sup>1</sup>)    | Not specifically <sup>2</sup> | No        |
| Embeddable                    | Yes                                        | Yes                           | Yes       |
| Sorted KV access              | Yes                                        | Yes                           | Yes       |
| Pure Go (no Cgo)              | Yes                                        | No                            | Yes       |
| Transactions                  | Yes, ACID, concurrent with SSI<sup>3</sup> | Yes (but non-ACID)            | Yes, ACID |
| Snapshots                     | Yes                                        | Yes                           | Yes       |
| TTL support                   | Yes                                        | Yes                           | No        |
| 3D access (key-value-version) | Yes<sup>4</sup>                            | No                            | No        |

<sup>1</sup> The [WISCKEY paper][wisckey] (on which Badger is based) saw big wins with separating
values from keys, significantly reducing the write amplification compared to a typical LSM tree.

<sup>2</sup> RocksDB is an SSD optimized version of LevelDB, which was designed specifically for
rotating disks. As such RocksDB's design isn't aimed at SSDs.

<sup>3</sup> SSI: Serializable Snapshot Isolation. For more details, see the blog post
[Concurrent ACID Transactions in Badger](https://hypermode.com/blog/badger-txn/)

<sup>4</sup> Badger provides direct access to value versions via its Iterator API. Users can also
specify how many versions to keep per key via Options.

### Benchmarks

We have run comprehensive benchmarks against RocksDB, Bolt and LMDB. The benchmarking code, and the
detailed logs for the benchmarks can be found in the [badger-bench] repo. More explanation,
including graphs can be found the blog posts (linked above).

[badger-bench]: https://github.com/dgraph-io/badger-bench

## Projects Using Badger

Below is a list of known projects that use Badger:

- [Dgraph](https://github.com/hypermodeinc/dgraph) - Distributed graph database.
- [Jaeger](https://github.com/jaegertracing/jaeger) - Distributed tracing platform.
- [go-ipfs](https://github.com/ipfs/go-ipfs) - Go client for the InterPlanetary File System (IPFS),
  a new hypermedia distribution protocol.
- [Riot](https://github.com/go-ego/riot) - An open-source, distributed search engine.
- [emitter](https://github.com/emitter-io/emitter) - Scalable, low latency, distributed pub/sub
  broker with message storage, uses MQTT, gossip and badger.
- [OctoSQL](https://github.com/cube2222/octosql) - Query tool that allows you to join, analyse and
  transform data from multiple databases using SQL.
- [Dkron](https://dkron.io/) - Distributed, fault tolerant job scheduling system.
- [smallstep/certificates](https://github.com/smallstep/certificates) - Step-ca is an online
  certificate authority for secure, automated certificate management.
- [Sandglass](https://github.com/celrenheit/sandglass) - distributed, horizontally scalable,
  persistent, time sorted message queue.
- [TalariaDB](https://github.com/grab/talaria) - Grab's Distributed, low latency time-series
  database.
- [Sloop](https://github.com/salesforce/sloop) - Salesforce's Kubernetes History Visualization
  Project.
- [Usenet Express](https://usenetexpress.com/) - Serving over 300TB of data with Badger.
- [gorush](https://github.com/appleboy/gorush) - A push notification server written in Go.
- [0-stor](https://github.com/zero-os/0-stor) - Single device object store.
- [Dispatch Protocol](https://github.com/dispatchlabs/disgo) - Blockchain protocol for distributed
  application data analytics.
- [GarageMQ](https://github.com/valinurovam/garagemq) - AMQP server written in Go.
- [RedixDB](https://alash3al.github.io/redix/) - A real-time persistent key-value store with the
  same redis protocol.
- [BBVA](https://github.com/BBVA/raft-badger) - Raft backend implementation using BadgerDB for
  Hashicorp raft.
- [Fantom](https://github.com/Fantom-foundation/go-lachesis) - aBFT Consensus platform for
  distributed applications.
- [decred](https://github.com/decred/dcrdata) - An open, progressive, and self-funding
  cryptocurrency with a system of community-based governance integrated into its blockchain.
- [OpenNetSys](https://github.com/opennetsys/c3-go) - Create useful dApps in any software language.
- [HoneyTrap](https://github.com/honeytrap/honeytrap) - An extensible and opensource system for
  running, monitoring and managing honeypots.
- [Insolar](https://github.com/insolar/insolar) - Enterprise-ready blockchain platform.
- [IoTeX](https://github.com/iotexproject/iotex-core) - The next generation of the decentralized
  network for IoT powered by scalability- and privacy-centric blockchains.
- [go-sessions](https://github.com/kataras/go-sessions) - The sessions manager for Go net/http and
  fasthttp.
- [Babble](https://github.com/mosaicnetworks/babble) - BFT Consensus platform for distributed
  applications.
- [Tormenta](https://github.com/jpincas/tormenta) - Embedded object-persistence layer / simple JSON
  database for Go projects.
- [BadgerHold](https://github.com/timshannon/badgerhold) - An embeddable NoSQL store for querying Go
  types built on Badger
- [Goblero](https://github.com/didil/goblero) - Pure Go embedded persistent job queue backed by
  BadgerDB
- [Surfline](https://www.surfline.com) - Serving global wave and weather forecast data with Badger.
- [Cete](https://github.com/mosuka/cete) - Simple and highly available distributed key-value store
  built on Badger. Makes it easy bringing up a cluster of Badger with Raft consensus algorithm by
  hashicorp/raft.
- [Volument](https://volument.com/) - A new take on website analytics backed by Badger.
- [KVdb](https://kvdb.io/) - Hosted key-value store and serverless platform built on top of Badger.
- [Terminotes](https://gitlab.com/asad-awadia/terminotes) - Self hosted notes storage and search
  server - storage powered by BadgerDB
- [Pyroscope](https://github.com/pyroscope-io/pyroscope) - Open source continuous profiling platform
  built with BadgerDB
- [Veri](https://github.com/bgokden/veri) - A distributed feature store optimized for Search and
  Recommendation tasks.
- [bIter](https://github.com/MikkelHJuul/bIter) - A library and Iterator interface for working with
  the `badger.Iterator`, simplifying from-to, and prefix mechanics.
- [ld](https://github.com/MikkelHJuul/ld) - (Lean Database) A very simple gRPC-only key-value
  database, exposing BadgerDB with key-range scanning semantics.
- [Souin](https://github.com/darkweak/Souin) - A RFC compliant HTTP cache with lot of other features
  based on Badger for the storage. Compatible with all existing reverse-proxies.
- [Xuperchain](https://github.com/xuperchain/xupercore) - A highly flexible blockchain architecture
  with great transaction performance.
- [m2](https://github.com/qichengzx/m2) - A simple http key/value store based on the raft protocol.
- [chaindb](https://github.com/ChainSafe/chaindb) - A blockchain storage layer used by
  [Gossamer](https://chainsafe.github.io/gossamer/), a Go client for the
  [Polkadot Network](https://polkadot.network/).
- [vxdb](https://github.com/vitalvas/vxdb) - Simple schema-less Key-Value NoSQL database with
  simplest API interface.
- [Opacity](https://github.com/opacity/storage-node) - Backend implementation for the Opacity
  storage project
- [Vephar](https://github.com/vaccovecrana/vephar) - A minimal key/value store using hashicorp-raft
  for cluster coordination and Badger for data storage.
- [gowarcserver](https://github.com/nlnwa/gowarcserver) - Open-source server for warc files. Can be
  used in conjunction with pywb
- [flow-go](https://github.com/onflow/flow-go) - A fast, secure, and developer-friendly blockchain
  built to support the next generation of games, apps and the digital assets that power them.
- [Wrgl](https://www.wrgl.co) - A data version control system that works like Git but specialized to
  store and diff CSV.
- [Loggie](https://github.com/loggie-io/loggie) - A lightweight, cloud-native data transfer agent
  and aggregator.
- [raft-badger](https://github.com/rfyiamcool/raft-badger) - raft-badger implements LogStore and
  StableStore Interface of hashcorp/raft. it is used to store raft log and metadata of
  hashcorp/raft.
- [DVID](https://github.com/janelia-flyem/dvid) - A dataservice for branched versioning of a variety
  of data types. Originally created for large-scale brain reconstructions in Connectomics.
- [KVS](https://github.com/tauraamui/kvs) - A library for making it easy to persist, load and query
  full structs into BadgerDB, using an ownership hierarchy model.
- [LLS](https://github.com/Boc-chi-no/LLS) - LLS is an efficient URL Shortener that can be used to
  shorten links and track link usage. Support for BadgerDB and MongoDB. Improved performance by more
  than 30% when using BadgerDB
- [lakeFS](https://github.com/treeverse/lakeFS) - lakeFS is an open-source data version control that
  transforms your object storage to Git-like repositories. lakeFS uses BadgerDB for its underlying
  local metadata KV store implementation
- [Goptivum](https://github.com/smegg99/Goptivum) - Goptivum is a better frontend and API for the
  Vulcan Optivum schedule program
- [ActionManager](https://mftlabs.io/actionmanager) - A dynamic entity manager based on rjsf schema
  and badger db
- [MightyMap](https://github.com/thisisdevelopment/mightymap) - Mightymap: Conveys both robustness
  and high capability, fitting for a powerful concurrent map.
- [FlowG](https://github.com/link-society/flowg) - A low-code log processing facility
- [Bluefin](https://github.com/blinklabs-io/bluefin) - Bluefin is a TUNA Proof of Work miner for the
  Fortuna smart contract on the Cardano blockchain
- [cDNSd](https://github.com/blinklabs-io/cdnsd) - A Cardano blockchain backed DNS server daemon
- [Dingo](https://github.com/blinklabs-io/dingo) - A Cardano blockchain data node

If you are using Badger in a project please send a pull request to add it to the list.

## Contributing

If you're interested in contributing to Badger see [CONTRIBUTING](./CONTRIBUTING.md).

## Contact

- Please use [Github issues](https://github.com/hypermodeinc/badger/issues) for filing bugs.
- Please use [discuss.hypermode.com](https://discuss.hypermode.com) for questions, discussions, and
  feature requests.
- Follow us on Twitter [@hypermodeinc](https://twitter.com/hypermodeinc).

---

# Learned Index Research Project

## Paper Title

**"When Learned Indexes Fail: A Case Study of Hash-Based Key Access in LSM-Tree Storage"**

## Research Overview

This project investigates replacing Bloom filters with learned indexes in BadgerDB, an LSM-tree
based key-value store. Our research reveals a **critical finding**: learned indexes fundamentally
require key ordering to be effective, but many databases (including BadgerDB) use hash-based key
access patterns that destroy this ordering.

### Research Question

Can learned indexes replace Bloom filters in LSM-tree databases that use hash-based key lookups?

### Key Finding

**Learned indexes fail when keys are accessed via hash values.** This is because:

- Bloom filters work with `Hash(key)` for membership testing
- Hashing destroys key ordering (adjacent keys have completely different hash values)
- Linear regression cannot find patterns in random/hashed data
- Search range becomes 100% of the table (no benefit)

However, when keys are accessed in **sorted order**, learned indexes achieve only **3% search
range** - a 33x improvement!

### The Solution

We discovered the fix: **use key insertion position instead of hash values** for training. See
`y/solution_test.go` for the demonstration.

📖 **For detailed technical documentation, see
[TECHNICAL_DOCUMENTATION.md](TECHNICAL_DOCUMENTATION.md)**

---

## Quick Start

```bash
# Navigate to project
cd /Users/andrews/Desktop/badger-learnedindex

# Build to verify no errors
go build ./...

# Run the MAIN paper contribution test (shows hash problem)
go test -v -run TestPaperContribution ./y/

# Run the SOLUTION test (shows how to fix it)
go test -v -run TestLearnedIndexWithKeyPosition ./y/

# Run all paper-related tests
go test -v -run "TestPaper|TestDataDistribution|TestBloomVsLearned|TestSolution" ./y/

# Run comparison benchmarks
go test -v -run TestCompareLearnedIndexVsBloomFilter ./y/

# Run benchmarks with memory info
go test -bench=BenchmarkCompare -benchmem ./y/

# Verify nothing is broken
go test ./... -short
```

---

## Files Created

| File                                   | Description                                                   |
| -------------------------------------- | ------------------------------------------------------------- |
| `y/learned_index.go`                   | Core learned index implementation (linear regression model)   |
| `y/learned_index_test.go`              | Unit tests for learned index                                  |
| `y/learned_vs_bloom_benchmark_test.go` | Comparison benchmarks (Bloom vs Learned)                      |
| `y/hybrid_filter.go`                   | Hybrid Filter combining Bloom + Learned Index                 |
| `y/hybrid_filter_test.go`              | Hybrid filter comparison tests                                |
| `y/paper_contribution_test.go`         | **Main paper tests** - demonstrates when learned indexes fail |
| `y/compact_hybrid_test.go`             | Bloom filter size/accuracy trade-off analysis                 |
| `y/solution_test.go`                   | **Solution tests** - shows how to fix the hash problem        |
| `TECHNICAL_DOCUMENTATION.md`           | Detailed technical documentation of all changes               |

## Files Modified

| File                    | Changes                                                                            |
| ----------------------- | ---------------------------------------------------------------------------------- |
| `table/builder.go`      | Added `keyBlockIndices` field, trains learned index instead of Bloom filter        |
| `table/table.go`        | Added `learnedIndex` field, `PredictBlockRange()` method, modified `DoesNotHave()` |
| `table/builder_test.go` | Updated test expectations for new semantics                                        |
| `table/table_test.go`   | Updated race test for new semantics                                                |

---

## Key Benchmark Results

### Storage Comparison (100,000 keys)

| Metric        | Bloom Filter | Learned Index | Winner                            |
| ------------- | ------------ | ------------- | --------------------------------- |
| Storage Size  | 87,501 bytes | 32 bytes      | ✅ Learned Index (2,734x smaller) |
| Build Time    | 444 µs       | 192 µs        | ✅ Learned Index (2.3x faster)    |
| Lookup Time   | ~15 ns       | ~9 ns         | ✅ Learned Index (1.6x faster)    |
| Memory Allocs | 1 alloc      | 1 alloc       | Tie                               |

### Search Range by Access Pattern

| Access Pattern             | Search Range   | Effectiveness       |
| -------------------------- | -------------- | ------------------- |
| **Sorted keys**            | 3.0% of table  | ✅ Excellent        |
| **Hashed keys** (BadgerDB) | 100% of table  | ❌ Complete failure |
| **Random/shuffled**        | 100% of table  | ❌ Complete failure |
| **Clustered (80/20)**      | 66.3% of table | ⚠️ Partial          |

---

## Technical Implementation

### Learned Index Model

```
block_index = slope × hash(key) + intercept
```

**Stored parameters (32 bytes total):**

- `Slope` (float64): 8 bytes
- `Intercept` (float64): 8 bytes
- `MinErr` (int32): 4 bytes - lower bound correction
- `MaxErr` (int32): 4 bytes - upper bound correction
- `KeyCount` (uint32): 4 bytes
- `MaxPos` (uint32): 4 bytes - maximum block index

### Prediction Function

```go
func (li *LearnedIndex) Predict(keyHash uint32) (predictedBlock, minBlock, maxBlock int) {
    predicted := li.Slope*float64(keyHash) + li.Intercept
    predictedBlock = int(predicted)
    minBlock = predictedBlock + li.MinErr  // Apply error bounds
    maxBlock = predictedBlock + li.MaxErr
    // Clamp to valid range [0, MaxPos]
    return
}
```

### Hybrid Filter (Novel Contribution)

Combines Bloom filter (for table skipping) with learned index (for position prediction):

```go
type HybridFilter struct {
    // Bloom filter component (can skip tables)
    BloomBits []byte
    BloomHashK uint8

    // Learned index component (narrows search)
    Slope, Intercept float64
    MinErr, MaxErr   int32
    MaxPos           uint32
}
```

---

## Paper Contribution: Why Hash Breaks Learned Indexes

### The Problem

```
Hash values for sequential keys:
  Key 0:    key_0000000000  hash=2795452986
  Key 1:    key_0000000001  hash=1262931415
  Key 2:    key_0000000002  hash=4025376883
  Key 100:  key_0000000100  hash=3512541005
  Key 101:  key_0000000101  hash=1980019434

Notice: Adjacent keys have COMPLETELY DIFFERENT hash values!
```

### Statistical Evidence

```
Correlation (sorted position → block): 0.9999 (near perfect)
Correlation (hash value → block):      0.0000 (essentially random)
```

### Key Insight

The LearnedKV paper assumes sorted key access, but many real databases use hash-based access
patterns. This is a fundamental limitation of learned indexes that is **underexplored in the
literature**.

---

## Trade-off Analysis

| Aspect                     | Bloom Filter                           | Learned Index                           |
| -------------------------- | -------------------------------------- | --------------------------------------- |
| **Purpose**                | "Is key definitely NOT in this table?" | "Where in this table might the key be?" |
| **Can skip tables**        | ✅ Yes (if key not found)              | ❌ No                                   |
| **Narrows search range**   | ❌ No                                  | ✅ Yes (with sorted keys)               |
| **False positives**        | Yes (~1%)                              | N/A (always says "might be here")       |
| **False negatives**        | No                                     | Possible if error bounds exceeded       |
| **Works with hash(key)**   | ✅ Yes                                 | ❌ No                                   |
| **Works with sorted keys** | ✅ Yes                                 | ✅ Yes                                  |

---

## Running Paper Tests

### Main Paper Contribution Test

```bash
go test -v -run TestPaperContribution ./y/
```

Shows the dramatic difference between sorted (3%) and hashed (100%) access patterns.

### Data Distribution Impact

```bash
go test -v -run TestDataDistributionImpact ./y/
```

Analyzes how different data distributions affect learned index effectiveness.

### Bloom vs Learned Trade-offs

```bash
go test -v -run TestBloomVsLearnedTradeoffs ./y/
```

Feature comparison table between both approaches.

### Hybrid Filter Comparison

```bash
go test -v -run TestHybridFilterComparison ./y/
```

Compares Bloom-only, Learned-only, and Hybrid approaches.

### Bloom Size Trade-off

```bash
go test -v -run TestBloomSizeTradeoff ./y/
```

Analyzes Bloom filter size vs false positive rate.

---

## Conclusions for Paper

### Finding 1: Learned Indexes Require Key Ordering

- When keys are accessed in sorted order: **3% search range** (excellent)
- When keys are accessed by hash: **100% search range** (useless)

### Finding 2: Bloom Filters Remain Essential

- Bloom filters work regardless of key ordering
- They can definitively say "key NOT present" (skip table entirely)
- Learned indexes can only say "key MIGHT be here" (cannot skip)

### Finding 3: Hybrid Approach Has Limited Benefit

- Combining Bloom + Learned only helps when learned index is effective
- With hash-based access, the learned component provides no benefit

### Finding 4: Storage Advantage is Real

- Learned index: 32 bytes vs Bloom filter: 87,501 bytes (100K keys)
- But storage savings are meaningless if effectiveness is lost

### Practical Recommendation

- Keep Bloom filters for table-level filtering in hash-based systems
- Only use learned indexes if you have **sorted key access patterns**
- Consider the access pattern before applying learned indexes

---

## References

1. Kraska et al., "The Case for Learned Index Structures" (2018)
2. Cai et al., "LearnedKV: Integrating LSM-Trees with Learned Indexes" (2024) - arXiv:2406.18892
3. Lu et al., "WiscKey: Separating Keys from Values in SSD-conscious Storage" (2016)

---

## Project Structure

```
badger-learnedindex/
├── y/
│   ├── learned_index.go              # Core learned index
│   ├── learned_index_test.go         # Unit tests
│   ├── learned_vs_bloom_benchmark_test.go  # Comparison benchmarks
│   ├── hybrid_filter.go              # Hybrid Bloom + Learned
│   ├── hybrid_filter_test.go         # Hybrid tests
│   ├── paper_contribution_test.go    # MAIN PAPER TESTS
│   ├── compact_hybrid_test.go        # Size analysis
│   └── bloom.go                      # Original Bloom filter
├── table/
│   ├── builder.go                    # Modified: trains learned index
│   ├── table.go                      # Modified: uses learned index
│   └── ...
└── README.md                         # This file
```
