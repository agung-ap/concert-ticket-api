package load

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Get the code
	code := m.Run()

	// Exit
	os.Exit(code)
}
