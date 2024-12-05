package bj

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	SavedGameFile  = "bj.json"
	MinBet         = 5.0
	MaxBet         = 1000.0
	CardsPerDeck   = 52
	MaxPlayerHands = 7
	MinNumDecks    = 1
	MaxNumDecks    = 8

	HardCount = false
	SoftCount = true

	Unknown = -2
	Lost    = -1
	Push    = 0
	Won     = 1
)

var ShuffleSpecs = []int{80, 81, 82, 84, 86, 89, 92, 95}

var CardFaces = [14][4]string{
	{"2♠", "2♥", "2♣", "2♦"},
	{"3♠", "3♥", "3♣", "3♦"},
	{"4♠", "4♥", "4♣", "4♦"},
	{"5♠", "5♥", "5♣", "5♦"},
	{"6♠", "6♥", "6♣", "6♦"},
	{"7♠", "7♥", "7♣", "7♦"},
	{"8♠", "8♥", "8♣", "8♦"},
	{"9♠", "9♥", "9♣", "9♦"},
	{"T♠", "T♥", "T♣", "T♦"},
	{"J♠", "J♥", "J♣", "J♦"},
	{"Q♠", "Q♥", "Q♣", "Q♦"},
	{"K♠", "K♥", "K♣", "K♦"},
	{"??", "", "", ""},
}

var CardFaces2 = [14][4]string{
	{"🂡", "🂱", "🃁", "🃑"},
	{"🂢", "🂲", "🃂", "🃒"},
	{"🂣", "🂳", "🃃", "🃓"},
	{"🂤", "🂴", "🃄", "🃔"},
	{"🂥", "🂵", "🃅", "🃕"},
	{"🂦", "🂶", "🃆", "🃖"},
	{"🂧", "🂷", "🃇", "🃗"},
	{"🂨", "🂸", "🃈", "🃘"},
	{"🂩", "🂹", "🃉", "🃙"},
	{"🂪", "🂺", "🃊", "🃚"},
	{"🂫", "🂻", "🃋", "🃛"},
	{"🂭", "🂽", "🃍", "🃝"},
	{"🂮", "🂾", "🃎", "🃞"},
	{"🂠", "", "", ""},
}

func (BJ *Game) saveGame() {
	f, err := os.OpenFile(SavedGameFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Printf("error opening saved game file: %v", err)
		os.Exit(1)
	}
	defer f.Close()
	sg := &SavedGame{NumberOfDecks: BJ.NumberOfDecks, FaceType: BJ.FaceType, CurrentBet: BJ.CurrentBet, Money: BJ.Money}
	g, err := json.Marshal(sg)
	if err != nil {
		fmt.Printf("error saving game data: %v", err)
		os.Exit(1)
	}
	_, err = f.Write(g)
	if err != nil {
		fmt.Printf("error writing saved game data: %v", err)
		os.Exit(1)
	}
}

func LoadGame() (BJ *Game, err error) {
	bytes, err := ioutil.ReadFile(SavedGameFile)
	if err == nil {
		sg := &SavedGame{}
		err := json.Unmarshal(bytes, sg)
		if err == nil {
			BJ = &Game{NumberOfDecks: sg.NumberOfDecks, FaceType: sg.FaceType, Money: sg.Money, CurrentBet: sg.CurrentBet, CurrentPlayerHand: -1}
			return BJ, nil
		}
	}
	return BJ, err
}

type SavedGame struct {
	NumberOfDecks int
	FaceType      int
	CurrentBet    float64
	Money         float64
}

type Game struct {
	Shoe
	DealerHand        Hand
	PlayerHands       []Hand
	NumberOfDecks     int
	FaceType          int
	CurrentPlayerHand int
	CurrentBet        float64
	Money             float64
}

func New() {
	BJ, _ := LoadGame()
	// fmt.Println("BJ: ", BJ)
	// fmt.Println("err: ", err)

	if BJ == nil {
		BJ = &Game{NumberOfDecks: 8, FaceType: 2, Money: 100, CurrentBet: 5.0, CurrentPlayerHand: -1}
	}

	s := NewShoe(BJ)
	BJ.Shoe = s
	BJ.DealNewHand()
}

func (BJ *Game) String() string {
	return fmt.Sprintf("NumberOfDecks: %d FaceType: %d Money: %.2f CurrentBet: %.2f", BJ.NumberOfDecks, BJ.FaceType, BJ.Money, BJ.CurrentBet)
}

func (BJ *Game) MoreHandsToPlay() bool {
	return BJ.CurrentPlayerHand < (len(BJ.PlayerHands) - 1)
}

func (BJ *Game) PlayMoreHands() {
	BJ.CurrentPlayerHand++
	h := &BJ.PlayerHands[BJ.CurrentPlayerHand]
	h.DealCard()
	if h.IsDone() {
		h.Process()
		return
	}
	BJ.DrawHands()
	h.drawPlayerHandOptions()
}

func (BJ *Game) NeedToPlayDealerHand() bool {
	for _, h := range BJ.PlayerHands {
		if !(h.IsBusted() || h.IsBlackjack()) {
			return true
		}
	}
	return false
}

func (BJ *Game) PlayDealerHand() {

	if BJ.DealerHand.IsBlackjack() {
		BJ.DealerHand.HideDownCard = false
	}
	if !BJ.NeedToPlayDealerHand() {
		BJ.DealerHand.Played = true
		BJ.PayHands()
		return
	}

	// unhide so the count is correct
	BJ.DealerHand.HideDownCard = false

	softCount := BJ.DealerHand.GetValue(SoftCount)
	hardCount := BJ.DealerHand.GetValue(HardCount)
	for softCount < 18 && hardCount < 17 {
		BJ.DealerHand.DealCard()
		softCount = BJ.DealerHand.GetValue(SoftCount)
		hardCount = BJ.DealerHand.GetValue(HardCount)
	}
	BJ.DealerHand.Played = true
	BJ.PayHands()
}

func (BJ *Game) PayHands() {
	BJ.CurrentPlayerHand = -1
	dhv := BJ.DealerHand.GetValue(SoftCount)
	dhb := BJ.DealerHand.IsBusted()
	for hand := 0; hand < len(BJ.PlayerHands); hand++ {
		h := &BJ.PlayerHands[hand]
		if h.Paid {
			continue
		}
		h.Paid = true
		phv := h.GetValue(SoftCount)
		if dhb || phv > dhv {
			if h.IsBlackjack() {
				h.Bet *= 1.5
				BJ.Money += h.Bet
			} else {
				BJ.Money += h.Bet
			}
			h.Status = Won
		} else if phv < dhv {
			BJ.Money -= h.Bet
			h.Status = Lost
		} else {
			h.Status = Push
		}
	}
	BJ.saveGame()
}

func (BJ *Game) Clear() {
	c := exec.Command("clear")
	c.Stdout = os.Stdout
	c.Run()
}

func (BJ *Game) GetCardFace(value int, suit int) string {
	if BJ.FaceType == 2 {
		return CardFaces2[value][suit]
	}

	return CardFaces[value][suit]
}

func (BJ *Game) DrawHands() {
	BJ.Clear()

	// dealer
	fmt.Println("\n Dealer:")
	fmt.Printf(" ")
	for card := 0; card < len(BJ.DealerHand.Cards); card++ {
		if card == 1 && BJ.DealerHand.HideDownCard {
			fmt.Printf("%s ", BJ.GetCardFace(13, 0))
		} else {
			c := BJ.DealerHand.Cards[card]
			fmt.Printf("%s ", BJ.GetCardFace(c.Value-1, c.SuitValue))
		}
	}
	fmt.Printf(" ⇒ %2d", BJ.DealerHand.GetValue(SoftCount))

	if !BJ.DealerHand.HideDownCard {
		fmt.Printf("  ")
		if BJ.DealerHand.IsBusted() {
			fmt.Printf("Busted!")
		} else if BJ.DealerHand.IsBlackjack() {
			fmt.Printf("Blackjack!")
		}
	}
	fmt.Println()

	// player
	fmt.Printf("\n Player $%.2f:\n", BJ.Money)
	for hand := 0; hand < len(BJ.PlayerHands); hand++ {
		h := BJ.PlayerHands[hand]
		fmt.Printf(" ")
		for card := 0; card < len(h.Cards); card++ {
			c := h.Cards[card]
			fmt.Printf("%s ", BJ.GetCardFace(c.Value-1, c.SuitValue))
		}
		fmt.Printf(" ⇒ %2d  ", h.GetValue(SoftCount))
		if h.Status == Lost {
			fmt.Printf("-")
		} else if h.Status == Won {
			fmt.Printf("+")
		}
		fmt.Printf("$%.2f", h.Bet)
		if !h.Played && hand == BJ.CurrentPlayerHand {
			fmt.Printf(" ⇐")
		}
		fmt.Printf("  ")
		if h.Status == Lost {
			if h.IsBusted() {
				fmt.Printf("Busted!")
			} else {
				fmt.Printf("Lose!")
			}
		} else if h.Status == Won {
			if h.IsBlackjack() {
				fmt.Printf("Blackjack!")
			} else {
				fmt.Printf("Win!")
			}
		} else if h.Status == Push {
			fmt.Printf("Push")
		}
		fmt.Println()
		fmt.Println()
	}
}

func (BJ *Game) insureHand() {
	hx := &BJ.PlayerHands[BJ.CurrentPlayerHand]
	hx.Bet /= 2
	hx.Played = true
	hx.Paid = true
	hx.Status = Lost
	BJ.Money -= hx.Bet
	BJ.DrawHands()
	BJ.DrawPlayerBetOptions()
}

func (BJ *Game) noInsurance() {
	if BJ.DealerHand.IsBlackjack() {
		BJ.DealerHand.HideDownCard = false
		BJ.DealerHand.Played = true
		BJ.PayHands()
		BJ.DrawHands()
		BJ.DrawPlayerBetOptions()
		return
	}

	h := &BJ.PlayerHands[BJ.CurrentPlayerHand]
	if h.IsDone() {
		BJ.PlayDealerHand()
		BJ.DrawHands()
		BJ.DrawPlayerBetOptions()
		return
	}

	BJ.DrawHands()
	h.drawPlayerHandOptions()
}

func (BJ *Game) AskPlayerInsurance() {
	s := " Insurance, Y/N ?"
	fmt.Println(s)

	br := false
	for {
		b := GetChar()
		switch strings.ToLower(string(b)) {
		case "y":
			br = true
			BJ.insureHand()
		case "n":
			br = true
			BJ.noInsurance()
		}
		if br {
			break
		}
	}
}

func (BJ *Game) GetNewBet() {
	BJ.DrawHands()

	fmt.Printf("  Current Bet: $%.2f\n  Enter new Bet: ", BJ.CurrentBet)
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	Bet, _ := strconv.ParseFloat(strings.TrimSpace(text), 64)

	if Bet < MinBet {
		Bet = MinBet
	} else if Bet > MaxBet {
		Bet = MaxBet
	}
	BJ.CurrentBet = Bet
	BJ.DealNewHand()
}

func (BJ *Game) DrawPlayerBetOptions() {
	s := " (D) Deal Hand  (B) Change Bet  (O) Options  (Q) Quit"
	fmt.Println(s)

	br := false
	for {
		b := GetChar()
		switch strings.ToLower(string(b)) {
		case "d":
			br = true
			BJ.DealNewHand()
		case "b":
			br = true
			BJ.GetNewBet()
		case "o":
			br = true
			BJ.GameOptions()
		case "q":
			br = true
			os.Exit(0)
		}
		if br {
			break
		}
	}
}

func (BJ *Game) GameOptions() {
	BJ.Clear()
	BJ.DrawHands()

	s := " (N) Number of Decks  (T) Deck Type  (F) Face Type  (B) Back"
	fmt.Println(s)

	br := false
	for {
		b := GetChar()
		switch strings.ToLower(string(b)) {
		case "n":
			br = true
			BJ.GetNumDecks()
		case "t":
			br = true
			BJ.GetNewDeckType()
		case "f":
			br = true
			BJ.GetNewFaceType()
		case "b":
			br = true
			BJ.Clear()
			BJ.DrawHands()
			BJ.DrawPlayerBetOptions()
		default:
			BJ.Clear()
			BJ.DrawHands()
			BJ.GameOptions()
		}

		if br {
			break
		}
	}
}

func (BJ *Game) GetNumDecks() {
	BJ.Clear()
	BJ.DrawHands()

	fmt.Printf("  Number Of Decks: %d  Enter New Number Of Decks (1-8): ", BJ.NumberOfDecks)
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	NumOfDecks, _ := strconv.ParseInt(strings.TrimSpace(text), 10, 32)

	if NumOfDecks < MinNumDecks {
		NumOfDecks = MinNumDecks
	} else if NumOfDecks > MaxNumDecks {
		NumOfDecks = MaxNumDecks
	}
	BJ.NumberOfDecks = int(NumOfDecks)
	BJ.GameOptions()
}

func (BJ *Game) GetNewFaceType() {
	BJ.Clear()
	BJ.DrawHands()

	s := " (1) A♠  (2) 🂡"
	fmt.Println(s)

	br := false
	for {
		b := GetChar()
		switch strings.ToLower(string(b)) {
		case "1":
			br = true
			BJ.FaceType = 1
		case "2":
			br = true
			BJ.FaceType = 2
		default:
			BJ.Clear()
			BJ.DrawHands()
			BJ.GetNewFaceType()
		}

		if br {
			break
		}
	}

	BJ.saveGame()
	BJ.DrawHands()
	BJ.DrawPlayerBetOptions()
}

func (BJ *Game) GetNewDeckType() {
	BJ.Clear()
	BJ.DrawHands()

	s := " (1) Regular  (2) Aces  (3) Jacks  (4) Aces & Jacks  (5) Sevens  (6) Eights"
	fmt.Println(s)

	br := false
	for {
		b := GetChar()
		switch strings.ToLower(string(b)) {
		case "1":
			br = true
			BJ.NewRegular()
		case "2":
			br = true
			BJ.NewAces()
		case "3":
			br = true
			BJ.NewJacks()
		case "4":
			br = true
			BJ.NewAcesJacks()
		case "5":
			br = true
			BJ.NewSevens()
		case "6":
			br = true
			BJ.NewEights()
		default:
			BJ.Clear()
			BJ.DrawHands()
			BJ.GetNewDeckType()
		}

		if br {
			break
		}
	}

	BJ.DrawHands()
	BJ.DrawPlayerBetOptions()
}

func (BJ *Game) DealNewHand() {
	if BJ.CheckNeedToShuffle() {
		BJ.Shuffle()
	}

	BJ.CurrentPlayerHand = 0
	BJ.DealerHand = Hand{Game: BJ, IsDealer: true, HideDownCard: true}
	BJ.PlayerHands = []Hand{}
	h := Hand{Game: BJ, Bet: BJ.CurrentBet, Status: Unknown}
	h.DealCard()
	BJ.DealerHand.DealCard()
	h.DealCard()
	BJ.DealerHand.DealCard()
	BJ.PlayerHands = append(BJ.PlayerHands, h)

	if BJ.DealerHand.Cards[0].IsAce() {
		BJ.DrawHands()
		BJ.AskPlayerInsurance()
		return
	}

	if BJ.DealerHand.IsDone() {
		BJ.DealerHand.HideDownCard = false
	}
	if BJ.PlayerHands[0].IsDone() {
		BJ.DealerHand.HideDownCard = false
	}
	if BJ.DealerHand.Played || BJ.PlayerHands[0].Played {
		BJ.PayHands()
		BJ.DrawHands()
		BJ.DrawPlayerBetOptions()
		return
	}

	BJ.DrawHands()
	BJ.PlayerHands[0].drawPlayerHandOptions()

	BJ.saveGame()
}

func (BJ *Game) l(s ...interface{}) {
	f, err := os.OpenFile("bj.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("error opening log file: %v", err)
		os.Exit(1)
	}
	defer f.Close()
	log.SetOutput(f)
	ss := ""
	for _, p := range s {
		switch p.(type) {
		case bool:
			ss += fmt.Sprintf("%t ", p.(bool))
		case int:
			ss += fmt.Sprintf("%d ", p.(int))
		case float64:
			ss += fmt.Sprintf("%.2f ", p.(float64))
		case string:
			ss += fmt.Sprintf("%s ", p.(string))
		}
	}
	log.Println(ss)
}
