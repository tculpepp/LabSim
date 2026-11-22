package scenes

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"labsim/data"
	"labsim/internal/content"
	"labsim/internal/save"
)

// Every text key used in scene code must exist in data/text.json.
var keyRe = regexp.MustCompile(`\.(?:T|Tf|Confirm)\("([^"]+)"|\.Say\("[^"]+", "([^"]+)"|notice(?:,? [^=]*)?= (?:nil, )?"([^"]+)"`)

func TestAllTextKeysExist(t *testing.T) {
	cat, err := content.Load(data.Text)
	if err != nil {
		t.Fatal(err)
	}
	files, _ := filepath.Glob("*.go")
	found := 0
	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range keyRe.FindAllStringSubmatch(string(src), -1) {
			key := m[1] + m[2] + m[3]
			found++
			if !cat.Has(key) {
				t.Errorf("%s uses missing text key %q", f, key)
			}
		}
	}
	if found < 10 {
		t.Fatalf("only found %d keys; regexp broken?", found)
	}
}

func TestStationLabels(t *testing.T) {
	a := newApp(t, &save.MemStore{})
	for _, st := range a.Map.Stations() {
		if !a.Text.Has(stationKey(st.ID)) {
			t.Errorf("station %q has no %q text", st.ID, stationKey(st.ID))
		}
	}
}

func TestAnimalTexts(t *testing.T) {
	a := newApp(t, &save.MemStore{})
	for _, an := range a.Animals {
		if !a.Text.Has(animalKey(an.Kind)) {
			t.Errorf("animal %q has no %q text", an.Kind, animalKey(an.Kind))
		}
	}
}
