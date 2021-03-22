package bj

type Card struct {
	Value     int
	SuitValue int
}

func (c *Card) IsAce() bool {
	return c.Value == 0
}

func (c *Card) IsTen() bool {
	return c.Value > 8
}
