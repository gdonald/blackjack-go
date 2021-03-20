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
	SavedGameFile   = "bj.json"
	MinBet          = 5.0
	MaxBet          = 1000.0
	CardsPerDeck    = 52
	MaxPlayerHands  = 7

	// h.count
	HardCount = false
	SoftCount = true

	// h.Status
	Unknown = -2
	Lost    = -1
	Push    = 0
	Won     = 1
)

var Suits = [4]string{"Spades", "Hearts", "Diamonds", "Clubs"}
var ShuffleSpecs = []int{80, 81, 82, 84, 86, 89, 92, 95}
var CardFaces = [14][4]string{
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
	sg := &SavedGame{NumberOfDecks: BJ.NumberOfDecks, CurrentBet: BJ.CurrentBet, Money: BJ.Money}
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

func loadGame() (BJ *Game, err error) {
	bytes, err := ioutil.ReadFile(SavedGameFile)
	if err == nil {
		sg := &SavedGame{}
		err := json.Unmarshal(bytes, sg)
		if err == nil {
			BJ = &Game{NumberOfDecks: sg.NumberOfDecks, Money: sg.Money, CurrentBet: sg.CurrentBet, CurrentPlayerHand: -1}
			return BJ, nil
		}
	}
	return nil, err
}

type SavedGame struct {
	NumberOfDecks int
	CurrentBet    float64
	Money         float64
}

type Game struct {
	Shoe
	DealerHand        Hand
	PlayerHands       []Hand
	NumberOfDecks     int
	CurrentPlayerHand int
	CurrentBet        float64
	Money             float64
}

func New() {
	BJ, err := loadGame()
	fmt.Println("BJ: ", BJ)
	fmt.Println("err: ", err)

	if BJ == nil {
		BJ = &Game{NumberOfDecks: 8, Money: 100, CurrentBet: 5.0, CurrentPlayerHand: -1}
	}

	s := NewShoe(BJ)
	BJ.Shoe = s
	BJ.DealNewHand()
}

func (BJ *Game) String() string {
	return fmt.Sprintf("NumberOfDecks: %d Money: %.2f CurrentBet: %.2f", BJ.NumberOfDecks, BJ.Money, BJ.CurrentBet)
}

func (BJ *Game) moreHandsToPlay() bool {
	return BJ.CurrentPlayerHand < (len(BJ.PlayerHands) - 1)
}

func (BJ *Game) playMoreHands() {
	BJ.CurrentPlayerHand++
	h := &BJ.PlayerHands[BJ.CurrentPlayerHand]
	h.dealCard()
	if h.isDone() {
		h.process()
		return
	}
	BJ.drawHands()
	h.drawPlayerHandOptions()
}

func (BJ *Game) needToPlayDealerHand() bool {
	for _, h := range BJ.PlayerHands {
		if !(h.isBusted() || h.isBlackjack()) {
			return true
		}
	}
	return false
}

func (BJ *Game) playDealerHand() {

	if BJ.DealerHand.isBlackjack() {
		BJ.DealerHand.HideDownCard = false
	}
	if !BJ.needToPlayDealerHand() {
		BJ.DealerHand.Played = true
		BJ.payHands()
		return
	}

	// unhide so the count is correct
	BJ.DealerHand.HideDownCard = false

	softCount := BJ.DealerHand.getValue(SoftCount)
	hardCount := BJ.DealerHand.getValue(HardCount)
	for softCount < 18 && hardCount < 17 {
		BJ.DealerHand.dealCard()
		softCount = BJ.DealerHand.getValue(SoftCount)
		hardCount = BJ.DealerHand.getValue(HardCount)
	}
	BJ.DealerHand.Played = true
	BJ.payHands()
}

func (BJ *Game) payHands() {
	BJ.CurrentPlayerHand = -1
	dhv := BJ.DealerHand.getValue(SoftCount)
	dhb := BJ.DealerHand.isBusted()
	for hand := 0; hand < len(BJ.PlayerHands); hand++ {
		h := &BJ.PlayerHands[hand]
		if h.Payed {
			continue
		}
		h.Payed = true
		phv := h.getValue(SoftCount)
		if dhb || phv > dhv {
			if h.isBlackjack() {
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

func (BJ *Game) drawHands() {
	c := exec.Command("clear")
	c.Stdout = os.Stdout
	c.Run()

	// dealer
	fmt.Println("\n Dealer:")
	fmt.Printf(" ")
	for card := 0; card < len(BJ.DealerHand.Cards); card++ {
		if card == 1 && BJ.DealerHand.HideDownCard {
			fmt.Printf("%s ", CardFaces[13][0])
		} else {
			c := BJ.DealerHand.Cards[card]
			fmt.Printf("%s ", CardFaces[c.Value-1][c.SuitValue])
		}
	}
	fmt.Printf(" ⇒ %2d", BJ.DealerHand.getValue(SoftCount))

	if !BJ.DealerHand.HideDownCard {
		fmt.Printf("  ")
		if BJ.DealerHand.isBusted() {
			fmt.Printf("Busted!")
		} else if BJ.DealerHand.isBlackjack() {
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
			fmt.Printf("%s ", CardFaces[c.Value-1][c.SuitValue])
		}
		fmt.Printf(" ⇒ %2d  ", h.getValue(SoftCount))
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
			if h.isBusted() {
				fmt.Printf("Busted!")
			} else {
				fmt.Printf("Lose!")
			}
		} else if h.Status == Won {
			if h.isBlackjack() {
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
	hx.Payed = true
	hx.Status = Lost
	BJ.Money -= hx.Bet
	BJ.drawHands()
	BJ.drawPlayerBetOptions()
}

func (BJ *Game) noInsurance() {
	if BJ.DealerHand.isBlackjack() {
		BJ.DealerHand.HideDownCard = false
		BJ.DealerHand.Played = true
		BJ.payHands()
		BJ.drawHands()
		BJ.drawPlayerBetOptions()
		return
	}

	h := &BJ.PlayerHands[BJ.CurrentPlayerHand]
	if h.isDone() {
		BJ.playDealerHand()
		BJ.drawHands()
		BJ.drawPlayerBetOptions()
		return
	}

	BJ.drawHands()
	h.drawPlayerHandOptions()
}

func (BJ *Game) askPlayerInsurance() {
	s := " Insurance, Y/N ?"
	fmt.Println(s)

	br := false
	for {
		b := getch()
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

func (BJ *Game) getNewBet() {
	BJ.drawHands()

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

func (BJ *Game) drawPlayerBetOptions() {
	s := " (D) Deal Hand  (B) Change Bet  (Q) Quit"
	fmt.Println(s)

	br := false
	for {
		b := getch()
		switch strings.ToLower(string(b)) {
		case "d":
			br = true
			BJ.DealNewHand()
		case "b":
			br = true
			BJ.getNewBet()
		case "q":
			br = true
			os.Exit(0)
		}
		if br {
			break
		}
	}
}

func (BJ *Game) DealNewHand() {
	if BJ.checkNeedToShuffle() {
		BJ.shuffle()
	}

	BJ.CurrentPlayerHand = 0
	BJ.DealerHand = Hand{Game: BJ, IsDealer: true, HideDownCard: true}
	BJ.PlayerHands = []Hand{}
	h := Hand{Game: BJ, Bet: BJ.CurrentBet, Status: Unknown}
	h.dealCard()
	BJ.DealerHand.dealCard()
	h.dealCard()
	BJ.DealerHand.dealCard()
	BJ.PlayerHands = append(BJ.PlayerHands, h)

	if BJ.DealerHand.Cards[0].isAce() {
		BJ.drawHands()
		BJ.askPlayerInsurance()
		return
	}

	if BJ.DealerHand.isDone() {
		BJ.DealerHand.HideDownCard = false
	}
	if BJ.PlayerHands[0].isDone() {
		BJ.DealerHand.HideDownCard = false
	}
	if BJ.DealerHand.Played || BJ.PlayerHands[0].Played {
		BJ.payHands()
		BJ.drawHands()
		BJ.drawPlayerBetOptions()
		return
	}

	BJ.drawHands()
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
