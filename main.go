package main

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

type Deck struct {
	Cards []Card // list of cards
}

type Player struct {
	Nickname       string // nickname
	Cards          []Card // hand
	AttackingCards []Card // attacking hand
	AttackCard     Card
	IsBot          bool // isBot
}

type Session struct {
	Deck       Deck     // deck contains not used cards, deck needs to give new cards to players
	Players    []Player // list of players
	Turn       int      // number of turn
	Trump      Card     // copy of trump card
	WonPlayers []Player // list of won players
	APNumber   int      // attacking player now
	DPNumber   int      // defensing player now
}

type Stringer interface {
	String() string
}

func (c Card) String() string {
	class := []string{"♣", "♦", "♥", "♠"}
	number := []string{"6", "7", "8", "9", "10", "J", "Q", "K", "A"}
	return fmt.Sprint(number[c.Number] + class[c.Class])
}

func (p Player) String() string {
	s := ""
	for i, e := range p.Cards {
		s += fmt.Sprint("[", i+1, "]", e.String(), " ")
	}
	return s
}

func (d *Deck) Shuffle() {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(d.Cards), func(i, j int) { d.Cards[i], d.Cards[j] = d.Cards[j], d.Cards[i] })
}

func (d *Deck) Create() {
	for i := 0; i != 9; i++ {
		for j := 0; j != 4; j++ {
			card := Card{i, j}
			d.Cards = append(d.Cards, card)
		}
	}
}

// create players by number
func (s *Session) PlayersInit(players int) (err error) {
	randomNicknames := []string{"плющь", "жидяра", "хуй", "шизоид", "лис", "фембой", "трапик"}
	randomSurnames := []string{"анальный", "жопный", "сильный", "вонючий", "непобедимый", "милый"}

	if players < 2 {
		return fmt.Errorf("not enough players")
	} else if players >= 5 {
		return fmt.Errorf("so enough players")
	}

	s.Deck.Create()
	s.Deck.Shuffle()
	s.Trump = s.Deck.Cards[len(s.Deck.Cards)-1]

	for i := 0; i != players; i++ {
		s.Players = append(
			s.Players,
			Player{
				Cards:    s.Deck.Cards[0:6],
				Nickname: randomSurnames[rand.Intn(len(randomSurnames))] + " " + randomNicknames[rand.Intn(len(randomNicknames))]})
		s.Deck.Cards = s.Deck.Cards[6:]
	}
	return nil
}

func (s *Session) Refill(playerNumber int) {
	if len(s.Players[playerNumber].Cards) < 6 {
		for i := 0; ; i++ {
			if len(s.Deck.Cards) == 0 {
				return
			}
			s.Players[playerNumber].Cards = append(s.Players[playerNumber].Cards, s.Deck.Cards[0])
			s.Deck.Cards = s.Deck.Cards[1:]
			if len(s.Players[playerNumber].Cards) == 6 {
				return
			}
		}
	}
	return
}

// Battle check who will win
func (s *Session) Battle() (int, error) {
	attackingPlayer := s.APNumber
	defensingPlayer := s.DPNumber
	exhaust := []Card{}
	s.Turn += 1

	apc := s.Players[attackingPlayer].AttackCard
	dpc := s.Players[defensingPlayer].AttackCard

	s.Players[defensingPlayer].AttackCard = Card{}
	s.Players[attackingPlayer].AttackCard = Card{}

	counter := 1
	exhaust = append(exhaust, apc, dpc)
	if apc.Class == dpc.Class && apc.Number != dpc.Number {
		if apc.Number < dpc.Number {
			counter -= 1
		}
	} else if apc.Class != dpc.Class {
		if !(apc.Class == s.Trump.Class || dpc.Class != s.Trump.Class) {
			counter -= 1
		}
	} else {
		return 0, fmt.Errorf("doubleCard: ")
	}

	s.Refill(attackingPlayer)

	if counter == 0 {
		s.Refill(defensingPlayer)
		s.APNumber += 1
		return -1, nil
	} else {
		s.Players[defensingPlayer].Cards = append(exhaust, s.Players[defensingPlayer].Cards...)
		return defensingPlayer, nil
	}
}

func remove(slice []Card, s int) []Card {
	return append(slice[:s], slice[s+1:]...)
}

// choose card
func (p *Player) GetAttackCard(input string) error {
	inputString := strings.Split(input, "")[0]
	inputNumber, err := strconv.Atoi(inputString)
	if err != nil {
		return err
	}
	p.AttackCard = p.Cards[inputNumber-1]
	p.Cards = remove(p.Cards, inputNumber-1)
	return nil
}

// bot takes card to attack
func (p *Player) BGetAttackCard() error {
	number := rand.Intn(6)
	p.AttackCard = p.Cards[number]
	p.Cards = remove(p.Cards, number)
	return nil
}

func (s *Session) Stdout() {
	fmt.Println("-----------")
	fmt.Println("атакующий игрок:", s.Players[s.APNumber].Nickname)
	fmt.Println("защищающийся игрок:", s.Players[s.DPNumber].Nickname)
	fmt.Println("козыри:", s.Trump)
	fmt.Println("ход:", s.Turn)
	fmt.Println("карт в колоде:", len(s.Deck.Cards))
	fmt.Println("у врага карт:", len(s.Players[1].Cards))
	fmt.Println("твои карты:", s.Players[0])
}

//
func main() {
	var session Session
	err := session.PlayersInit(2) // init N players
	if err != nil {
		panic(err)
	}

	firstBot := 1
	me := 0

	fmt.Println("козыри:", session.Trump)
	fmt.Println("враги по имени:", session.Players[firstBot].Nickname) // need loop
	fmt.Print("\nвведите свой никнейм: ")
	fmt.Scan(&session.Players[me].Nickname)
	fmt.Println()

	for session.Turn = 1; session.Turn != 100; {
		if len(session.Deck.Cards) == 0 {
			for i, e := range session.Players {
				if len(e.Cards) == 0 {
					fmt.Println("игрок", e.Nickname, "выбыл")
					session.WonPlayers = append(session.WonPlayers, e)
					session.Players = append(session.Players[:i], session.Players[i+1:]...)
				}
			}
		}
		if len(session.Players) == 1 {
			fmt.Println("игра завершена")
			fmt.Println("побидитель", session.WonPlayers[0].Nickname)
			break
		}

		session.Stdout()

		if session.APNumber >= len(session.Players) {
			session.APNumber = 0
		}
		if session.APNumber >= len(session.Players)-1 {
			session.DPNumber = 0
		} else {
			session.DPNumber = session.APNumber + 1
		}

		if session.APNumber == firstBot {
			err = session.Players[session.APNumber].BGetAttackCard()
			if err != nil {
				panic(err)
			}
			fmt.Println("тебя атакуют этим:", session.Players[session.APNumber].AttackCard)
			fmt.Println("выбери защиту из своих карт")
			fmt.Print("> ")

			input := ""
			fmt.Scan(&input)
			err = session.Players[session.DPNumber].GetAttackCard(input)
			if err != nil {
				panic(err)
			}

			res, err := session.Battle()
			if err != nil {
				panic(err)
			}

			if res == -1 {
				fmt.Println("бито")
			} else {
				fmt.Println("игрок", session.Players[res], "проиграл")
			}
		} else if session.APNumber == me {
			fmt.Println("выбери атаку из своих карт")
			fmt.Print("> ")

			input := ""
			fmt.Scan(&input)

			err = session.Players[session.APNumber].GetAttackCard(input)
			if err != nil {
				panic(err)
			}
			err = session.Players[session.DPNumber].BGetAttackCard()
			if err != nil {
				panic(err)
			}

			fmt.Println("тебя отбили картой: ", session.Players[session.DPNumber].AttackCard)

			res, err := session.Battle()
			if err != nil {
				panic(err)
			}

			if res == -1 {
				fmt.Println("бито")
			} else {
				fmt.Println("игрок", session.Players[res], "проиграл")
			}
		}
	}
	fmt.Println("хуй")
}
