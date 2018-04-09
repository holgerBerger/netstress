package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var hostname string

// SLURM_JOB_NODELIST=hsw[017-031,033-037,069-078,080-088,109]

func sender() {
	// read list of hosts to connect to
	hostlist := make([]string, 0, 500)
	if len(os.Args) > 1 {
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
	} else if len(os.Getenv("SLURM_JOB_NODELIST")) > 0 {
		env := os.Getenv("SLURM_JOB_NODELIST")
		prefix := env[:strings.Index(env, "[")]
		for _, ranges := range strings.Split(env[strings.Index(env, "[")+1:len(env)-1], ",") {
			if strings.Index(ranges, "-") >= 0 {
				flds := strings.Split(ranges, "-")
				start, err1 := strconv.Atoi(flds[0])
				end, err2 := strconv.Atoi(flds[1])
				if err1 == nil && err2 == nil {
					for i := start; i <= end; i++ {
						hostlist = append(hostlist, fmt.Sprintf("%s%0*d", prefix, len(flds[0]), i))
					}
				}
			} else {
				hostlist = append(hostlist, prefix+ranges)
			}
		}
	} else {
		fmt.Println("need nodelist as file or SLURM env")
		os.Exit(-1)
	}

	time.Sleep(5 * time.Second)

	rand.Seed(time.Now().UnixNano())

	var min, max, avg time.Duration

	min = 100000000
	max = 0

	count := 0

	// do the stress test
	for sl := 10; sl > 0; sl-- {
		for i := 0; i < len(hostlist)*(11-sl); i++ {
			// random target
			t := hostlist[rand.Int31n(int32(len(hostlist)))]
			if t != hostname {
				count++
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
				// fmt.Println(hostname, "with", t, ":", dt)
			}
		}
		// random sleep
		//time.Sleep(time.Duration(rand.Int31n(int32(sl))) * time.Second)
		// non-random sleep
		time.Sleep(time.Duration(sl) * time.Second)
	}

	fmt.Println(">> stats from", hostname, ": min=", min, "max=", max, "avg=", (avg.Seconds()/float64(count))*1000000.0, "µs", "count=", count)

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
