package bj

type Card struct {
	Value     int
	Suit      string
	SuitValue int
}

func (c *Card) isAce() bool {
	return c.Value == 1
}

func (c *Card) isTen() bool {
	return c.Value > 8
}
