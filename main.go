package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

func main() {
	dir := flag.String("dir", "", "Path to directory to store xml outputs. Default: ~/scans/armada/")
	flag.Parse()
	if *dir == "" {
		dirname, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		*dir = dirname + "/scans/armada/"
	}

	err := os.MkdirAll(*dir, 0755)
	if err != nil {
		log.Println(err)
	}

	mm := make(map[string][]string)
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		// Expecting only ipv4
		// Expected input 127.0.0.1:443
		ip := sc.Text()
		tmp := strings.Split(ip, ":")
		mm[tmp[0]] = append(mm[tmp[0]], tmp[1])
	}

	threads := make(chan string, 30)

	var wg sync.WaitGroup

	for i := 0; i < cap(threads); i++ {
		go worker(&wg, threads)
	}

	for i, v := range mm {
		ports := strings.Join(v, ",")
		cmd := fmt.Sprintf("sudo nmap %s -p %s -oX ~/scans/armada/%s.xml --dns-servers `curl -s https://raw.githubusercontent.com/k-sau/resolvers/master/nmap-resolvers.txt` -Pn -sS --host-timeout 20m --open -T4 --max-retries 3 -sV", i, ports, i)

		wg.Add(1)
		threads <- cmd
	}

	wg.Wait()
	close(threads)
}

func worker(wg *sync.WaitGroup, cmdChan chan string) {
	for cmd := range cmdChan {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		nmap := exec.Command("bash", "-c", cmd)
		nmap.Stdout = &stdout
		nmap.Stderr = &stderr
		err := nmap.Run()
		if stderr.String() != "" {
			fmt.Println(stderr.String())
		}
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		//fmt.Println(stdout.String())
		wg.Done()
	}
}
