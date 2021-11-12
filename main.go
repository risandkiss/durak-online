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

type Deck []Card

type Player struct {
	Nickname       string // nickname
	Cards          []Card // hand
	AttackingCards []Card // attacking hand
	AttackCard     Card
	IsActive       bool
}

type Players []Player

type Session struct {
	Deck                 Deck    // deck contains not used cards, deck needs to give new cards to players
	Players              Players // list of players
	Turn                 int     // number of turn
	Trump                Card    // copy of trump card
	Dumb                 Player  // list of won players
	APNumber             int     // attacking player now
	DPNumber             int     // defensing player now
	NumberOfActivePlayer int
}

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
		s += fmt.Sprint(e.Nickname, ", ")
	}
	return
}

// if game should finish return true
func (s *Session) IsFinish() bool {
	c := 0
	pl := Player{}
	for _, e := range s.Players {
		if e.IsActive {
			pl = e
			c++
		}
	}
	if c == 1 { // если с < 1 то это ошибка, но мне похуй
		s.Dumb = pl
		return true
	}
	return false
}

func (d *Deck) Shuffle() {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(*d), func(i, j int) { (*d)[i], (*d)[j] = (*d)[j], (*d)[i] })
}

func (d *Deck) Create() {
	for i := 0; i != 9; i++ {
		for j := 0; j != 4; j++ {
			card := Card{i, j}
			*d = append(*d, card)
		}
	}
}

// create players by number
func (s *Session) PlayersInit(players int) (err error) {
	randomNicknames := []string{"плющь", "жидяра", "хуй", "шизоид", "лис", "фембой", "трапик"}
	randomSurnames := []string{"анальный", "жопный", "сильный", "вонючий", "непобедимый", "милый"}

	if players < 2 {
		return fmt.Errorf("not enough players")
	} else if players >= 6 {
		return fmt.Errorf("so enough players")
	}

	s.Deck.Create()
	s.Deck.Shuffle()
	s.Trump = s.Deck[len(s.Deck)-1]

	for i := 0; i != players; i++ {
		s.Players = append(
			s.Players,
			Player{
				Cards:    s.Deck[0:6],
				Nickname: randomSurnames[rand.Intn(len(randomSurnames))] + " " + randomNicknames[rand.Intn(len(randomNicknames))],
				IsActive: true})
		s.Deck = s.Deck[6:]
	}
	return nil
}

func (s *Session) Refill(playerNumber int) {
	if len(s.Players[playerNumber].Cards) < 6 {
		for i := 0; ; i++ {
			if len(s.Deck) == 0 {
				return
			}
			s.Players[playerNumber].Cards = append(s.Players[playerNumber].Cards, s.Deck[0])
			s.Deck = s.Deck[1:]
			if len(s.Players[playerNumber].Cards) == 6 {
				return
			}
		}
	}
}

// if wins defender then true
func (s *Session) Battle() (bool, error) {
	attackingPlayer := s.APNumber
	defensingPlayer := s.DPNumber
	exhaust := []Card{}
	s.Turn += 1

	apc := s.Players[attackingPlayer].AttackCard
	dpc := s.Players[defensingPlayer].AttackCard
	s.Players[defensingPlayer].AttackCard = Card{}
	s.Players[attackingPlayer].AttackCard = Card{}

	counter := false
	exhaust = append(exhaust, apc, dpc)
	if apc.Class == dpc.Class && apc.Number != dpc.Number {
		if apc.Number < dpc.Number {
			counter = true
		}
	} else if apc.Class != dpc.Class {
		if !(apc.Class == s.Trump.Class || dpc.Class != s.Trump.Class) {
			counter = true
		}
	} else {
		return false, fmt.Errorf("doubleCard: ")
	}

	s.Refill(attackingPlayer)

	if counter {
		s.Refill(defensingPlayer)
		return true, nil
	} else {
		s.Players[defensingPlayer].Cards = append(exhaust, s.Players[defensingPlayer].Cards...)
		return false, nil
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
	number := rand.Intn(len(p.Cards))
	p.AttackCard = p.Cards[number]
	p.Cards = remove(p.Cards, number)
	return nil
}

func (s Session) Stdout(me int) {
	fmt.Println("-----------")
	fmt.Println("атакующий игрок:", s.Players[s.APNumber].Nickname)
	fmt.Println("защищающийся игрок:", s.Players[s.DPNumber].Nickname)
	fmt.Println("козыри:", s.Trump)
	fmt.Println("ход:", s.Turn)
	fmt.Println("карт в колоде:", len(s.Deck))
	for _, e := range s.Players[1:] {
		fmt.Println("у", e.Nickname, len(e.Cards), "карт")
	}
	fmt.Println("твои карты:", s.Players[me])
}

func (s *Session) SomeoneGone() (Player, bool) {
	if len(s.Deck) == 0 {
		for i, e := range s.Players {
			if len(e.Cards) == 0 {
				s.Players[i].IsActive = false
				s.NumberOfActivePlayer--
				return s.Players[i], true
			}
		}
	}
	return Player{}, false
}

//
func main() {
	var session Session
	err := session.PlayersInit(5) // init N players
	if err != nil {
		panic(err)
	}

	me := 0
	fmt.Print("\nвведите свой никнейм: ")
	fmt.Scan(&session.Players[me].Nickname)
	fmt.Println()

	fmt.Println("козыри:", session.Trump)
	fmt.Println("игроки:", session.Players) // need loop

	session.APNumber = 0
	session.DPNumber = 1

	session.NumberOfActivePlayer = len(session.Players)

	for session.Turn = 1; session.Turn != 100; {
		if gone, yes := session.SomeoneGone(); yes {
			fmt.Println("игрок", gone.Nickname, "выбыл")
		}

		if session.IsFinish() {
			fmt.Println("игра завершена")
			fmt.Println("дурак -", session.Dumb.Nickname)
			break
		}

		session.Stdout(me)

		if session.APNumber != me && session.DPNumber == me {
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
		} else if session.APNumber == me && session.DPNumber != me {
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
			fmt.Println("тебя попытались отбить: ", session.Players[session.DPNumber].AttackCard)

		} else {
			err = session.Players[session.APNumber].BGetAttackCard()
			if err != nil {
				panic(err)
			}
			err = session.Players[session.DPNumber].BGetAttackCard()
			if err != nil {
				panic(err)
			}
		}

		res, err := session.Battle()
		if err != nil {
			panic(err)
		}

		if res {
			fmt.Println("бито")
			session.APNumber++
			session.DPNumber++
		} else {
			fmt.Println("игрок", session.Players[session.DPNumber].Nickname, "забрал")
			session.APNumber += 2
			session.DPNumber += 2
		}

		if session.APNumber >= session.NumberOfActivePlayer {
			session.APNumber = session.APNumber - session.NumberOfActivePlayer
		}
		if session.DPNumber >= session.NumberOfActivePlayer {
			session.DPNumber = session.DPNumber - session.NumberOfActivePlayer
		}

		// if !(session.DPNumber-session.APNumber == 1 && (session.DPNumber == 0 && session.APNumber == session.NumberOfActivePlayer-1)) {
		// 	fmt.Println("пиздец")
		// 	break
		// }

		//fmt.Println(session.APNumber, session.DPNumber)
	}
	fmt.Println("игра окончена всем спасибо")
}
