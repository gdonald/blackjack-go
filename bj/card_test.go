package bj

import (
	"testing"
)

func TestIsAce(t *testing.T) {
	card := Card{Value: 0}
	got := card.IsAce()
	if !got {
		t.Errorf("IsAce = %t; wanted true", got)
	}

	card.Value = 1
	got = card.IsAce()
	if got {
		t.Errorf("IsAce = %t; wanted false", got)
	}
}

func TestIsTen(t *testing.T) {
	card := Card{Value: 8}
	got := card.IsTen()
	if got {
		t.Errorf("IsTen = %t; wanted false", got)
	}

	card.Value = 9
	got = card.IsTen()
	if !got {
		t.Errorf("IsTen = %t; wanted true", got)
	}
}
