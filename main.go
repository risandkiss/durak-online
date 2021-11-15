package main

import (
	"durak-online/engine"
	"fmt"
	"log"
)

// main
func main() {
	var session engine.Session
	err := session.PlayersInit(5) // init N players
	if err != nil {
		log.Fatal(err)
	}

	me := 0
	var ok bool
	session.Attacker, ok = session.Players.ByID(me)
	if !ok {
		log.Fatal("no players")
	}
	session.Defender, ok = session.Players.ByID(1)
	if !ok {
		log.Fatal("no players")
	}

	fmt.Print("\nвведите свой никнейм: ")
	fmt.Scan(&session.Players[me].Nickname)
	fmt.Println()

	for session.Turn = 1; session.Turn != 100; {
		session.Stdout(me)

		if session.Defender.ID == me {
			err = session.Attacker.BGetBattleCard()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("враг:", session.Attacker.BattleCard)
			fmt.Println("защищайся> ")

			input := ""
			fmt.Scan(&input)
			err = session.Defender.GetBattleCard(input)
			if err != nil {
				log.Fatal(err)
			}
		} else if session.Attacker.ID == me {
			fmt.Println("атакуй> ") // тут было бы прикольно сделать вместо ввода цифры сразу карту и vs карта врага
			input := ""
			fmt.Scan(&input)
			err = session.Attacker.GetBattleCard(input)
			if err != nil {
				log.Fatal(err)
			}

			err = session.Defender.BGetBattleCard()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("враг: ", session.Defender.BattleCard)
		} else {
			err = session.Attacker.BGetBattleCard()
			if err != nil {
				log.Fatal("AttackerError: ", err)
			}
			err = session.Defender.BGetBattleCard()
			if err != nil {
				log.Fatal("DefenderError: ", err)
			}
			fmt.Println(session.Attacker.Nickname, "против", session.Defender.Nickname)
		}

		res, err := session.Battle()
		if err != nil {
			log.Fatal(err)
		}

		gone, yes := session.SomeoneGone()

		if res == "" {
			fmt.Println("бито")
		} else {
			fmt.Println("игрок", res, "забрал")
		}

		if session.IsFinish() {
			fmt.Println("игра завершена")
			fmt.Println("дурак -", session.Dumb.Nickname)
			break
		}

		if session.Attacker == session.Defender {
			log.Fatal("attacker is defender")
		}

		if yes {
			fmt.Println("игроки", gone, "выбыли")
		}
	}

	fmt.Println("игра окончена всем спасибо")
}
