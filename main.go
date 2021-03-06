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

type IPs struct {
	Ports []string
	IPv6  bool
}

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

	mm := make(map[string]IPs)
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		// if there is more than one : then it's ipv6
		ip := sc.Text()
		tmp := strings.Split(ip, ":")
		length := len(tmp)
		if length == 2 {
			tmpMap := mm[tmp[0]]
			tmpMap.Ports = append(tmpMap.Ports, tmp[1])
			tmpMap.IPv6 = false
			mm[tmp[0]] = tmpMap
		} else {
			port := tmp[length-1]
			ipv6 := strings.Join(tmp[:length-1], ":")
			tmpMap := mm[ipv6]
			tmpMap.Ports = append(tmpMap.Ports, port)
			tmpMap.IPv6 = true
			mm[ipv6] = tmpMap

		}
	}

	threads := make(chan string, 30)

	var wg sync.WaitGroup

	for i := 0; i < cap(threads); i++ {
		go worker(&wg, threads)
	}

	for i, v := range mm {
		ports := strings.Join(v.Ports, ",")
		cmd := ""

		if v.IPv6 {
			cmd = fmt.Sprintf("sudo nmap -6 %s -p %s -oX %s%s.xml --dns-servers `curl -s https://raw.githubusercontent.com/k-sau/resolvers/master/nmap-resolvers.txt` -Pn -sS -w --host-timeout 20m --script-timeout 22m --open -T4 --max-retries 3 -sV", i, ports, *dir, i)
		} else {
			cmd = fmt.Sprintf("sudo nmap %s -p %s -oX %s%s.xml --dns-servers `curl -s https://raw.githubusercontent.com/k-sau/resolvers/master/nmap-resolvers.txt` -Pn -sS --host-timeout 20m --script-timeout 22m --open -T4 --max-retries 3 -sV", i, ports, *dir, i)
		}
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
