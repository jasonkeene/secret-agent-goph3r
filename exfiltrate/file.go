package exfiltrate

import (
	"regexp"
	"strconv"
)

var filePattern = regexp.MustCompile(`list -- \| +([\w\.]+) +(\d+)KB +(\d+)`)
var bandwidthPattern = regexp.MustCompile(`Remaining Bandwidth: (\d+) *KB`)

type File struct {
	Name   string
	Size   int
	Weight int
	Owner  *Agent
	Score  float64
	Packed bool
}

func MoveFiles(conduit Conduit, agentCount int) {
	files := make([]*File, 0)
	agents := make([]*Agent, 0)
	for i := 0; i < agentCount*2; i++ {
		select {
		case rfiles := <-conduit.fileChan:
			for _, file := range rfiles {
				files = append(files, file)
			}
		case agent := <-conduit.agentChan:
			agents = append(agents, agent)
		}
	}

	defer func() {
		for i := 0; i < agentCount; i++ {
			conduit.doneChan <- true
		}
	}()

	delta, push_files := pack(agents, files)

	for _, fd := range delta {
		fd.From.Conn.WriteLine("/send " + fd.To.GopherName + " " + fd.File.Name)
	}

	for _, pf := range push_files {
		pf.Owner.Conn.WriteLine("/send Glenda " + pf.Name)
	}
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
