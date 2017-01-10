package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/montanaflynn/stats"
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
	mongohost              = "10.0.0.95"
	selected         string
	everSelected     bool
)

func main() {

	if len(os.Args) > 1 {
		requestPerSecond, _ = strconv.Atoi(os.Args[1])
	}

	if len(os.Args) > 2 {
		sleepTm, _ = strconv.Atoi(os.Args[2])
	}

	if len(os.Getenv("MONGO_HOST")) > 0 {
		mongohost = os.Getenv("MONGO_HOST")
	}

	fmt.Println("will send", requestPerSecond, "requests per second")

	reader := bufio.NewReader(os.Stdin)

	session, err := mgo.Dial(mongohost + ":27017")
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
	fmt.Println("press a key to continue into the loop ...")
	reader.ReadLine()

	// mgo.SetDebug(true)
	// mgo.SetLogger(log.New(os.Stderr, "ff", 0))

	statsb := make(chan stats.Float64Data)
	go printStats(statsb)
	sb := make(stats.Float64Data, requestPerSecond)

	loop := 0
	for {
		fmt.Println("sleeping in a loop", loop)
		time.Sleep(1000 * time.Millisecond)
		loop++

		tmLoopBegin := time.Now()

		if loop == 2 && requestPerSecond > 1000 {
			// add more to time to establish sockets on a first run
			time.Sleep(time.Duration(requestPerSecond*2/1000) * time.Second)
		}
		if clients != 0 {
			log.Fatalln("id fell behind", clients)
		}

		for i := 0; i < requestPerSecond; i++ {
			if sleepTm > 0 {
				time.Sleep(time.Duration(sleepTm) * time.Microsecond)
			}
			go func(sess *mgo.Session, i int) {

				id := atomic.AddInt32(&clients, 1)
				// fmt.Println("+ID", id)

				ns := sess.Copy()

				result := Person{}
				c := ns.DB("hydra").C("users")
				tm_b := time.Now()
				err = c.Find(bson.M{"name": "Ale"}).One(&result)
				if err != nil {
					log.Println("ERROR on conn", i)
					log.Fatal(err)
				}

				sb[i] = time.Since(tm_b).Seconds()
				if !everSelected {
					selected = result.Phone
					fmt.Println("selected", selected)
					everSelected = true
				} else if selected != result.Phone {
					log.Fatalln("go unexpected select value", result.Phone)
				}
				// fmt.Println("Phone:", result.Phone, "took", sb[i], "seconds")

				id = atomic.AddInt32(&clients, -1)
				// fmt.Println("-ID", id)
				if id == 0 {
					tmLoopTotal := time.Since(tmLoopBegin).Seconds()
					fmt.Printf("took %.7f seconds per %d, avg %.7f seconds per request\n", tmLoopTotal, requestPerSecond, tmLoopTotal/float64(requestPerSecond))
					statsb <- sb
				}

				ns.Close()
			}(session, i)
		}

		// fmt.Println("press a key before the next loop...")
		// reader.ReadLine()

	}

	fmt.Println("press a key to exit...")
	reader.ReadLine()

}

func printStats(statsb <-chan stats.Float64Data) {

	for {
		select {
		case b := <-statsb:
			min, _ := stats.Min(b)
			mean, _ := stats.Mean(b)
			max, _ := stats.Max(b)
			fmt.Printf("min/mean/max %.7f/%.7f/%.7f\n", min, mean, max)
		}
	}
}
