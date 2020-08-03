package main

import (
	"testing"
)

func TestToCFriend(t *testing.T) {
	ts := []struct {
		f Friend
	} {
		{ f: Friend{ ID: 1, Age: 10, }, },
		{ f: Friend{ ID: 2, Age: 20, }, },
	}

	for _, tc := range ts {
		f := tc.f

		cf := toCFriend(f)

		goID, cID := f.ID, cf.id
		if goID != int(cID) {
			t.Errorf("go id=%d c id=%d", goID, cID)
		}

		goAge, cAge := f.Age, cf.age
		if goAge != int(cAge) {
			t.Errorf("go age=%d c age=%d", goAge, cAge)
		}
	}
}
