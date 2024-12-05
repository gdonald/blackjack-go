package bj

import (
	"fmt"
	"strings"
)

type Hand struct {
	*Game
	IsDealer     bool
	HideDownCard bool
	Stood        bool
	Played       bool
	Paid         bool
	Status       int
	Bet          float64
	Cards        []Card
}

func (h *Hand) IsBusted() bool {
	return h.GetValue(HardCount) > 21
}

func (h *Hand) IsBlackjack() bool {
	if len(h.Cards) != 2 {
		return false
	}
	if h.Cards[0].IsAce() && h.Cards[1].IsTen() {
		return true
	}
	if h.Cards[1].IsAce() && h.Cards[0].IsTen() {
		return true
	}
	return false
}

func (h *Hand) GetValue(softCount bool) int {

	v := 0
	total := 0

	for x := 0; x < len(h.Cards); x++ {

		if x == 1 && h.HideDownCard {
			continue
		}

		if h.Cards[x].Value > 9 {
			v = 10
		} else {
			v = h.Cards[x].Value
		}

		if softCount && v == 1 && total < 11 {
			v = 11
		}

		total += v
	}

	if softCount && total > 21 {
		return h.GetValue(HardCount)
	}

	return total
}

func (h *Hand) IsDone() bool {
	if h.Stood || h.Played || h.IsBlackjack() || h.IsBusted() || 21 == h.GetValue(SoftCount) || 21 == h.GetValue(HardCount) {
		h.Played = true
		if !h.Paid {
			if h.IsBusted() {
				h.Paid = true
				h.Status = Lost
				h.Money -= h.Bet
			}
		}
		return true
	}
	return false
}

func (h *Hand) Hit() {
	h.DealCard()
	if h.IsDone() {
		h.Process()
	} else {
		h.DrawHands()
		hx := &h.PlayerHands[h.CurrentPlayerHand]
		hx.drawPlayerHandOptions()
		return
	}
}

func (h *Hand) Stand() {
	h.Stood = true
	h.Played = true
	if h.MoreHandsToPlay() {
		h.PlayMoreHands()
		return
	}
	h.PlayDealerHand()
	h.DrawHands()
	h.DrawPlayerBetOptions()
}

func (h *Hand) double() {
	h.DealCard()
	h.Played = true
	h.Bet *= 2
	if h.IsDone() { // always true
		h.Process()
	}
}

func (h *Hand) split() {
	if !h.CanSplit() {
		h.DrawHands()
		h.drawPlayerHandOptions()
		return
	}

	h.PlayerHands = append(h.PlayerHands, Hand{Game: h.Game})

	x := len(h.PlayerHands) - 1
	for x > h.CurrentPlayerHand {
		h.PlayerHands[x] = h.PlayerHands[x-1]
		x--
	}

	Cards := []Card{}
	Cards = append(Cards, h.Cards[0])
	h.PlayerHands[h.CurrentPlayerHand+1].Cards = Cards
	h.Cards[0] = h.Shoe.GetNextCard()

	if h.IsDone() {
		h.Process()
	} else {
		h.DrawHands()
		hx := &h.PlayerHands[h.CurrentPlayerHand]
		hx.drawPlayerHandOptions()
		return
	}
}

func (h *Hand) Process() {
	if h.MoreHandsToPlay() {
		h.PlayMoreHands()
		return
	}
	h.PlayDealerHand()
	h.DrawHands()
	h.DrawPlayerBetOptions()
}

func (h *Hand) CanHit() bool {
	if h.Played || h.Stood || 21 == h.GetValue(HardCount) || h.IsBlackjack() || h.IsBusted() {
		return false
	}
	return true
}

func (h *Hand) CanStand() bool {
	if h.Stood || h.IsBusted() || h.IsBlackjack() {
		return false
	}
	return true
}

func (h *Hand) CanSplit() bool {
	if h.Stood || len(h.PlayerHands) >= MaxPlayerHands {
		return false
	}
	if len(h.Cards) == 2 && h.Cards[0].Value == h.Cards[1].Value {
		return true
	}
	return false
}

func (h *Hand) CanDouble() bool {
	if h.Stood || len(h.Cards) != 2 || h.IsBusted() || h.IsBlackjack() {
		return false
	}
	return true
}

func (h *Hand) DealCard() {
	h.Cards = append(h.Cards, h.GetNextCard())
}

func (h *Hand) drawPlayerHandOptions() {
	s := " "
	if h.CanHit() {
		s += "(H) Hit  "
	}
	if h.CanStand() {
		s += "(S) Stand  "
	}
	if h.CanSplit() {
		s += "(P) Split  "
	}
	if h.CanDouble() {
		s += "(D) Double  "
	}
	fmt.Println(s)

	br := false
	for {
		b := GetChar()
		switch strings.ToLower(string(b)) {
		case "h":
			br = true
			h.Hit()
		case "s":
			br = true
			h.Stand()
		case "d":
			br = true
			h.double()
		case "p":
			br = true
			h.split()
		}
		if br {
			break
		}
	}
}
