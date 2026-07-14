package sensor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPowerOptimisticReadTrimsWhitespace(t *testing.T) {
	path := filepath.Join(t.TempDir(), "capacity")
	require.NoError(t, os.WriteFile(path, []byte(" \t64\r\n"), 0o600))

	value := (Power{}).optimisticRead(path)

	require.Equal(t, "64", value)
}
