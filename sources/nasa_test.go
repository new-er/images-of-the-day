package sources

import (
	"strings"
	"testing"
)

func TestNasa_GetImageLinks(t *testing.T) {
	n := Nasa{}
	items, err := n.GetImageLinks()
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if len(items) == 0 {
		t.Errorf("No items found")
	}
	t.Logf("found items:\n %v", strings.Join(items, "\n"))
}
