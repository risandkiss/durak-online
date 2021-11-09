package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type Card struct {
	Number int
	Class  int // 0,1,2,3 - крести, бубны, черви, пики
}

type Deck struct {
	Cards []Card
}

type Player struct {
	Nickname       string
	Cards          []Card
	AttackingCards []Card
}

type Session struct {
	Deck        Deck
	Used        []Card
	Players     []Player
	Turn        int
	BattleEnded bool
	Trump       Card
	Battles     []int
}

type Stringer interface {
	String() string
}

func (x Session) String() string {
	return fmt.Sprintf("%v", x.Deck)
}

func (x Deck) String() string {
	return fmt.Sprintf("%v", x.Cards)
}

func (x Card) String() string {
	return fmt.Sprintf("%v", x.Number)
}

func (deck *Deck) Shuffle() {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(deck.Cards), func(i, j int) { deck.Cards[i], deck.Cards[j] = deck.Cards[j], deck.Cards[i] })
}

func (deck *Deck) Create() {
	for i := 0; i != 9; i++ {
		for j := 0; j != 4; j++ {
			card := Card{i, j}
			deck.Cards = append(deck.Cards, card)
		}
	}
}

// create players by number
func (s *Session) PlayersInit(players int) (err error) {
	randomNicknames := []string{"плющь", "жидяра", "хуй", "шизоид", "лис", "фембой", "трапик"}
	randomSurnames := []string{"анальный", "жопный", "сильный", "вонючий", "непобедимый", "милый"}

	if players < 2 {
		return fmt.Errorf("not enough players")
	}
	if players >= 5 {
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

func (s *Session) Refill(playerNumber int) error {
	if len(s.Players[playerNumber].Cards) < 6 {
		for i := 0; ; i++ {
			if len(s.Deck.Cards) == 0 {
				return fmt.Errorf("deckIsEmpty: ")
			}
			s.Players[playerNumber].Cards = append(s.Players[playerNumber].Cards, s.Deck.Cards[0])
			s.Deck.Cards = s.Deck.Cards[1:]
			if len(s.Players[playerNumber].Cards) == 6 {
				return nil
			}
		}
	}
	return nil
}

// выводит имя карты, руки и так далее
func (s Session) View(c interface{}) string {
	class := []string{"♣", "♦", "♥", "♠"}
	number := []string{"6", "7", "8", "9", "10", "В", "Д", "К", "Т"}
	switch v := c.(type) {
	case Card:
		return number[v.Number] + class[v.Class]
	case Player:
		s := ""
		for i := range v.Cards {
			s += "[" + strconv.Itoa(i+1) + "]" + number[v.Cards[i].Number] + class[v.Cards[i].Class] + ", "
		}
		return s
	case []Card:
		s := ""
		for _, e := range v {
			s += number[e.Number] + class[e.Class] + " "
		}
		return s
	}
	return ""
}

// решает кто победил в сражении и кто получает мусорные карты
// ошибка используется для выхода из батла и исправлении ошибок игроками или мной
func (s *Session) Battle(attackingPlayer, defensingPlayer int) error {
	exhaust := []Card{}
	wonPlayer := -1
	aph := s.Players[attackingPlayer].AttackingCards
	dph := s.Players[defensingPlayer].AttackingCards

	if len(aph) != len(dph) {
		return fmt.Errorf("wrongHands: ")
	}

	counter := len(aph)
	for i := range aph {
		if aph[i].Class == dph[i].Class {
			if aph[i].Number == dph[i].Number {
				return fmt.Errorf("wrongDeck: ")
			} else if aph[i].Number < dph[i].Number {
				exhaust = append(exhaust, aph[i])
				exhaust = append(exhaust, dph[i])
				aph = remove(aph, i)
				dph = remove(dph, i)
				counter -= 1
			} else if aph[i].Number > dph[i].Number {
				exhaust = append(exhaust, aph[i])
				exhaust = append(exhaust, dph[i])
				aph = remove(aph, i)
				dph = remove(dph, i)
				break
			}
		} else {
			if aph[i].Class == s.Trump.Class && dph[i].Class != s.Trump.Class {
				exhaust = append(exhaust, aph[i])
				exhaust = append(exhaust, dph[i])
				aph = remove(aph, i)
				dph = remove(dph, i)
				break
			}
			if dph[i].Class == s.Trump.Class && aph[i].Class != s.Trump.Class {
				exhaust = append(exhaust, aph[i])
				exhaust = append(exhaust, dph[i])
				aph = remove(aph, i)
				dph = remove(dph, i)
				counter -= 1
			}
		}
	}

	s.Turn += 1
	s.Players[attackingPlayer].AttackingCards = []Card{}
	s.Players[defensingPlayer].AttackingCards = []Card{}

	if counter == 0 {
		wonPlayer = defensingPlayer
		err := s.Refill(attackingPlayer)
		if err != nil {
			return err
		}
		err = s.Refill(defensingPlayer)
		if err != nil {
			return err
		}
	} else {
		wonPlayer = attackingPlayer
		err := s.Refill(attackingPlayer)
		if err != nil {
			return err
		}
		s.Players[defensingPlayer].Cards = append(s.Players[defensingPlayer].Cards, exhaust...)
	}

	if wonPlayer != -1 {
		s.Battles = append(s.Battles, wonPlayer)
		return nil
	} else {
		return fmt.Errorf("unknownError: ")
	}
}

func remove(slice []Card, s int) []Card {
	return append(slice[:s], slice[s+1:]...)
}

// позволяет выбрать карты для атаки
func (s *Session) Attack(playerNumber int, input string) error {
	stringsInput := strings.Split(input, "")
	toHand := []Card{}

	for i, e := range stringsInput {
		v, err := strconv.Atoi(e)
		if err != nil {
			return err
		}

		toHand = append(toHand, s.Players[playerNumber].Cards[v-1])
		s.Players[playerNumber].Cards = remove(s.Players[playerNumber].Cards, v-1-i)
	}
	s.Players[playerNumber].AttackingCards = toHand
	return nil
}

// this function should create attack to bot
func (p *Player) AIAttack() error {
	numbers := []int{}

	numbers = append(numbers, rand.Intn(6))

	for i, e := range numbers {
		p.AttackingCards = append(p.AttackingCards, p.Cards[e])
		p.Cards = remove(p.Cards, e-i) // скорее всего что бы получить точное значение
		// нужно удалять не элемент по индексу, а раньше на i
	}

	return nil
}

func (s *Session) Stdout() string {
	return ""
}

//если сейчас первый ход, то сражается первый и второй игроки и тд
func main() {
	var session Session

	err := session.PlayersInit(2) // инициализация n игроков
	if err != nil {
		panic(err)
	}

	fmt.Println("козыри:", session.View(session.Trump))
	fmt.Println("враг по имени:", session.Players[0].Nickname)

	for session.Turn = 1; session.Turn != 100; {
		fmt.Println("-----------")
		fmt.Println("ход:", session.Turn)
		fmt.Println("карт в колоде:", len(session.Deck.Cards))
		fmt.Println("у врага карт:", len(session.Players[0].Cards))
		fmt.Println("твои карты:", session.View(session.Players[1]))

		if session.Turn%2 == 0 {
			err = session.Players[0].AIAttack()
			if err != nil {
				panic(err)
			}
			fmt.Println("тебя атакуют этим:", session.View(session.Players[0].AttackingCards))
			fmt.Println("выбери защиту из своих карт")
			fmt.Print("> ")

			input := ""
			fmt.Scan(&input)
			err = session.Attack(1, input)
			if err != nil {
				panic(err)
			}

			err = session.Battle(0, 1)
			if err == fmt.Errorf("deckIsEmpty: ") {
				fmt.Println("кто то победил, хз даже кто хд\nнажмите ентер или не ентер для выхода...")
				d := ""
				fmt.Scan(&d)
			} else if err != nil {
				panic(err)
			}

		} else {
			fmt.Println("выбери атаку из своих карт")
			fmt.Print("> ")

			input := ""
			fmt.Scan(&input)

			err = session.Attack(1, input)
			if err != nil {
				panic(err)
			}
			err = session.Players[0].AIAttack()
			if err != nil {
				panic(err)
			}

			fmt.Println("тебя отбили картой: ", session.View(session.Players[0].AttackingCards))

			err = session.Battle(1, 0)
			if err == fmt.Errorf("deckIsEmpty: ") {
				fmt.Println("кто то победил, хз даже кто хд\nнажмите ентер или не ентер для выхода...")
				d := ""
				fmt.Scan(&d)
			} else if err != nil {
				panic(err)
			}
		}

		if session.Battles[len(session.Battles)-1] == 1 {
			fmt.Println("битва за тобой")
		} else {
			fmt.Println("ты проиграл")
		}

	}
}
