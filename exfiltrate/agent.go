package exfiltrate

import (
	"io"
	"log"
	"strings"
	"sync"
)

type Agent struct {
	Name        string
	GopherName  string
	ChannelName string
	Wg          *sync.WaitGroup
	Conn        Connection

	Bandwidth int
	Files     []*File
}

func (a *Agent) Exfiltrate(conduit Conduit) {
	defer a.Wg.Done()

	var err error
	var data string

	// read channel question
	_, err = a.Conn.ReadUntilPause()
	if err != nil {
		log.Println(err)
		return
	}

	// send channel name
	a.Conn.WriteLine(a.ChannelName)

	// read connection info
	_, err = a.Conn.ReadUntilPause()
	if err != nil {
		log.Println(err)
		return
	}

	// message glenda
	a.Conn.WriteLine("/look")
	_, err = a.Conn.ReadUntilPause()
	if err != nil {
		log.Println(err)
		return
	}
	a.Conn.WriteLine("/msg Glenda Hello!")
	_, err = a.Conn.ReadUntilPause()
	if err != nil {
		log.Println(err)
		return
	}

	// list files
	a.Conn.WriteLine("/list")
	data, err = a.Conn.ReadUntilPause()
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

	// send files and agents off to MoveFiles
	conduit.fileChan <- a.Files
	conduit.agentChan <- a

	// read done from moveFiles
	<-conduit.doneChan
	_, err = a.Conn.ReadUntilPause()
	if err != nil {
		log.Println(err)
		return
	}
	a.Conn.WriteLine("/msg Glenda done")

	// read result in
	data, err = a.Conn.ReadUntilPause()
	if err == io.EOF {
		if strings.Contains(data, "End -- |") {
			conduit.ResultChan <- data
		}
	} else if err != nil {
		log.Println(err)
		return
	}
}

func (a *Agent) DumpLog() {
	a.Conn.DumpLog(a.Name)
}
