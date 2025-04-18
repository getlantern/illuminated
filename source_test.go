//go:build integration

package illuminated

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	testWikiURL = "https://github.com/getlantern/guide.wiki.git"
	testWikiDir = "git-temp"
)

func TestCloneRepo(t *testing.T) {
	err := cloneRepo(testWikiURL, testWikiDir)
	require.NoError(t, err)
	t.Logf("cloned repo to %q", testWikiDir)

	err = os.RemoveAll(testWikiDir)
	require.NoError(t, err)
	t.Logf("removed cloned repo %q", testWikiDir)
}
