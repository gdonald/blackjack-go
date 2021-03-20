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
	Payed        bool
	Status       int
	Bet          float64
	Cards        []Card
}

func (h *Hand) isBusted() bool {
	return h.getValue(HardCount) > 21
}

func (h *Hand) isBlackjack() bool {
	if len(h.Cards) != 2 {
		return false
	}
	if h.Cards[0].isAce() && h.Cards[1].isTen() {
		return true
	}
	if h.Cards[1].isAce() && h.Cards[0].isTen() {
		return true
	}
	return false
}

func (h *Hand) getValue(softCount bool) int {

	v := 0
	total := 0

	for x := 0; x < len(h.Cards); x++ {

		if x == 1 && h.HideDownCard {
			continue
		}

		// face Cards are 10
		if h.Cards[x].Value > 9 {
			v = 10
		} else {
			v = h.Cards[x].Value
		}

		// raise ace to 11 if possible
		if softCount && v == 1 && total < 11 {
			v = 11
		}

		total += v
	}

	// redo as hard count if soft counting and busted
	if softCount && total > 21 {
		return h.getValue(HardCount)
	}

	return total
}

func (h *Hand) isDone() bool {
	if h.Stood || h.Played || h.isBlackjack() || h.isBusted() || 21 == h.getValue(SoftCount) || 21 == h.getValue(HardCount) {
		h.Played = true
		if !h.Payed {
			if h.isBusted() {
				h.Payed = true
				h.Status = Lost
				h.Money -= h.Bet
			}
		}
		return true
	}
	return false
}

func (h *Hand) hit() {
	h.dealCard()
	if h.isDone() {
		h.process()
	} else {
		h.drawHands()
		hx := &h.PlayerHands[h.CurrentPlayerHand]
		hx.drawPlayerHandOptions()
		return
	}
}

func (h *Hand) stand() {
	h.Stood = true
	h.Played = true
	if h.moreHandsToPlay() {
		h.playMoreHands()
		return
	}
	h.playDealerHand()
	h.drawHands()
	h.drawPlayerBetOptions()
}

func (h *Hand) double() {
	h.dealCard()
	h.Played = true
	h.Bet *= 2
	if h.isDone() { // always true
		h.process()
	}
}

func (h *Hand) split() {
	if !h.canSplit() {
		h.drawHands()
		h.drawPlayerHandOptions()
		return
	}

	// add new hand
	h.PlayerHands = append(h.PlayerHands, Hand{Game: h.Game})

	// move hands down
	x := len(h.PlayerHands) - 1
	for x > h.CurrentPlayerHand {
		h.PlayerHands[x] = h.PlayerHands[x-1]
		x--
	}

	// split
	Cards := []Card{}
	Cards = append(Cards, h.Cards[0])
	h.PlayerHands[h.CurrentPlayerHand+1].Cards = Cards
	h.Cards[0] = h.Shoe.getNextCard()

	if h.isDone() {
		h.process()
	} else {
		h.drawHands()
		hx := &h.PlayerHands[h.CurrentPlayerHand]
		hx.drawPlayerHandOptions()
		return
	}
}

func (h *Hand) process() {
	if h.moreHandsToPlay() {
		h.playMoreHands()
		return
	}
	h.playDealerHand()
	h.drawHands()
	h.drawPlayerBetOptions()
}

func (h *Hand) canHit() bool {
	if h.Played || h.Stood || 21 == h.getValue(HardCount) || h.isBlackjack() || h.isBusted() {
		return false
	}
	return true
}

func (h *Hand) canStand() bool {
	if h.Stood || h.isBusted() || h.isBlackjack() {
		return false
	}
	return true
}

func (h *Hand) canSplit() bool {
	if h.Stood || len(h.PlayerHands) >= MaxPlayerHands {
		return false
	}
	if len(h.Cards) == 2 && h.Cards[0].Value == h.Cards[1].Value {
		return true
	}
	return false
}

func (h *Hand) canDouble() bool {
	if h.Stood || len(h.Cards) != 2 || h.isBusted() || h.isBlackjack() {
		return false
	}
	return true
}

func (h *Hand) dealCard() {
	h.Cards = append(h.Cards, h.getNextCard())
}

func (h *Hand) drawPlayerHandOptions() {
	s := " "
	if h.canHit() {
		s += "(H) Hit  "
	}
	if h.canStand() {
		s += "(S) Stand  "
	}
	if h.canSplit() {
		s += "(P) Split  "
	}
	if h.canDouble() {
		s += "(D) Double  "
	}
	fmt.Println(s)

	br := false
	for {
		b := getch()
		switch strings.ToLower(string(b)) {
		case "h":
			br = true
			h.hit()
		case "s":
			br = true
			h.stand()
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
