package illuminated

import (
	"testing"
)

var testOverrides = []override{
	{
		Title:       "Lantern",
		Language:    "zh",
		Original:    "灯笼",
		Replacement: "蓝灯",
	},
	{
		Title:       "Block",
		Language:    "en",
		Original:    "blacklist",
		Replacement: "block list",
	},
	{
		Title:       "Allow",
		Language:    "en",
		Original:    "whitelist",
		Replacement: "allow list",
	},
}

func TestOverrides(t *testing.T) {
	path := "test_overrides.yaml"
	err := WriteOverrideFile(path, testOverrides)
	if err != nil {
		t.Fatalf("failed to write overrides: %v", err)
	}
	// defer os.Remove(path)

	overrides, err := ReadOverrideFile(path)
	if err != nil {
		t.Fatalf("failed to read overrides: %v", err)
	}
	for i := range overrides {
		if overrides[i].Title != testOverrides[i].Title ||
			overrides[i].Language != testOverrides[i].Language ||
			overrides[i].Original != testOverrides[i].Original ||
			overrides[i].Replacement != testOverrides[i].Replacement {
			t.Errorf("override mismatch at index %d: got %+v, want %+v", i, overrides[i], testOverrides[i])
		}
	}
}
