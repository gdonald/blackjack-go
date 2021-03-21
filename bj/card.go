package bj

type Card struct {
	Value     int
	Suit      string
	SuitValue int
}

func (c *Card) IsAce() bool {
	return c.Value == 1
}

func (c *Card) IsTen() bool {
	return c.Value > 8
}
