package exfiltrate

type Conduit struct {
	fileChan   chan []*File
	agentChan  chan *Agent
	doneChan   chan bool
	ResultChan chan string
}

func NewConduit() Conduit {
	return Conduit{
		fileChan:   make(chan []*File),
		agentChan:  make(chan *Agent),
		doneChan:   make(chan bool),
		ResultChan: make(chan string, 1),
	}
}
