package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Person struct {
	Name  string
	Phone string
}

var (
	requestPerSecond       = 10
	clients          int32 = 0
	sleepTm                = 0
)

func main() {

	if len(os.Args) > 1 {
		requestPerSecond, _ = strconv.Atoi(os.Args[1])
	}

	if len(os.Args) > 2 {
		sleepTm, _ = strconv.Atoi(os.Args[2])
	}

	fmt.Println("will send", requestPerSecond, "requests per second")

	reader := bufio.NewReader(os.Stdin)

	// session, err := mgo.Dial("localhost")
	session, err := mgo.Dial("10.0.0.95:27017")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	fmt.Println("created one session, press a key...")
	reader.ReadLine()

	newS := session.Copy()
	defer newS.Close()

	fmt.Println("created second session, press a key...")
	reader.ReadLine()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("hydra").C("users")
	err = c.Insert(&Person{"Ale", "+55 53 8116 9639"},
		&Person{"Cla", "+55 53 8402 8510"})
	// err = c.Update(&Person)
	if err != nil {
		log.Fatal(err)
	}

	result := Person{}
	err = newS.DB("hydra").C("users").Find(bson.M{"name": "Ale"}).One(&result)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Phone:", result.Phone)
	fmt.Println("press a key to contunue into the loop ...")
	reader.ReadLine()

	// mgo.SetDebug(true)
	mgo.SetLogger(log.New(os.Stderr, "ff", 0))

	loop := 0
	for {
		fmt.Println("sleeping in a loop", loop)
		time.Sleep(1000 * time.Millisecond)
		loop++

		for i := 0; i < requestPerSecond; i++ {
			if sleepTm > 0 {
				time.Sleep(time.Duration(sleepTm) * time.Microsecond)
			}
			go func(sess *mgo.Session) {

				fmt.Println("+ID", atomic.AddInt32(&clients, 1))

				ns := sess.Copy()

				result := Person{}
				err = ns.DB("hydra").C("users").Find(bson.M{"name": "Ale"}).One(&result)
				if err != nil {
					log.Println("ERROR on conn", i)
					log.Fatal(err)
				}

				fmt.Println("Phone:", result.Phone)

				fmt.Println("-ID", atomic.AddInt32(&clients, -1))

				ns.Close()
			}(session)
		}

		fmt.Println("press a key before the next loop...")
		// reader.ReadLine()

	}

	fmt.Println("press a key to exit...")
	reader.ReadLine()

}
