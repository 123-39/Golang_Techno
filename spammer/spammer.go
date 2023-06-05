package main

import (
	"log"
	"sort"
	"strconv"
	"sync"
)

// максимальное число параллельных потоков для антиспама
const maxGoroutinesNum = 5

// размер батча для GetUser()
const batchSize = 2

func RunPipeline(cmds ...cmd) {

	in := make(chan interface{})
	out := make(chan interface{})
	wg := &sync.WaitGroup{}

	for _, command := range cmds {
		wg.Add(1)
		go func(wg *sync.WaitGroup, command cmd, in, out chan interface{}) {
			command(in, out)
			wg.Done()
			close(out)
		}(wg, command, in, out)
		in = out                                // выходные данные -> входные данные для следующей команды
		out = make(chan interface{}, batchSize) // создаем новый канал для записи
	}
	wg.Wait()

}

func SelectUsers(in, out chan interface{}) {
	// 	in - string
	// 	out - User
	var wg = &sync.WaitGroup{}
	mu := &sync.Mutex{}
	users := make(map[uint64]string) // map для контроля уникальности юзеров
	var currentUser User

	for email := range in {
		wg.Add(1)
		email := email
		go func(out chan interface{}) {
			defer wg.Done()
			currentUser = GetUser(email.(string))
			_, keyExist := users[currentUser.ID]
			if !keyExist { // контроль уникальности юзеров
				mu.Lock()
				users[currentUser.ID] = currentUser.Email
				mu.Unlock()
				out <- currentUser
			}
		}(out)
	}
	wg.Wait()
}

func SelectMessages(in, out chan interface{}) {
	// 	in - User
	// 	out - MsgID
	var wg = &sync.WaitGroup{}
	var nextPerson interface{}

	for person := range in {
		wg.Add(1)
		person := person.(User)
		go func(out chan interface{}) {
			defer wg.Done()
			var morePerson []User
			morePerson = append(morePerson, person)
			for i := 0; i < batchSize-1; i++ {
				nextPerson = <-in
				if nextPerson != nil {
					morePerson = append(morePerson, nextPerson.(User))
				}
			}
			messages, errUser := GetMessages(morePerson...)
			if errUser != nil {
				log.Printf("err:%v", errUser)
			}
			out <- messages
		}(out)
	}
	wg.Wait()
}

func CheckSpamWorker(outValue MsgData, wg *sync.WaitGroup,
	in <-chan interface{}, out chan interface{}) {
	defer func() {
		wg.Done()
	}()

	for messageID := range in {
		isSpam, errSpam := HasSpam(messageID.(MsgID)) // запускаем антиспам
		if errSpam != nil {
			log.Printf("err:%v", errSpam)
		}
		outValue.HasSpam = isSpam
		outValue.ID = messageID.(MsgID)
		out <- outValue
	}
}

func CheckSpam(in, out chan interface{}) {
	// in - MsgID
	// out - MsgData
	var outValue MsgData // структура с парой полей: id и факт того является ли письмо спамом
	var wg = &sync.WaitGroup{}
	wg.Add(maxGoroutinesNum)
	workerInput := make(chan interface{})

	for i := 0; i < maxGoroutinesNum; i++ {
		go CheckSpamWorker(outValue, wg, workerInput, out)
	}
	for listMessagesID := range in {
		for _, messageID := range listMessagesID.([]MsgID) {
			workerInput <- messageID
		}
	}
	close(workerInput)
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	// in - MsgData
	// out - string
	// данные вида "<has_spam> <msg_id>"
	var allMessageContent = map[bool][]MsgID{
		true:  {},
		false: {},
	}
	var wg = &sync.WaitGroup{}
	mu := &sync.Mutex{}

	for messageInfo := range in {
		wg.Add(1)
		messageInfo := messageInfo.(MsgData)
		go func(messageInfo MsgData) {
			defer wg.Done()
			mu.Lock()
			allMessageContent[messageInfo.HasSpam] = append(allMessageContent[messageInfo.HasSpam],
				messageInfo.ID)
			mu.Unlock()
		}(messageInfo)
	}
	wg.Wait()

	sort.Slice(allMessageContent[true], func(i, j int) bool {
		return allMessageContent[true][i] < allMessageContent[true][j]
	})
	sort.Slice(allMessageContent[false], func(i, j int) bool {
		return allMessageContent[false][i] < allMessageContent[false][j]
	})
	for _, messageID := range allMessageContent[true] {
		out <- "true " + strconv.FormatUint(uint64(messageID), 10)
	}
	for _, messageID := range allMessageContent[false] {
		out <- "false " + strconv.FormatUint(uint64(messageID), 10)
	}

}

func main() {
}
