package illuminated

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCmdExists(t *testing.T) {
	t.Run("ls", func(t *testing.T) {
		_, err := exec.LookPath("ls")
		require.NoError(t, err)
	})
	t.Run("nonexistent", func(t *testing.T) {
		_, err := exec.LookPath("nonexistent")
		require.Error(t, err)
	})
}
