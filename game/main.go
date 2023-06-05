package main

import (
	"fmt"
	"strings"
)

type User struct {
	storage     map[string]bool // хранилище (в данном случае рюкзак)
	currentRoom string          // комната, в которой находится игрок в данный момент
}

var player User

type World struct {
	room2roomConnect          map[string][]string        // в какие комнаты можно попасть из данной
	room2itemConnect          map[string][]string        // какие предметы присутствуют в данной комнате
	applicationItem           map[string]map[string]bool // к чему можно применить предметы в данной комнате
	room2storageConnect       map[string]map[string]bool // наличие хранилищ (в данном случае рюкзака) в данной комнате
	room2phraseConnect        map[string]string          // уникальные фразы для каждой комнаты
	room2doorConnection       map[string]bool            // наличие дверей в комнатах
	room2lookAroundConnection map[string]interface{}     // реакция на команду "осмотреться"
}

var currentWorld World

func main() {
	initGame()
	continueInput := "yes"
	for continueInput != "no" {
		fmt.Println("Input command: ")
		var newCommand, command, par1, par2, par3 string
		fmt.Scanf("%s %s %s %s", &command, &par1, &par2, &par3)
		newCommand = command + " " + par1 + " " + par2 + " " + par3
		fmt.Println(handleCommand(newCommand))
		fmt.Println("Do you wanna continue? (yes/no): ")
		fmt.Scanf("%s\n", &continueInput)
	}
}

func lookAround() string {

	answer := currentWorld.room2lookAroundConnection[player.currentRoom].(func() string)()
	availableRoom := currentWorld.room2roomConnect[player.currentRoom] // доступные пути
	Rooms := strings.Join(availableRoom, ", ")
	answer += ". можно пройти - " + Rooms

	return answer
}

func going(nextRoom string) string {

	answer := ""

	availableNextRoom, nextRoomExist := currentWorld.room2roomConnect[nextRoom] // существует ли комната, куда нужно идти
	availableRoom := currentWorld.room2roomConnect[player.currentRoom]          // возможные комнаты из текущей
	wayExist := false                                                           // существует ли путь в комнату
	for _, room := range availableRoom {
		if nextRoom == room {
			wayExist = true
		}
	}

	if nextRoomExist && wayExist {
		if currentWorld.room2doorConnection[nextRoom] { // есть ли дверь в следующую комнату
			for item, itemReady := range currentWorld.applicationItem[player.currentRoom] { // открыты ли двери
				if !itemReady {
					answer += item + " закрыта"
					return answer
				}
			}
		}
		Rooms := strings.Join(availableNextRoom, ", ")
		answer := currentWorld.room2phraseConnect[nextRoom] + "можно пройти - " + Rooms
		player.currentRoom = nextRoom
		return answer
	}
	return "нет пути в " + nextRoom
}

func putOn(putOnItem string) string {

	alreadyPutOn, itemExist := currentWorld.room2storageConnect[player.currentRoom][putOnItem]
	if !alreadyPutOn && itemExist { // есть ли предмет в комнате и не надет ли еще
		player.storage = map[string]bool{}                                     // надеваем рюкзак
		currentWorld.room2storageConnect[player.currentRoom][putOnItem] = true // удаляем предмет из доступа
		return "вы надели: " + putOnItem
	}
	return "нет такого"
}

func taking(takeItem string) string {

	itemPosition := -1
	items := currentWorld.room2itemConnect[player.currentRoom] // какие предметы есть в комнате
	for i, item := range items {
		if item == takeItem {
			itemPosition = i
		}
	}
	if itemPosition != -1 { // не в рюкзаке ли предмет и существует ли
		if player.storage != nil { // надет ли рюкзак
			player.storage[takeItem] = true                                                                             // кладем в рюкзак
			currentWorld.room2itemConnect[player.currentRoom] = append(items[:itemPosition], items[itemPosition+1:]...) // удаляем предмет из комнаты
			return "предмет добавлен в инвентарь: " + takeItem
		}
		return "некуда класть"
	}
	return "нет такого"
}

func apply(applyWhat string, applyTo string) string {
	_, itemExist := player.storage[applyWhat] // есть ли предмет в инвентаре
	if itemExist {
		_, applicable := currentWorld.applicationItem[player.currentRoom][applyTo] // есть к чему применять?
		if applicable {                                                            // есть к чему применять?
			currentWorld.applicationItem[player.currentRoom][applyTo] = true // помечаем, как примененное
			return applyTo + " открыта"
		}
		return "не к чему применить"
	}
	return "нет предмета в инвентаре - " + applyWhat
}

func initGame() {
	/*
		эта функция инициализирует игровой мир - все комнаты
		если что-то было - оно корректно перезатирается
	*/
	currentWorld.room2roomConnect = map[string][]string{
		"кухня":   {"коридор"},
		"коридор": {"кухня", "комната", "улица"},
		"комната": {"коридор"},
		"улица":   {"домой"},
	}
	currentWorld.room2itemConnect = map[string][]string{
		"кухня":   {},
		"коридор": {},
		"комната": {"ключи", "конспекты"},
	}
	currentWorld.applicationItem = map[string]map[string]bool{
		"коридор": {"дверь": false},
	}
	currentWorld.room2storageConnect = map[string]map[string]bool{
		"комната": {"рюкзак": false},
	}
	currentWorld.room2phraseConnect = map[string]string{
		"кухня":   "кухня, ничего интересного. ",
		"коридор": "ничего интересного. ",
		"комната": "ты в своей комнате. ",
		"улица":   "на улице весна. ",
	}
	currentWorld.room2doorConnection = map[string]bool{
		"кухня":   false,
		"коридор": false,
		"комната": false,
		"улица":   true,
	}
	currentWorld.room2lookAroundConnection = map[string]interface{}{
		"кухня": func() string {
			answer := "ты находишься на кухне, на столе: чай, надо "
			itemsInRooms := 0
			for _, items := range currentWorld.room2itemConnect { // остались ли предметы в комнатах
				itemsInRooms += len(items)
			}
			if itemsInRooms > 0 {
				answer += "собрать рюкзак и "
			}
			answer += "идти в универ"
			return answer
		},
		"коридор": func() string {
			return "пустая комната"
		},
		"комната": func() string {
			answer := ""
			itemInRoom := currentWorld.room2itemConnect[player.currentRoom] // какие предметы есть в комнате
			if len(itemInRoom) > 0 {
				Items := strings.Join(itemInRoom, ", ")
				answer += "на столе: " + Items
			} else {
				answer += "пустая комната"
			}

			storageInRoom := []string{}
			for storage, itemExist := range currentWorld.room2storageConnect[player.currentRoom] { // остались ли не надетые хранилища
				if !itemExist {
					storageInRoom = append(storageInRoom, storage)
				}
			}
			if len(storageInRoom) > 0 {
				Storages := strings.Join(storageInRoom, ", ")
				answer += ", на стуле: " + Storages
			}
			return answer
		},
		"улица": func() string {
			return "пустая комната"
		},
	}

	player.storage = nil
	player.currentRoom = "кухня"

}

func handleCommand(command string) string {
	/*
		данная функция принимает команду от "пользователя"
		и наверняка вызывает какой-то другой метод или функцию у "мира" - списка комнат
	*/
	parseСommand := strings.Split(command, " ")

	action := parseСommand[0]
	switch action {
	case "осмотреться":
		return lookAround()
	case "идти":
		nextRoom := parseСommand[1]
		return going(nextRoom)
	case "надеть":
		item := parseСommand[1]
		return putOn(item)
	case "взять":
		takeItem := parseСommand[1]
		return taking(takeItem)
	case "применить":
		applyWhat := parseСommand[1]
		applyTo := parseСommand[2]
		return apply(applyWhat, applyTo)
	default:
		return "неизвестная команда"
	}
}
