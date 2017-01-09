package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
)

/*
 * db.users.save( {username:"mkyong"} )
 * db.createUser( { "user" : "mdbhydra", "pwd" : "hydra", "roles" : [ { role: "readWrite", db: "hydra" } ] } )
 */

var nfiles uint

func main() {

	m, _ := strconv.ParseInt(os.Args[1], 10, 0)

	// for i := 0; i < int(m); i++ {
	// 	if _, err := os.Open("testfile.txt"); err != nil {
	// 		fmt.Println("err", err)
	// 		fmt.Println("files open", nfiles)
	// 		return
	// 	}
	// 	nfiles++
	// }

	for i := 0; i < int(m); i++ {
		if _, err := net.Dial("tcp", "localhost:27017"); err != nil {
			fmt.Println("err", err)
			fmt.Println("files open", nfiles)
			return
		}
		nfiles++
	}

	fmt.Println("files open", nfiles)

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("press a key to contunue...")
	reader.ReadLine()

}
