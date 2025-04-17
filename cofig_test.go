package illuminated

import (
	"path"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	testProjectDir   = "testProject"
	nonDefaultConfig = config{
		Source:  "en",
		Targets: []string{"fa", "ru"},
	}
)

func TestConfigWrite(t *testing.T) {
	err := nonDefaultConfig.write(testProjectDir)
	require.NoError(t, err)
}

func TestConfigRead(t *testing.T) {
	err := nonDefaultConfig.write(testProjectDir)
	require.NoError(t, err)

	var c config
	err = c.read(path.Join(testProjectDir, DefaultConfigFilename))
	require.NoError(t, err)
	require.Equal(t, true, reflect.DeepEqual(c, nonDefaultConfig))
	t.Logf("config read as written: \n%v", c)
}
