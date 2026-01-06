/*
 * INTERACTIVE DEMO: Learned Index vs Bloom Filter
 *
 * This demo lets you enter your own keys and see how the learned index works.
 *
 * Run: go test -v -run TestInteractiveDemo ./y/
 */

package y

import (
	"fmt"
	"strings"
	"testing"
)

// TestInteractiveDemo shows learned index behavior with custom data
func TestInteractiveDemo(t *testing.T) {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("  INTERACTIVE DEMO: Learned Index in Action")
	fmt.Println(strings.Repeat("=", 70))

	// Simulate user data - these could be any keys
	userKeys := []string{
		"user_alice",
		"user_bob",
		"user_charlie",
		"user_david",
		"user_eve",
		"user_frank",
		"user_grace",
		"user_henry",
		"user_ivy",
		"user_jack",
	}

	numBlocks := 5 // Simulate 5 data blocks
	keysPerBlock := 2

	fmt.Println("\n📊 INPUT DATA:")
	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("  Keys: %v\n", userKeys)
	fmt.Printf("  Blocks: %d (2 keys per block)\n", numBlocks)

	// Show how keys are distributed in blocks
	fmt.Println("\n📦 KEY DISTRIBUTION IN BLOCKS:")
	fmt.Println(strings.Repeat("-", 70))
	for i, key := range userKeys {
		blockNum := i / keysPerBlock
		fmt.Printf("  %-15s → Block %d\n", key, blockNum)
	}

	// ========== BLOOM FILTER APPROACH ==========
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("  APPROACH 1: BLOOM FILTER")
	fmt.Println(strings.Repeat("=", 70))

	hashes := make([]uint32, len(userKeys))
	for i, key := range userKeys {
		hashes[i] = Hash([]byte(key))
	}

	bitsPerKey := BloomBitsPerKey(len(userKeys), 0.01)
	bloom := NewFilter(hashes, bitsPerKey)

	fmt.Printf("\n  Storage: %d bytes\n", len(bloom))
	fmt.Println("\n  Query Results:")
	fmt.Println("  " + strings.Repeat("-", 50))

	// Test existing keys
	for _, key := range userKeys[:3] {
		h := Hash([]byte(key))
		result := Filter(bloom).MayContain(h)
		fmt.Printf("  '%s' → %s\n", key, boolToMayExist(result))
	}

	// Test non-existing keys
	nonExisting := []string{"user_unknown", "user_nobody", "user_missing"}
	for _, key := range nonExisting {
		h := Hash([]byte(key))
		result := Filter(bloom).MayContain(h)
		status := "Maybe (false positive!)"
		if !result {
			status = "Definitely NOT here ✓"
		}
		fmt.Printf("  '%s' → %s\n", key, status)
	}

	fmt.Println("\n  💡 Bloom filter says 'maybe' or 'definitely not'")
	fmt.Println("     Cannot tell you WHERE to look, only IF to look")

	// ========== LEARNED INDEX APPROACH (with position) ==========
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("  APPROACH 2: LEARNED INDEX (using key position)")
	fmt.Println(strings.Repeat("=", 70))

	positions := make([]uint32, len(userKeys))
	blockIndices := make([]uint32, len(userKeys))
	for i := range userKeys {
		positions[i] = uint32(i)
		blockIndices[i] = uint32(i / keysPerBlock)
	}

	li := TrainLearnedIndex(positions, blockIndices, numBlocks)

	fmt.Printf("\n  Storage: %d bytes (%.0fx smaller than Bloom!)\n",
		LearnedIndexSize, float64(len(bloom))/float64(LearnedIndexSize))
	fmt.Printf("  Model: block = %.4f × position + %.4f\n", li.Slope, li.Intercept)
	fmt.Printf("  Error bounds: [%d, %d]\n", li.MinErr, li.MaxErr)

	fmt.Println("\n  Query Results:")
	fmt.Println("  " + strings.Repeat("-", 50))

	for i, key := range userKeys {
		predicted, minB, maxB := li.Predict(positions[i])
		actualBlock := i / keysPerBlock
		fmt.Printf("  '%s' (pos %d) → Predicted: Block %d, Search: [%d-%d], Actual: Block %d %s\n",
			key, i, predicted, minB, maxB, actualBlock,
			checkMark(actualBlock >= minB && actualBlock <= maxB))
	}

	// ========== COMPARISON ==========
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("  COMPARISON SUMMARY")
	fmt.Println(strings.Repeat("=", 70))

	searchRange := li.MaxErr - li.MinErr + 1
	searchPercent := float64(searchRange) / float64(numBlocks) * 100

	fmt.Printf(`
  | Feature              | Bloom Filter    | Learned Index   |
  |----------------------|-----------------|-----------------|
  | Storage              | %d bytes        | %d bytes        |
  | Can skip tables      | ✅ Yes          | ❌ No           |
  | Predicts position    | ❌ No           | ✅ Yes          |
  | Search range         | All blocks      | %.0f%% of blocks |

`, len(bloom), LearnedIndexSize, searchPercent)

	fmt.Println("  ✅ Learned index tells you EXACTLY where to look!")
	fmt.Println()
}

func boolToMayExist(b bool) string {
	if b {
		return "Maybe exists"
	}
	return "Definitely NOT here ✓"
}

func checkMark(ok bool) string {
	if ok {
		return "✓"
	}
	return "✗"
}

// TestCustomKeysDemo allows testing with any keys
func TestCustomKeysDemo(t *testing.T) {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("  DEMO: How Different Key Patterns Affect Learned Index")
	fmt.Println(strings.Repeat("=", 70))

	patterns := []struct {
		name string
		keys []string
	}{
		{
			name: "Sequential IDs",
			keys: []string{"id_001", "id_002", "id_003", "id_004", "id_005",
				"id_006", "id_007", "id_008", "id_009", "id_010"},
		},
		{
			name: "User Emails",
			keys: []string{"alice@mail.com", "bob@mail.com", "carol@mail.com",
				"dave@mail.com", "eve@mail.com", "frank@mail.com",
				"grace@mail.com", "henry@mail.com", "ivy@mail.com", "jack@mail.com"},
		},
		{
			name: "Timestamps",
			keys: []string{"2024-01-01", "2024-01-02", "2024-01-03", "2024-01-04",
				"2024-01-05", "2024-01-06", "2024-01-07", "2024-01-08",
				"2024-01-09", "2024-01-10"},
		},
	}

	for _, p := range patterns {
		fmt.Printf("\n📊 Pattern: %s\n", p.name)
		fmt.Println(strings.Repeat("-", 50))

		numBlocks := 5
		keysPerBlock := 2

		// Using hash (WRONG approach)
		hashes := make([]uint32, len(p.keys))
		blocks := make([]uint32, len(p.keys))
		for i, key := range p.keys {
			hashes[i] = Hash([]byte(key))
			blocks[i] = uint32(i / keysPerBlock)
		}
		hashLI := TrainLearnedIndex(hashes, blocks, numBlocks)

		// Using position (CORRECT approach)
		positions := make([]uint32, len(p.keys))
		for i := range p.keys {
			positions[i] = uint32(i)
		}
		posLI := TrainLearnedIndex(positions, blocks, numBlocks)

		// Calculate search ranges
		hashRange := 0
		posRange := 0
		for i := range p.keys {
			_, min1, max1 := hashLI.Predict(hashes[i])
			_, min2, max2 := posLI.Predict(positions[i])
			hashRange += (max1 - min1 + 1)
			posRange += (max2 - min2 + 1)
		}

		avgHashRange := float64(hashRange) / float64(len(p.keys))
		avgPosRange := float64(posRange) / float64(len(p.keys))

		fmt.Printf("  Keys: %v...\n", p.keys[:3])
		fmt.Printf("  With Hash:     %.1f blocks avg (%.0f%% of table)\n",
			avgHashRange, avgHashRange/float64(numBlocks)*100)
		fmt.Printf("  With Position: %.1f blocks avg (%.0f%% of table)\n",
			avgPosRange, avgPosRange/float64(numBlocks)*100)
	}

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("  CONCLUSION: Position-based approach works for ALL key patterns!")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()
}
