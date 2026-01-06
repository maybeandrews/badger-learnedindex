/*
 * HybridFilter: Combining Bloom Filters with Learned Indexes
 *
 * This implements a hybrid approach that uses:
 * 1. Bloom Filter - to quickly determine if a key MIGHT be in the table
 * 2. Learned Index - to predict WHERE in the table to search
 *
 * The insight is that Bloom filters and Learned Indexes are complementary:
 * - Bloom Filter: Good at saying "definitely not here" (saves I/O by skipping tables)
 * - Learned Index: Good at saying "probably around position X" (narrows search range)
 *
 * Combined, we get both benefits with minimal overhead.
 */

package y

import (
	"encoding/binary"
	"math"
)

// HybridFilter combines a compact Bloom filter with a Learned Index
// for optimal SSTable lookup performance.
//
// Total size: ~64-128 bytes (configurable) vs. kilobytes for Bloom alone
type HybridFilter struct {
	// Compact Bloom filter (reduced size since we have learned index backup)
	BloomBits  []byte // Small bloom filter
	BloomHashK uint8  // Number of hash functions

	// Learned Index component
	Slope     float64
	Intercept float64
	MinErr    int32
	MaxErr    int32
	MaxPos    uint32
	KeyCount  uint32
}

// HybridFilterConfig controls the hybrid filter parameters
type HybridFilterConfig struct {
	// BloomSizeBytes controls bloom filter size (default: 64 bytes = 512 bits)
	// Smaller than traditional bloom, since learned index provides backup
	BloomSizeBytes int

	// TargetFPRate is the target false positive rate for bloom (default: 5%)
	// Higher than traditional 1% since we prioritize space efficiency
	TargetFPRate float64
}

// DefaultHybridConfig returns sensible defaults for the hybrid filter
func DefaultHybridConfig() HybridFilterConfig {
	return HybridFilterConfig{
		BloomSizeBytes: 64,   // 64 bytes = 512 bits (vs. thousands for traditional)
		TargetFPRate:   0.05, // 5% FP rate (acceptable since learned index helps)
	}
}

// HybridFilterSize returns the total size of a hybrid filter with given config
func HybridFilterSize(config HybridFilterConfig) int {
	// BloomBits + BloomHashK + Slope + Intercept + MinErr + MaxErr + MaxPos + KeyCount
	return config.BloomSizeBytes + 1 + 8 + 8 + 4 + 4 + 4 + 4
}

// TrainHybridFilter creates a hybrid filter from sorted key data
func TrainHybridFilter(keyHashes []uint32, blockIndices []uint32, numBlocks int, config HybridFilterConfig) *HybridFilter {
	if len(keyHashes) == 0 {
		return &HybridFilter{
			BloomBits:  make([]byte, config.BloomSizeBytes),
			BloomHashK: 1,
			MaxPos:     uint32(max(0, numBlocks-1)),
		}
	}

	hf := &HybridFilter{
		KeyCount: uint32(len(keyHashes)),
		MaxPos:   uint32(max(0, numBlocks-1)),
	}

	// === Build compact Bloom filter ===
	nBits := config.BloomSizeBytes * 8
	// Calculate optimal k based on size and number of keys
	// k = (m/n) * ln(2), where m = bits, n = keys
	kFloat := float64(nBits) / float64(len(keyHashes)) * 0.693
	k := uint8(max(1, min(30, int(kFloat))))
	hf.BloomHashK = k
	hf.BloomBits = make([]byte, config.BloomSizeBytes)

	// Add all keys to bloom filter
	for _, h := range keyHashes {
		delta := h>>17 | h<<15
		for j := uint8(0); j < k; j++ {
			bitPos := h % uint32(nBits)
			hf.BloomBits[bitPos/8] |= 1 << (bitPos % 8)
			h += delta
		}
	}

	// === Build Learned Index (same as before) ===
	n := len(keyHashes)

	if n == 1 {
		hf.Slope = 0
		hf.Intercept = float64(blockIndices[0])
		hf.MinErr = -1
		hf.MaxErr = 1
		return hf
	}

	// Linear regression
	var sumX, sumY, sumXY, sumX2 float64
	for i := 0; i < n; i++ {
		x := float64(keyHashes[i])
		y := float64(blockIndices[i])
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	nf := float64(n)
	denominator := nf*sumX2 - sumX*sumX

	if math.Abs(denominator) < 1e-10 {
		hf.Slope = 0
		hf.Intercept = sumY / nf
	} else {
		hf.Slope = (nf*sumXY - sumX*sumY) / denominator
		hf.Intercept = (sumY - hf.Slope*sumX) / nf
	}

	// Calculate error bounds
	var minErr, maxErr int32
	for i := 0; i < n; i++ {
		predicted := hf.Slope*float64(keyHashes[i]) + hf.Intercept
		actual := float64(blockIndices[i])
		err := int32(actual - predicted)
		if err < minErr {
			minErr = err
		}
		if err > maxErr {
			maxErr = err
		}
	}
	hf.MinErr = minErr - 1
	hf.MaxErr = maxErr + 1

	return hf
}

// MayContain returns true if the key MIGHT be in the table (Bloom filter check)
func (hf *HybridFilter) MayContain(keyHash uint32) bool {
	if hf == nil || len(hf.BloomBits) == 0 {
		return true // No filter = assume present
	}

	nBits := uint32(len(hf.BloomBits) * 8)
	h := keyHash
	delta := h>>17 | h<<15

	for j := uint8(0); j < hf.BloomHashK; j++ {
		bitPos := h % nBits
		if hf.BloomBits[bitPos/8]&(1<<(bitPos%8)) == 0 {
			return false // Definitely not present
		}
		h += delta
	}
	return true // Might be present
}

// PredictRange returns the predicted block range for a key (Learned Index)
func (hf *HybridFilter) PredictRange(keyHash uint32) (minBlock, maxBlock int) {
	if hf == nil || hf.KeyCount == 0 {
		return 0, int(hf.MaxPos)
	}

	pos := hf.Slope*float64(keyHash) + hf.Intercept
	predicted := int(math.Round(pos))

	minBlock = predicted + int(hf.MinErr)
	maxBlock = predicted + int(hf.MaxErr)

	// Clamp to valid range
	maxPosInt := int(hf.MaxPos)
	if minBlock < 0 {
		minBlock = 0
	}
	if maxBlock > maxPosInt {
		maxBlock = maxPosInt
	}

	return minBlock, maxBlock
}

// Query performs a complete hybrid lookup:
// 1. Check Bloom filter - if negative, key definitely not present
// 2. If positive, use learned index to get search range
// Returns: (maybePresent, minBlock, maxBlock)
func (hf *HybridFilter) Query(keyHash uint32) (maybePresent bool, minBlock, maxBlock int) {
	// Step 1: Bloom filter check
	if !hf.MayContain(keyHash) {
		return false, 0, 0 // Definitely not here - skip this table!
	}

	// Step 2: Learned index prediction
	minBlock, maxBlock = hf.PredictRange(keyHash)
	return true, minBlock, maxBlock
}

// Serialize converts the HybridFilter to bytes
func (hf *HybridFilter) Serialize() []byte {
	size := len(hf.BloomBits) + 1 + 8 + 8 + 4 + 4 + 4 + 4
	buf := make([]byte, size)

	offset := 0
	// Bloom filter
	copy(buf[offset:], hf.BloomBits)
	offset += len(hf.BloomBits)
	buf[offset] = hf.BloomHashK
	offset++

	// Learned index
	binary.LittleEndian.PutUint64(buf[offset:], math.Float64bits(hf.Slope))
	offset += 8
	binary.LittleEndian.PutUint64(buf[offset:], math.Float64bits(hf.Intercept))
	offset += 8
	binary.LittleEndian.PutUint32(buf[offset:], uint32(hf.MinErr))
	offset += 4
	binary.LittleEndian.PutUint32(buf[offset:], uint32(hf.MaxErr))
	offset += 4
	binary.LittleEndian.PutUint32(buf[offset:], hf.MaxPos)
	offset += 4
	binary.LittleEndian.PutUint32(buf[offset:], hf.KeyCount)

	return buf
}

// DeserializeHybridFilter reads a HybridFilter from bytes
func DeserializeHybridFilter(data []byte, bloomSize int) *HybridFilter {
	if len(data) < bloomSize+33 {
		return nil
	}

	hf := &HybridFilter{}
	offset := 0

	hf.BloomBits = make([]byte, bloomSize)
	copy(hf.BloomBits, data[offset:offset+bloomSize])
	offset += bloomSize
	hf.BloomHashK = data[offset]
	offset++

	hf.Slope = math.Float64frombits(binary.LittleEndian.Uint64(data[offset:]))
	offset += 8
	hf.Intercept = math.Float64frombits(binary.LittleEndian.Uint64(data[offset:]))
	offset += 8
	hf.MinErr = int32(binary.LittleEndian.Uint32(data[offset:]))
	offset += 4
	hf.MaxErr = int32(binary.LittleEndian.Uint32(data[offset:]))
	offset += 4
	hf.MaxPos = binary.LittleEndian.Uint32(data[offset:])
	offset += 4
	hf.KeyCount = binary.LittleEndian.Uint32(data[offset:])

	return hf
}

// Stats returns statistics about the hybrid filter
func (hf *HybridFilter) Stats() HybridFilterStats {
	return HybridFilterStats{
		TotalSizeBytes:   len(hf.BloomBits) + 33,
		BloomSizeBytes:   len(hf.BloomBits),
		LearnedSizeBytes: 33,
		BloomBits:        len(hf.BloomBits) * 8,
		BloomHashFuncs:   int(hf.BloomHashK),
		ErrorRange:       int(hf.MaxErr - hf.MinErr),
		KeyCount:         int(hf.KeyCount),
	}
}

// HybridFilterStats contains statistics about the hybrid filter
type HybridFilterStats struct {
	TotalSizeBytes   int
	BloomSizeBytes   int
	LearnedSizeBytes int
	BloomBits        int
	BloomHashFuncs   int
	ErrorRange       int
	KeyCount         int
}
