package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
)

var hostname string

func sender() {
	// read list of hosts to connect to
	hostlist := make([]string, 0, 500)
	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			break
		}
		if line[0] != '#' {
			hostlist = append(hostlist, strings.Trim(string(line), "\n \t"))
		}
	}
	f.Close()

	time.Sleep(5 * time.Second)

	var min, max, avg time.Duration

	min = 100000
	max = 0

	count := 0

	// do the stress test
	for sl := 10; sl > 0; sl-- {
		for i := 0; i < len(hostlist)*(11-sl); i++ {
			count++
			// random target
			t := hostlist[rand.Int31n(int32(len(hostlist)))]
			t1 := time.Now()
			test(t)
			t2 := time.Now()
			dt := t2.Sub(t1)
			if dt > max {
				max = dt
			}
			if dt < min {
				min = dt
			}
			avg += dt
			fmt.Println(hostname, "with", t, ":", dt)
		}
		// random sleep
		//time.Sleep(time.Duration(rand.Int31n(int32(sl))) * time.Second)
		// non-random sleep
		time.Sleep(time.Duration(sl) * time.Second)
	}

	fmt.Println(">> stats from", hostname, ": min=", min, "max=", max, "avg=", (avg.Seconds()/float64(count))*1000000.0, "µs")

	time.Sleep(10 * time.Second)
	os.Exit(0)
}

func test(target string) {
	var answer string
	conn, err := net.Dial("tcp", target+":4141")
	if err != nil {
		fmt.Println("!!! error on", hostname, "connecting to", target, ":", err)
	} else {
		fmt.Fprintln(conn, "hello")
		fmt.Fscan(conn, answer)
		conn.Close()
	}
}

func handleConnection(conn net.Conn) {
	var data string
	fmt.Scan(conn, data)
	fmt.Fprintln(conn, data)
	_ = conn.RemoteAddr()
	//from := conn.RemoteAddr()
	//fmt.Println("  ", hostname, " received from ", from)
	conn.Close()
}

func main() {

	hostname, _ = os.Hostname()

	go sender()

	ln, err := net.Listen("tcp", ":4141")
	if err != nil {
		panic(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Print("!!! error on", hostname, ":", err)
		}
		go handleConnection(conn)
	}

}
