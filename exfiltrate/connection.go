package exfiltrate

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

type Connection interface {
	ReadUntilPause() (string, error)
	WriteLine(string)
	DumpLog(string)
}

const pauseDelay = 1

type TCPConnection struct {
	Target    string
	Conn      *net.Conn
	Log       []string
	ReadChan  chan string
	ErrChan   chan error
	BufReader *bufio.Reader
}

func NewTCPConnection(target string) (*TCPConnection, error) {
	connection := &TCPConnection{
		Target:   target,
		Log:      make([]string, 0),
		ReadChan: make(chan string, 3),
		ErrChan:  make(chan error),
	}

	// establish tcp connection
	conn, err := net.Dial("tcp", target)
	if err != nil {
		return connection, err
	}
	connection.Conn = &conn
	connection.BufReader = bufio.NewReader(conn)

	go connection.reader()

	return connection, nil
}

func (c *TCPConnection) reader() {
	for {
		line, err := c.readLine()
		if err != nil {
			c.ErrChan <- err
		}
		c.ReadChan <- line
	}
}

func (c *TCPConnection) readLine() (string, error) {
	data, err := c.BufReader.ReadString('\n')
	if err != nil {
		return "", err
	}
	data = strings.TrimSuffix(data, "\n")
	data = strings.TrimSuffix(data, "\r")
	c.Log = append(c.Log, "<- "+data)
	log.Printf("<- %s", data)

	return data, nil
}

func (c *TCPConnection) ReadUntilPause() (string, error) {
	data := ""

	for {
		select {
		case line := <-c.ReadChan:
			data += line
			data += "\n"
		case <-time.After(time.Second * pauseDelay):
			if data != "" {
				return data, nil
			}
		case err := <-c.ErrChan:
			return data, err
		}
	}
}

func (c *TCPConnection) WriteLine(data string) {
	c.Log = append(c.Log, "-> "+data)
	log.Printf("-> %s", data)
	fmt.Fprintf(*c.Conn, data+"\r\n")
}

func (c *TCPConnection) Teardown() {
	(*c.Conn).Close()
}

func (c *TCPConnection) DumpLog(name string) {
	for _, line := range c.Log {
		fmt.Printf("%s: %s\n", name, line)
	}
}
