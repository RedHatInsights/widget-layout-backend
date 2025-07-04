// Package test_util provides utilities for reliable, conflict-free testing.
//
// The unique ID generator system ensures that tests don't interfere with each other
// by providing collision-free IDs and managing reserved IDs for special test scenarios.
//
// Usage:
//   - Use GetUniqueID() for all database record IDs in tests
//   - Use reserved constants (NonExistentID, NoDBTestID) for special cases
//   - Call ResetIDGenerator() in TestMain before running tests
//   - Reserve special IDs in TestMain to prevent conflicts
//
// For comprehensive documentation, see docs/TESTING.md
package test_util

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Reserved IDs for specific test scenarios.
//
// These constants should be used instead of magic numbers in tests to ensure
// consistency and clarity. They are automatically reserved by TestMain to
// prevent conflicts with generated IDs.
const (
	// NonExistentID is used for testing scenarios with IDs that don't exist in the database.
	// Use this when testing 404 responses or "record not found" error cases.
	NonExistentID uint = 999999

	// NoDBTestID is used for testing scenarios where database operations are mocked or skipped.
	// Use this when testing request validation, JSON parsing, or other logic that doesn't
	// require actual database records.
	NoDBTestID uint = 123456
)

// UniqueIDGenerator generates unique IDs for tests and tracks used IDs to prevent conflicts.
//
// This generator is thread-safe and ensures that no two calls to GenerateUniqueID()
// will return the same value, even in concurrent test execution.
//
// The generator uses a range of 100000-999999 for generated IDs, avoiding conflicts
// with reserved constants and providing a large enough space for test execution.
type UniqueIDGenerator struct {
	usedIDs map[uint]bool // Track which IDs have been used
	mutex   sync.Mutex    // Ensure thread safety
	rng     *rand.Rand    // Random number generator with unique seed
}

var (
	globalIDGenerator *UniqueIDGenerator
	once              sync.Once
)

// GetUniqueIDGenerator returns a singleton instance of UniqueIDGenerator
func GetUniqueIDGenerator() *UniqueIDGenerator {
	once.Do(func() {
		globalIDGenerator = &UniqueIDGenerator{
			usedIDs: make(map[uint]bool),
			mutex:   sync.Mutex{},
			rng:     rand.New(rand.NewSource(time.Now().UnixNano())),
		}
	})
	return globalIDGenerator
}

// GenerateUniqueID generates a unique ID that hasn't been used before
// It generates random IDs in the range 100000-999999 to avoid conflicts with hardcoded values
func (g *UniqueIDGenerator) GenerateUniqueID() uint {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	for {
		// Generate random ID between 100000 and 999999
		id := uint(g.rng.Intn(900000) + 100000)

		if !g.usedIDs[id] {
			g.usedIDs[id] = true
			return id
		}
	}
}

// ReserveID marks an ID as used (useful for hardcoded IDs in existing tests)
func (g *UniqueIDGenerator) ReserveID(id uint) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.usedIDs[id] = true
}

// Reset clears all used IDs (useful for test cleanup)
func (g *UniqueIDGenerator) Reset() {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.usedIDs = make(map[uint]bool)
}

// GetUniqueID is a convenience function to get a unique ID
func GetUniqueID() uint {
	return GetUniqueIDGenerator().GenerateUniqueID()
}

// ReserveID is a convenience function to reserve an ID
func ReserveID(id uint) {
	GetUniqueIDGenerator().ReserveID(id)
}

// ResetIDGenerator is a convenience function to reset the ID generator
func ResetIDGenerator() {
	GetUniqueIDGenerator().Reset()
}

// UniqueUserIDGenerator generates unique user IDs for tests to prevent conflicts between tests.
//
// This generator ensures that each test gets a unique user ID, preventing database conflicts
// and test interference when multiple tests create records for different users.
//
// The generator is thread-safe and provides collision-free user IDs.
type UniqueUserIDGenerator struct {
	usedUserIDs map[string]bool // Track which user IDs have been used
	mutex       sync.Mutex      // Ensure thread safety
	counter     int             // Simple counter for generating IDs
}

var (
	globalUserIDGenerator *UniqueUserIDGenerator
	userIDOnce            sync.Once
)

// GetUniqueUserIDGenerator returns a singleton instance of UniqueUserIDGenerator
func GetUniqueUserIDGenerator() *UniqueUserIDGenerator {
	userIDOnce.Do(func() {
		globalUserIDGenerator = &UniqueUserIDGenerator{
			usedUserIDs: make(map[string]bool),
			mutex:       sync.Mutex{},
			counter:     1,
		}
	})
	return globalUserIDGenerator
}

// GenerateUniqueUserID generates a unique user ID that hasn't been used before
// Returns a string in the format "test-user-{counter}" where counter is incremented for each call
func (g *UniqueUserIDGenerator) GenerateUniqueUserID() string {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	for {
		userID := fmt.Sprintf("test-user-%d", g.counter)
		g.counter++

		if !g.usedUserIDs[userID] {
			g.usedUserIDs[userID] = true
			return userID
		}
	}
}

// ReserveUserID marks a user ID as used (useful for hardcoded user IDs in existing tests)
func (g *UniqueUserIDGenerator) ReserveUserID(userID string) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.usedUserIDs[userID] = true
}

// Reset clears all used user IDs (useful for test cleanup)
func (g *UniqueUserIDGenerator) Reset() {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.usedUserIDs = make(map[string]bool)
	g.counter = 1
}

// GetUniqueUserID is a convenience function to get a unique user ID
func GetUniqueUserID() string {
	return GetUniqueUserIDGenerator().GenerateUniqueUserID()
}

// ReserveUserID is a convenience function to reserve a user ID
func ReserveUserID(userID string) {
	GetUniqueUserIDGenerator().ReserveUserID(userID)
}

// ResetUserIDGenerator is a convenience function to reset the user ID generator
func ResetUserIDGenerator() {
	GetUniqueUserIDGenerator().Reset()
}
