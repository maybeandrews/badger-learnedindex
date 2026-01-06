package y

import (
	"fmt"
	"testing"
)

// TestWalkthroughBloomVsLearned shows step-by-step how both work
func TestWalkthroughBloomVsLearned(t *testing.T) {
	fmt.Print(`
╔══════════════════════════════════════════════════════════════════════╗
║           BLOOM FILTER vs LEARNED INDEX WALKTHROUGH                  ║
╚══════════════════════════════════════════════════════════════════════╝

INPUT: 5 keys stored in an SSTable with 5 blocks
`)

	// Our input keys
	keys := []string{"apple", "banana", "cherry", "date", "elderberry"}

	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("  STEP 1: OUR INPUT DATA")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("  Key            Block    Hash(Key)")
	fmt.Println("  ─────────────  ─────    ──────────")

	keyHashes := make([]uint32, len(keys))
	blockIndices := make([]uint32, len(keys))

	for i, key := range keys {
		h := Hash([]byte(key))
		keyHashes[i] = uint32(h)
		blockIndices[i] = uint32(i) // Each key in its own block
		fmt.Printf("  %-12s   %d        %d\n", key, i, h)
	}

	fmt.Print(`
═══════════════════════════════════════════════════════════════
  STEP 2: BUILDING THE BLOOM FILTER
═══════════════════════════════════════════════════════════════

  Bloom filter = bit array where we "mark" each key's presence
  
  For each key, we compute multiple hash positions and set those bits to 1.
`)

	// Create bloom filter (small for demo - 50 bits)
	bitsPerKey := 10
	numBits := len(keys) * bitsPerKey        // 50 bits
	bloomBits := make([]byte, (numBits+7)/8) // 7 bytes

	fmt.Printf("\n  Bloom filter size: %d bits (%d bytes)\n\n", numBits, len(bloomBits))

	// Add each key to bloom filter
	for _, key := range keys {
		h := uint64(Hash([]byte(key)))
		// Simulate 3 hash functions
		pos1 := h % uint64(numBits)
		pos2 := (h * 31) % uint64(numBits)
		pos3 := (h * 97) % uint64(numBits)

		// Set bits
		bloomBits[pos1/8] |= 1 << (pos1 % 8)
		bloomBits[pos2/8] |= 1 << (pos2 % 8)
		bloomBits[pos3/8] |= 1 << (pos3 % 8)

		fmt.Printf("  Adding '%s':\n", key)
		fmt.Printf("    hash = %d\n", h)
		fmt.Printf("    Set bit %d (hash %% %d)\n", pos1, numBits)
		fmt.Printf("    Set bit %d ((hash×31) %% %d)\n", pos2, numBits)
		fmt.Printf("    Set bit %d ((hash×97) %% %d)\n", pos3, numBits)
		fmt.Println()
	}

	// Print final bit array
	fmt.Print("  Final Bloom filter bits: [")
	for i := 0; i < numBits; i++ {
		if bloomBits[i/8]&(1<<(i%8)) != 0 {
			fmt.Print("1")
		} else {
			fmt.Print("0")
		}
	}
	fmt.Println("]")

	fmt.Print(`
═══════════════════════════════════════════════════════════════
  STEP 3: BUILDING THE LEARNED INDEX
═══════════════════════════════════════════════════════════════

  Learned index = find a linear formula: block = slope × hash + intercept
  
  We use linear regression (least squares) to find the best line.
`)

	// Train learned index
	li := TrainLearnedIndex(keyHashes, blockIndices, 5)

	fmt.Println("\n  Training data (hash → block):")
	for i, key := range keys {
		fmt.Printf("    %s: hash=%d → block=%d\n", key, keyHashes[i], blockIndices[i])
	}

	fmt.Println()
	fmt.Println("  Linear regression result:")
	fmt.Printf("    slope     = %.10f\n", li.Slope)
	fmt.Printf("    intercept = %.4f\n", li.Intercept)
	fmt.Printf("    Formula: block = %.10f × hash + (%.4f)\n", li.Slope, li.Intercept)
	fmt.Println()
	fmt.Printf("  Error bounds: [%d, %d]\n", li.MinErr, li.MaxErr)
	fmt.Printf("  Total size: 32 bytes (constant!)\n")

	fmt.Print(`
═══════════════════════════════════════════════════════════════
  STEP 4: LOOKUP - Key EXISTS ("cherry")
═══════════════════════════════════════════════════════════════
`)

	lookupKey := "cherry"
	lookupHash := uint64(Hash([]byte(lookupKey)))

	fmt.Printf("\n  Looking for: '%s'\n", lookupKey)
	fmt.Printf("  Hash('%s') = %d\n\n", lookupKey, lookupHash)

	// Bloom filter lookup
	fmt.Println("  ┌─────────────────────────────────────────┐")
	fmt.Println("  │ BLOOM FILTER LOOKUP                     │")
	fmt.Println("  └─────────────────────────────────────────┘")

	pos1 := lookupHash % uint64(numBits)
	pos2 := (lookupHash * 31) % uint64(numBits)
	pos3 := (lookupHash * 97) % uint64(numBits)

	bit1 := bloomBits[pos1/8]&(1<<(pos1%8)) != 0
	bit2 := bloomBits[pos2/8]&(1<<(pos2%8)) != 0
	bit3 := bloomBits[pos3/8]&(1<<(pos3%8)) != 0

	fmt.Printf("    Check bit %d: %v\n", pos1, bit1)
	fmt.Printf("    Check bit %d: %v\n", pos2, bit2)
	fmt.Printf("    Check bit %d: %v\n", pos3, bit3)

	if bit1 && bit2 && bit3 {
		fmt.Println("    Result: ALL bits are 1 → Key MIGHT exist")
		fmt.Println("    Action: Search ALL 5 blocks")
	} else {
		fmt.Println("    Result: Some bit is 0 → Key definitely NOT here")
		fmt.Println("    Action: Skip this SSTable entirely!")
	}

	// Learned index lookup
	fmt.Println()
	fmt.Println("  ┌─────────────────────────────────────────┐")
	fmt.Println("  │ LEARNED INDEX LOOKUP                    │")
	fmt.Println("  └─────────────────────────────────────────┘")

	predicted, minBlock, maxBlock := li.Predict(uint32(lookupHash))

	fmt.Printf("    predicted = %.10f × %d + (%.4f)\n", li.Slope, lookupHash, li.Intercept)
	fmt.Printf("    predicted = %.2f → block %d\n", li.Slope*float64(lookupHash)+li.Intercept, predicted)
	fmt.Printf("    With error bounds [%d, %d]:\n", li.MinErr, li.MaxErr)
	fmt.Printf("    Search range: blocks %d to %d\n", minBlock, maxBlock)
	fmt.Printf("    Action: Search only %d blocks (out of 5)\n", maxBlock-minBlock+1)

	fmt.Print(`
═══════════════════════════════════════════════════════════════
  STEP 5: LOOKUP - Key DOES NOT EXIST ("zebra")
═══════════════════════════════════════════════════════════════
`)

	lookupKey = "zebra"
	lookupHash = uint64(Hash([]byte(lookupKey)))

	fmt.Printf("\n  Looking for: '%s'\n", lookupKey)
	fmt.Printf("  Hash('%s') = %d\n\n", lookupKey, lookupHash)

	// Bloom filter lookup
	fmt.Println("  ┌─────────────────────────────────────────┐")
	fmt.Println("  │ BLOOM FILTER LOOKUP                     │")
	fmt.Println("  └─────────────────────────────────────────┘")

	pos1 = lookupHash % uint64(numBits)
	pos2 = (lookupHash * 31) % uint64(numBits)
	pos3 = (lookupHash * 97) % uint64(numBits)

	bit1 = bloomBits[pos1/8]&(1<<(pos1%8)) != 0
	bit2 = bloomBits[pos2/8]&(1<<(pos2%8)) != 0
	bit3 = bloomBits[pos3/8]&(1<<(pos3%8)) != 0

	fmt.Printf("    Check bit %d: %v\n", pos1, bit1)
	fmt.Printf("    Check bit %d: %v\n", pos2, bit2)
	fmt.Printf("    Check bit %d: %v\n", pos3, bit3)

	if bit1 && bit2 && bit3 {
		fmt.Println("    Result: ALL bits are 1 → Key MIGHT exist (FALSE POSITIVE!)")
		fmt.Println("    Action: Search ALL 5 blocks (wasted work)")
	} else {
		fmt.Println("    Result: Some bit is 0 → Key definitely NOT here ✓")
		fmt.Println("    Action: Skip this SSTable entirely! (saved work)")
	}

	// Learned index lookup
	fmt.Println()
	fmt.Println("  ┌─────────────────────────────────────────┐")
	fmt.Println("  │ LEARNED INDEX LOOKUP                    │")
	fmt.Println("  └─────────────────────────────────────────┘")

	predicted, minBlock, maxBlock = li.Predict(uint32(lookupHash))

	fmt.Printf("    predicted = %.10f × %d + (%.4f)\n", li.Slope, lookupHash, li.Intercept)
	fmt.Printf("    predicted = %.2f → block %d\n", li.Slope*float64(lookupHash)+li.Intercept, predicted)
	fmt.Printf("    Search range: blocks %d to %d\n", minBlock, maxBlock)
	fmt.Printf("    Action: Search %d blocks, then find key is NOT there\n", maxBlock-minBlock+1)
	fmt.Println("    (Learned index cannot skip tables - it only narrows search)")

	fmt.Print(`
═══════════════════════════════════════════════════════════════
  SUMMARY: BLOOM FILTER vs LEARNED INDEX
═══════════════════════════════════════════════════════════════

  ┌────────────────┬─────────────────────┬─────────────────────┐
  │ Feature        │ Bloom Filter        │ Learned Index       │
  ├────────────────┼─────────────────────┼─────────────────────┤
  │ Storage        │ ~10 bits per key    │ 32 bytes (constant) │
  │ (5 keys)       │ 50 bits = 7 bytes   │ 32 bytes            │
  │ (100K keys)    │ 87,500 bytes        │ 32 bytes            │
  ├────────────────┼─────────────────────┼─────────────────────┤
  │ Answer type    │ "Maybe" or "No"     │ "Search blocks X-Y" │
  │ Can skip table │ YES (if "No")       │ NO (always search)  │
  │ Narrows search │ NO (all or nothing) │ YES (range)         │
  ├────────────────┼─────────────────────┼─────────────────────┤
  │ Best for       │ Any key pattern     │ Sequential keys     │
  │ Worst for      │ Nothing             │ Random/hashed keys  │
  └────────────────┴─────────────────────┴─────────────────────┘
`)
}
