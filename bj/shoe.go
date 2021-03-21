package bj

import (
	"math/rand"
	"time"
)

type Shoe struct {
	*Game
	Top   int
	Cards []Card
}

func NewShoe(BJ *Game) Shoe {
	rand.Seed(time.Now().UTC().UnixNano())
	s := Shoe{Game: BJ}
	s.NewRegular()
	s.Shuffle()
	return s
}

func (s *Shoe) GetNextCard() Card {
	c := s.Cards[s.Top]
	s.Top++
	return c
}

func (s *Shoe) NewRegular() {
	s.Cards = []Card{}
	for deck := 0; deck < s.NumberOfDecks; deck++ {
		for Suit := 0; Suit < 4; Suit++ {
			for Value := 1; Value < 14; Value++ {
				c := Card{Value: Value, Suit: Suits[Suit], SuitValue: Suit}
				s.Cards = append(s.Cards, c)
			}
		}
	}
}

func (s *Shoe) CheckNeedToShuffle() bool {
	used := int(float64(s.Top) / float64(len(s.Cards)) * 100.0)
	return used > ShuffleSpecs[s.NumberOfDecks - 1]
}

func (s *Shoe) Shuffle() {
	s.Top = 0
	totalCards := s.NumberOfDecks * CardsPerDeck
	for t := 0; t < 7; t++ {
		for x := 0; x < totalCards; x++ {
			y := rand.Intn(totalCards)
			s.Cards[x], s.Cards[y] = s.Cards[y], s.Cards[x]
		}
	}
}

func (s *Shoe) NewSevens() {
	s.Cards = []Card{}
	for deck := 0; deck < s.NumberOfDecks*5*13; deck++ {
		for Suit := 0; Suit < 4; Suit++ {
			c := Card{Value: 7, Suit: Suits[Suit], SuitValue: Suit}
			s.Cards = append(s.Cards, c)
		}
	}
}

func (s *Shoe) NewEights() {
	s.Cards = []Card{}
	for deck := 0; deck < s.NumberOfDecks*5*13; deck++ {
		for Suit := 0; Suit < 4; Suit++ {
			c := Card{Value: 8, Suit: Suits[Suit], SuitValue: Suit}
			s.Cards = append(s.Cards, c)
		}
	}
}

func (s *Shoe) NewAces() {
	s.Cards = []Card{}
	for deck := 0; deck < s.NumberOfDecks*5*13; deck++ {
		for Suit := 0; Suit < 4; Suit++ {
			c := Card{Value: 1, Suit: Suits[Suit], SuitValue: Suit}
			s.Cards = append(s.Cards, c)
		}
	}
}

func (s *Shoe) NewJacks() {
	s.Cards = []Card{}
	for deck := 0; deck < s.NumberOfDecks*5*13; deck++ {
		for Suit := 0; Suit < 4; Suit++ {
			c := Card{Value: 11, Suit: Suits[Suit], SuitValue: Suit}
			s.Cards = append(s.Cards, c)
		}
	}
}

func (s *Shoe) NewAcesJacks() {
	s.Cards = []Card{}
	for deck := 0; deck < s.NumberOfDecks*4*13; deck++ {
		for Suit := 0; Suit < 4; Suit++ {
			c := Card{Value: 1, Suit: Suits[Suit], SuitValue: Suit}
			s.Cards = append(s.Cards, c)
			c2 := Card{Value: 11, Suit: Suits[Suit], SuitValue: Suit}
			s.Cards = append(s.Cards, c2)
		}
	}
}
