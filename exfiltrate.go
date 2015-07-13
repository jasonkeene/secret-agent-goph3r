package main

import (
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jasonkeene/secret-agent-goph3r/exfiltrate"
)

const targetAgency = "gophercon2015.coreos.com:4001"
const channelName = "bonobo"
const agentCount = 3

var agentNames = []string{"Rob", "Ken", "Robert"}
var debug = flag.Bool("debug", false, "Shows verbose debug output.")

func main() {
	flag.Parse()

	// channels for agents to communicate with MoveFiles
	conduit := exfiltrate.NewConduit()

	// move files for agents
	// this will block until agents are ready with files
	go exfiltrate.MoveFiles(conduit, agentCount)

	// spawn agents
	wg := sync.WaitGroup{}
	wg.Add(agentCount)
	agents := make([]*exfiltrate.Agent, 0)

	for i := 0; i < agentCount; i++ {
		tcpConnection, err := exfiltrate.NewTCPConnection(targetAgency)
		if err != nil {
			log.Fatal(err)
		}
		defer tcpConnection.Teardown()
		agent := &exfiltrate.Agent{
			Name:        agentNames[i],
			GopherName:  fmt.Sprintf("Gopher%d", i+1),
			ChannelName: channelName,
			Wg:          &wg,
			Conn:        tcpConnection,
		}
		agents = append(agents, agent)
		go agent.Exfiltrate(conduit)
	}

	// block until agents are done exfiltrating
	wg.Wait()

	// dump logs
	if *debug {
		for _, agent := range agents {
			agent.DumpLog()
		}
	}

	// wait for result or timeout
	select {
	case result := <-conduit.ResultChan:
		fmt.Println(result)
	case <-time.After(time.Second * 2):
		fmt.Println("unable to determine result")
	}
}
