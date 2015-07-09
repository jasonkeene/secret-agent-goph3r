package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// determine send list for each agent
//    rank order documents by secrect_value/size
//    pack top documents into agents by their capacity
// send files between agents to match sendlist
// send files to glenda for each agent
// send done on each channel

const targetAgency = "gophercon2015.coreos.com:4001"
const channelName = "bonobo"
const agentCount = 3
const pauseDelay = 2

var agentNames = []string{"Rob", "Ken", "Robert"}
var debug = flag.Bool("debug", false, "")

type Agent struct {
	Name      string
	Wg        *sync.WaitGroup
	Conn      *net.Conn
	Log       []string
	ReadChan  chan string
	ErrChan   chan error
	BufReader *bufio.Reader
	Bandwidth int
	Files     []File
}

func NewAgent(name string, wg *sync.WaitGroup) (*Agent, error) {
	agent := Agent{
		Name: name,
		Wg:   wg,
	}

	// set up tcp connection
	conn, err := net.Dial("tcp", targetAgency)
	if err != nil {
		return &agent, err
	}
	agent.Conn = &conn

	// make read/err chans
	agent.ReadChan = make(chan string)
	agent.ErrChan = make(chan error)

	// setup bufio reader
	agent.BufReader = bufio.NewReader(conn)

	// spawn off the reader
	go agent.reader()

	return &agent, nil
}

func (a *Agent) reader() {
	for {
		line, err := a.readLine()
		if err != nil {
			a.ErrChan <- err
		}
		a.ReadChan <- line
	}
}

func (a *Agent) readLine() (string, error) {
	data, err := a.BufReader.ReadString('\n')
	if err != nil {
		return "", err
	}
	data = strings.TrimSuffix(data, "\n")
	data = strings.TrimSuffix(data, "\r")
	a.Log = append(a.Log, "<- "+data)

	if *debug {
		log.Printf("%s: read %#v", a.Name, data)
	}
	return data, nil
}

func (a *Agent) readUntilPause() (string, error) {
	data := ""

	for {
		select {
		case line := <-a.ReadChan:
			data += line
			data += "\n"
		case <-time.After(time.Second * pauseDelay):
			if data != "" {
				return data, nil
			}
		case err := <-a.ErrChan:
			return data, err
		}
	}
}

func (a *Agent) writeLine(data string) {
	if *debug {
		log.Printf("%s: writing %#v", a.Name, data)
	}
	a.Log = append(a.Log, "-> "+data)
	fmt.Fprintf(*a.Conn, data+"\r\n")
}

func (a *Agent) DumpLog() {
	for _, line := range a.Log {
		fmt.Printf("%s: %s\n", a.Name, line)
	}
}

func (a *Agent) Exfiltrate() {
	if *debug {
		log.Printf("agent %s is exfiltrating %s", a.Name, targetAgency)
	}
	defer a.Wg.Done()
	var err error
	var data string

	// read channel question
	_, err = a.readUntilPause()
	if err != nil {
		log.Println(err)
		return
	}

	// send channel name
	a.writeLine(channelName)

	// read connection info
	_, err = a.readUntilPause()
	if err != nil {
		log.Println(err)
		return
	}

	// list files
	a.writeLine("/list")
	data, err = a.readUntilPause()
	if err != nil {
		log.Println(err)
		return
	}
	a.Bandwidth, a.Files, err = parseList(data)
	if err != nil {
		log.Println(err)
		return
	}
}

func (a *Agent) Teardown() {
	(*a.Conn).Close()
}

type File struct {
	Name   string
	Size   int
	Weight int
}

var filePattern = regexp.MustCompile(`list -- \| +([\w\.]+) +(\d+)KB +(\d+)`)
var bandwidthPattern = regexp.MustCompile(`Remaining Bandwidth: (\d+) *KB`)

// pull out bandwidth and files from list data
func parseList(data string) (int, []File, error) {
	files := make([]File, 0)

	bandwidth_match := bandwidthPattern.FindStringSubmatch(data)
	bandwidth, err := strconv.Atoi(bandwidth_match[1])
	if err != nil {
		return bandwidth, files, err
	}

	file_match := filePattern.FindAllStringSubmatch(data, -1)
	for _, match := range file_match {
		size, err := strconv.Atoi(match[2])
		if err != nil {
			return 0, files, err
		}
		weight, err := strconv.Atoi(match[3])
		if err != nil {
			return bandwidth, files, err
		}

		file := File{
			Name:   match[1],
			Size:   size,
			Weight: weight,
		}
		files = append(files, file)
	}
	return bandwidth, files, nil
}

func main() {
	flag.Parse()
	agents := make([]*Agent, 0)
	wg := sync.WaitGroup{}
	wg.Add(agentCount)

	// dump logs
	defer func() {
		for _, agent := range agents {
			agent.DumpLog()
		}
	}()

	// spawn agents
	for i := 0; i < agentCount; i++ {
		agent, err := NewAgent(agentNames[i], &wg)
		if err != nil {
			log.Fatal(err)
		}
		defer agent.Teardown()
		agents = append(agents, agent)
		go agent.Exfiltrate()
	}

	// block till agents are done running
	wg.Wait()
}
