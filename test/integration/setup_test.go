package integration

import (
	"os"
	"testing"

	"concert-ticket-api/test/testutil"
)

func TestMain(m *testing.M) {
	// Setup
	db, err := testutil.SetupTestDB()
	if err != nil {
		println("Could not setup test DB:", err.Error())
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	if db != nil {
		db.Close()
	}
	testutil.TeardownTestDB()

	os.Exit(code)
}
