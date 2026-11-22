package filter

import (
	"errors"
	"testing"
)

func TestBlocksBadWords(t *testing.T) {
	for _, s := range []string{
		"shit", "Shit Head", "fuck", "FUCK", "bitch", "asshole", "Dr Ass", "ass",
		"dick", "damn", "crap", "piss", "bastard", "whore", "slut", "cunt",
		"stupid", "Idiot Face", "dumb", "loser", "moron", "killer", "kill", "hate",
		"sexy", "sex", "porn", "nazi", "hitler", "boobs", "boob", "butt", "rapist",
	} {
		if err := Check(s); !errors.Is(err, ErrBlocked) {
			t.Errorf("Check(%q) = %v, want ErrBlocked", s, err)
		}
	}
}

func TestBlocksTricks(t *testing.T) {
	for name, s := range map[string]string{
		"leet":            "sh1t",
		"leet2":           "a55",
		"leet3":           "b!tch",
		"symbols":         "$hit",
		"spaced":          "s h i t",
		"dotted":          "f.u.c.k",
		"dashed":          "d-i-c-k",
		"stretched":       "fuuuuck",
		"stretched2":      "shiiiit",
		"stretched3":      "asssss",
		"inside word":     "superfuckmaster",
		"inside word2":    "BigShitPants",
		"plural":          "dicks",
		"mixed case":      "sHiT",
		"spaced in name":  "Doctor k i l l",
		"two letter runs": "sh it",
	} {
		if err := Check(s); !errors.Is(err, ErrBlocked) {
			t.Errorf("%s: Check(%q) = %v, want ErrBlocked", name, s, err)
		}
	}
}

// The allow-list matters: kids have names and use words that contain bad substrings.
func TestAllowsInnocentWords(t *testing.T) {
	for _, s := range []string{
		"Doctor Sparkle Beaker", "Class Act", "Glass Beaker", "Grass", "Bass", "Peacock",
		"Cockatoo", "Hancock", "Scunthorpe", "Assistant", "Massive", "Passion", "Dickens",
		"Titan", "Title", "Skill", "Skillful", "Grape", "Drape", "Scrap", "Shell", "Hello",
		"Bob", "Boo Boo", "Pass", "Analysis", "Canal", "Essex", "Hatem", "Button",
		"Butterfly", "Mia", "Zoe", "Ava", "Noah", "Professor Pip", "Dr. O'Neil", "Ana-Lee",
		"Sussex", "Cassie", "Assemble", "Therapist", "Swank",
	} {
		if err := Check(s); err != nil {
			t.Errorf("Check(%q) = %v, want ok", s, err)
		}
	}
}

func TestEmptyTooLongBadChars(t *testing.T) {
	if err := Check("   "); !errors.Is(err, ErrEmpty) {
		t.Errorf("blank: %v", err)
	}
	if err := Check("Professor Sparkleberry Wigglesworth"); !errors.Is(err, ErrTooLong) {
		t.Errorf("long: %v", err)
	}
	for _, s := range []string{"Bob<script>", "Hi 😀", "a_b", "x/y"} {
		if err := Check(s); !errors.Is(err, ErrBadChars) {
			t.Errorf("Check(%q) = %v, want ErrBadChars", s, err)
		}
	}
}

func TestClean(t *testing.T) {
	if got := Clean("  Doctor   Pip  "); got != "Doctor Pip" {
		t.Fatalf("Clean = %q", got)
	}
}
