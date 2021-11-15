package engine

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type Card struct {
	Number int // шестёрка-туз
	Class  int // 0,1,2,3 - club, diamonds, hearts, spades
}

type Deck []Card

type Player struct {
	Nickname   string // nickname
	Cards      []Card // hand
	BattleCard Card
	ID         int
}

type Players []Player

type Session struct {
	Deck     Deck    // deck contains not used cards, deck needs to give new cards to players
	Players  Players // list of players
	Turn     int     // number of turn
	Trump    Card    // copy of trump card
	Dumb     Player  // list of won players
	Attacker *Player
	Defender *Player
}

// stringer
type Stringer interface {
	String() string
}

func (c Card) String() string {
	class := []string{"♣", "♦", "♥", "♠"}
	number := []string{"6", "7", "8", "9", "10", "J", "Q", "K", "A"}
	return fmt.Sprint(number[c.Number] + class[c.Class])
}

func (p Player) String() (s string) {
	for i, e := range p.Cards {
		s += fmt.Sprint("[", i+1, "]", e.String(), " ")
	}
	return
}

func (p Players) String() (s string) {
	for _, e := range p {
		s += fmt.Sprint(e.Nickname, "/", len(e.Cards), ", ")
	}
	return
}

// take next player in the player list IF IDs NOT SHUFFELED IN THE LIST
func (ps Players) NextFrom(p *Player) *Player {
	np, ok := ps.ByID(p.ID + 1)
	if ok {
		return np
	} else {
		for i := range ps {
			if ps[i].ID > p.ID {
				return &ps[i]
			}
		}
		return &ps[0]
	}
}

// gets player by id
func (p *Players) ByID(id int) (*Player, bool) {
	for i := range *p {
		if (*p)[i].ID == id {
			return &(*p)[i], true
		}
	}
	return &Player{}, false
}

// take battle card for player by input
func (p *Player) GetBattleCard(input string) error {
	inputString := strings.Split(input, "")[0]
	inputNumber, err := strconv.Atoi(inputString)
	if err != nil {
		return err
	}
	p.BattleCard = p.Cards[inputNumber-1]
	p.Cards = append(p.Cards[:inputNumber-1], p.Cards[inputNumber:]...)
	return nil
}

// take battle card for bot
func (p *Player) BGetBattleCard() error {
	if len(p.Cards) == 0 {
		return fmt.Errorf("indexError: ")
	}
	number := rand.Intn(len(p.Cards))
	p.BattleCard = p.Cards[number]
	p.Cards = append(p.Cards[:number], p.Cards[number+1:]...)
	return nil
}

// shuffeles deck
func (d *Deck) Shuffle() {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(*d), func(i, j int) { (*d)[i], (*d)[j] = (*d)[j], (*d)[i] })
}

// creates deck
func (d *Deck) Create() {
	for i := 0; i != 9; i++ {
		for j := 0; j != 4; j++ {
			card := Card{i, j}
			*d = append(*d, card)
		}
	}
}

// creates bots and deck before main loop
func (s *Session) PlayersInit(players int) (err error) {
	randomNicknames := []string{"игрок 2", "игрок 3", "игрок 4", "игрок 5", "игрок 6", "игрок 7", "игрок 8"}

	if players < 2 || players >= 6 {
		return fmt.Errorf("wrongPlayerNumber: ")
	}

	s.Deck.Create()
	s.Deck.Shuffle()
	s.Trump = s.Deck[len(s.Deck)-1]

	for i := 0; i != players; i++ {
		s.Players = append(
			s.Players,
			Player{
				Cards:    s.Deck[0:6],
				Nickname: randomNicknames[i], //rand.Intn(len(randomNicknames))
				ID:       i})
		s.Deck = s.Deck[6:]
	}
	return nil
}

// understand should game ends
func (s *Session) IsFinish() bool {
	if len(s.Players) == 1 {
		s.Dumb = s.Players[0]
		return true
	}
	return false
}

// refill hand of player
func (s *Session) Refill(p *Player) {
	if len(p.Cards) < 6 {
		for i := 0; ; i++ {
			if len(s.Deck) == 0 {
				return
			}
			p.Cards = append(p.Cards, s.Deck[0])
			s.Deck = s.Deck[1:]
			if len(p.Cards) == 6 {
				return
			}
		}
	}
}

// conition for know winner and looser
func (s *Session) Battle() (string, error) {
	exhaust := []Card{}
	s.Turn += 1

	defenderWon := false
	exhaust = append(exhaust, s.Attacker.BattleCard, s.Defender.BattleCard)
	if s.Attacker.BattleCard.Class == s.Defender.BattleCard.Class && s.Attacker.BattleCard.Number != s.Defender.BattleCard.Number {
		if s.Attacker.BattleCard.Number < s.Defender.BattleCard.Number {
			defenderWon = true
		}
	} else if s.Attacker.BattleCard.Class != s.Defender.BattleCard.Class {
		if !(s.Attacker.BattleCard.Class == s.Trump.Class || s.Defender.BattleCard.Class != s.Trump.Class) {
			defenderWon = true
		}
	} else if s.Attacker.BattleCard.Class == s.Defender.BattleCard.Class && s.Attacker.BattleCard.Number == s.Defender.BattleCard.Number {
		return "", fmt.Errorf("doubleCard: ")
	} else {
		return "", fmt.Errorf("unknownError: ")
	}
	s.Attacker.BattleCard = Card{}
	s.Defender.BattleCard = Card{}

	s.Refill(s.Attacker)

	if defenderWon {
		s.Refill(s.Defender)
		s.Attacker = s.Defender
		s.Defender = s.Players.NextFrom(s.Attacker)
		return "", nil
	} else {
		st := s.Defender.Nickname
		s.Defender.Cards = append(exhaust, s.Defender.Cards...)
		s.Attacker = s.Players.NextFrom(s.Defender)
		s.Defender = s.Players.NextFrom(s.Attacker)
		return st, nil
	}
}

func (s *Session) SomeoneGone() (Players, bool) {
	ps := []Player{}
	if len(s.Deck) == 0 {
		for i, e := range s.Players {
			if len(e.Cards) == 0 {
				ps = append(ps, s.Players[i])
				s.Players = append(s.Players[:i], s.Players[i+1:]...)
			}
		}
	}
	if len(ps) != 0 {
		return ps, true
	}
	return ps, false
}

// outputs, i will change this
func (s Session) Stdout(me int) {
	fmt.Println("=-=-=-=-=-=-=-=-=-=-=-=-=")
	fmt.Println("игрок/карт:", s.Players)
	fmt.Println("ход:", s.Turn, "|", "колода:", len(s.Deck), "|", "козырь:", s.Trump)
	fmt.Println("атакует:", s.Attacker.Nickname)
	fmt.Println("защищается:", s.Defender.Nickname)
	if p, ok := s.Players.ByID(me); ok {
		fmt.Println("твоя рука:", p)
	}
}
