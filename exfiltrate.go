package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const targetAgency = "gophercon2015.coreos.com:4001"
const channelName = "bonobo"
const agentCount = 3
const pauseDelay = 1

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
	Files     []*File
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

func (a *Agent) Exfiltrate(tv TV) {
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

	// message glenda
	a.writeLine("/look")
	_, err = a.readUntilPause()
	if err != nil {
		log.Println(err)
		return
	}
	a.writeLine("/msg Glenda Hello!")
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
	// set owner on files
	for _, file := range a.Files {
		file.Owner = a
	}

	// send files and agents off to moveFiles
	tv.fileChan <- a.Files
	tv.agentChan <- a

	// send done
	<-tv.doneChan
	a.writeLine("/msg Glenda done")
	data, err = a.readUntilPause()
	if err != nil {
		log.Println(err)
		return
	}
	data, err = a.readUntilPause()
	if err != nil {
		log.Println(err)
		return
	}
}

func (a *Agent) Teardown() {
	(*a.Conn).Close()
}

type ByBandwidth []*Agent

func (a ByBandwidth) Len() int      { return len(a) }
func (a ByBandwidth) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByBandwidth) Less(i, j int) bool {
	return a[i].Bandwidth > a[j].Bandwidth
}

type File struct {
	Name   string
	Size   int
	Weight int
	Owner  *Agent
	Score  float64
	Packed bool
}

type ByScore []*File

func (f ByScore) Len() int           { return len(f) }
func (f ByScore) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f ByScore) Less(i, j int) bool { return f[i].Score > f[j].Score }

var filePattern = regexp.MustCompile(`list -- \| +([\w\.]+) +(\d+)KB +(\d+)`)
var bandwidthPattern = regexp.MustCompile(`Remaining Bandwidth: (\d+) *KB`)

type FileDelta struct {
	File *File
	From *Agent
	To   *Agent
}

type TV struct {
	fileChan  chan []*File
	agentChan chan *Agent
	doneChan  chan bool
}

func NewTV() TV {
	tv := TV{}
	tv.fileChan = make(chan []*File)
	tv.agentChan = make(chan *Agent)
	tv.doneChan = make(chan bool)
	return tv
}

func moveFiles(tv TV) {
	// consume []File and []*Agent
	files := make([]*File, 0)
	agents := make([]*Agent, 0)
	for i := 0; i < agentCount*2; i++ {
		select {
		case rfiles := <-tv.fileChan:
			for _, file := range rfiles {
				files = append(files, file)
			}
		case agent := <-tv.agentChan:
			agents = append(agents, agent)
		}
	}

	// done
	defer func() {
		for i := 0; i < agentCount; i++ {
			tv.doneChan <- true
		}
	}()

	// pack
	delta, push_files := pack(agents, files)

	// reassign
	for _, fd := range delta {
		name := ""
		for i, aname := range agentNames {
			if aname == fd.To.Name {
				name = fmt.Sprintf("Gopher%d", i+1)
				break
			}
		}
		fd.From.writeLine("/send " + name + " " + fd.File.Name)
	}

	// push files
	for _, pf := range push_files {
		if *debug {
			log.Printf("pushing %s from %s", pf.Name, pf.Owner.Name)
		}
		pf.Owner.writeLine("/send Glenda " + pf.Name)
	}
}

// compute files to send to Glenda and other users
func pack(agents []*Agent, files []*File) ([]FileDelta, []*File) {
	delta := make([]FileDelta, 0)
	push_files := make([]*File, 0)

	// sort based on score
	sort.Sort(ByScore(files))

	// sort based on bandwidth
	sort.Sort(ByBandwidth(agents))

	for _, agent := range agents {
		if *debug {
			log.Printf("packing for %s", agent.Name)
		}
		remaining_bandwidth := agent.Bandwidth
		for _, file := range files {
			if *debug {
				log.Printf("looking at file %s (%d, %d, %f)",
					file.Name,
					file.Size,
					file.Weight,
					file.Score,
				)
				log.Printf("remaining_bandwidth: %d", remaining_bandwidth)
			}

			if !file.Packed && file.Size <= remaining_bandwidth {
				file.Packed = true
				remaining_bandwidth -= file.Size
				if file.Owner != agent {
					delta = append(delta, FileDelta{
						File: file,
						From: file.Owner,
						To:   agent,
					})
					file.Owner = agent
				}
				if *debug {
					log.Printf("%s (%d, %d, %f) maps to %s (%d)",
						file.Name,
						file.Size,
						file.Weight,
						file.Score,
						file.Owner.Name,
						file.Owner.Bandwidth,
					)
				}
				push_files = append(push_files, file)
			}
		}
	}
	return delta, push_files
}

// pull out bandwidth and files from list data
func parseList(data string) (int, []*File, error) {
	files := make([]*File, 0)

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

		file := &File{
			Name:   match[1],
			Size:   size,
			Weight: weight,
			Score:  float64(weight) / float64(size),
			Packed: false,
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

	// reorganize files for agents
	tv := NewTV()
	go moveFiles(tv)

	// spawn agents
	for i := 0; i < agentCount; i++ {
		agent, err := NewAgent(agentNames[i], &wg)
		if err != nil {
			log.Fatal(err)
		}
		defer agent.Teardown()
		agents = append(agents, agent)
		go agent.Exfiltrate(tv)
	}

	// block till agents are done running
	wg.Wait()
}
