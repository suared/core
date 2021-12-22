package queue

import (
	"fmt"
)

//Starting as private instead and will open up if needed later
//queue - should not be used directly, use StartQueue
//example: personQ := StartQueue(Person{}, processPerson, 3)
type queue [ msgType any ] struct {
	channel chan msgType
}
//Send - Use the return from StartQueue to call send
//example: personQ.Send(Person{Name: "David", Age: 5})
func(q *queue[msgType]) Send(msg msgType) {
	q.channel <- msg
}
//Start - should not be used directly, use StartQueue instead
func(q *queue[msgType]) start(fx func(msgType)) {
	for item := range q.channel {
		fx(item)
	}
}

//StartQueue - It is the caller's responsibility to ensure there is time to complete processing or use data from the embedded function to clean end.  Not done automatically as this is not always wanted/ intended for shutdown
//A simple sleep of a second or whatever is applicable to the use case at the end of running code is all that is needed in the exit point for now to allow sufficient processing time
//a more complex version using ctx cancel or other can be added if/ when makes sense.
//example: 	personQ := StartQueue(Person{}, processPerson, 3)
//  		personQ.Send(Person{Name: "David", Age: 5})
//The function run to process the messages is the function passed, in this case processPerson
//The buffer size shoud be 1 or greater to enable a buffered channel using standard Go channel semantics
//While "theType" is not used directly, user level code makes more sense this way to not escape the generics semantics outside of the library
func StartQueue[T any] (theType T, fx func(T), buffer int) (*queue[T]) {

	fmt.Println("change test")
	//Make a channel of the right type
	channel := make(chan T, buffer)
	q := &queue[T]{}
	q.channel = channel

	//Start a processor - simple queue, 1 receiver only
	go q.start(fx)
	
	//Build a return to enable sends 
	return q
}


