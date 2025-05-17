package illuminated

import (
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	testProjectDir   = "testProject"
	nonDefaultConfig = Config{
		Base:    "en",
		Targets: []string{"fa", "ru"},
	}
)

func TestConfigWrite(t *testing.T) {
	err := nonDefaultConfig.Write(testProjectDir, false)
	require.NoError(t, err)
	err = os.RemoveAll(testProjectDir)
	require.NoError(t, err)
}

func TestConfigRead(t *testing.T) {
	err := nonDefaultConfig.Write(testProjectDir, false)
	require.NoError(t, err)

	var c Config
	err = c.Read(path.Join(testProjectDir, DefaultConfigFilename))
	require.NoError(t, err)
	require.Equal(t, true, reflect.DeepEqual(c, nonDefaultConfig))
	t.Logf("config read as written: \n%v", c)

	err = os.RemoveAll(testProjectDir)
	require.NoError(t, err)
}
