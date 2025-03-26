package sources 

import (
	"strings"
	"testing"
)


func TestBing_GetImageLinks(t *testing.T) {
	b := Bing{}
	items, err := b.GetImageLinks()
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	if len(items) == 0 {
		t.Errorf("No items found")
	}
	t.Logf("found items: \n%v", strings.Join(items, "\n"))
}
