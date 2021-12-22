package test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/suared/core/queue"
)

var testcounter int

type Message struct {
	Msg string
}

func (msg Message) String() string {
	return msg.Msg
}

type Person struct {
	Name string
	Age  int
}

func (person Person) String() string {
	return person.Name + ", age:" + strconv.Itoa(person.Age)
}

func processMessage(m Message) {
	fmt.Println(m.Msg)
	testcounter++
}

func processPerson(p Person) {
	fmt.Println(p.Name + ", age:" + strconv.Itoa(p.Age))
	testcounter++
}

func TestQueue(t *testing.T) {
	msgQ := queue.StartQueue(Message{}, processMessage, 2)
	msgQ.Send(Message{Msg: "test1"})
	msgQ.Send(Message{Msg: "test2"})
	msgQ.Send(Message{Msg: "test3"})
	msgQ.Send(Message{Msg: "test4"})
	msgQ.Send(Message{Msg: "test5"})

	personQ := queue.StartQueue(Person{}, processPerson, 3)
	personQ.Send(Person{Name: "David", Age: 5})
	personQ.Send(Person{Name: "Colleen", Age: 8})
	personQ.Send(Person{Name: "Charlie", Age: 2})
	personQ.Send(Person{Name: "Liam", Age: 1})
	personQ.Send(Person{Name: "Rexie", Age: 10})
	personQ.Send(Person{Name: "Julie", Age: 7})
	personQ.Send(Person{Name: "Mike", Age: 13})
	personQ.Send(Person{Name: "Fleming", Age: 19})

	//It is the caller's responsibility to ensure there is time to complete processing, for this simple case...
	time.Sleep(time.Second * 1)

	//confirm all processed
	if testcounter != 13 {
		t.Error("Expected 13 items, have " + strconv.Itoa(testcounter))
	}

}
