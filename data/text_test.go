package data

import (
	"testing"

	"labsim/internal/content"
)

func TestTextCatalogLoads(t *testing.T) {
	c, err := content.Load(Text)
	if err != nil {
		t.Fatalf("text.json invalid: %v", err)
	}
	if len(c.Keys()) == 0 {
		t.Fatal("text.json is empty")
	}
}
