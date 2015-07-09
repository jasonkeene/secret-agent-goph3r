package main

import (
	"fmt"
	"log"
	"testing"
)

const listFixture = `
   list -- | Remaining Bandwidth: 44254 KB
   list -- |                       Name   Size Secrecy Value
   list -- |                   641A.ppt  754KB            10
   list -- |     BoundlessInformant.doc 2206KB            87
   list -- |     BoundlessInformant.ppt 2459KB            70
   list -- |             EgoGiraffe.doc 2925KB            85
   list -- |             EgoGiraffe.ppt 2092KB            65
   list -- |                   GCHQ.doc 2892KB            44
   list -- |                   GCHQ.ppt 2195KB            25
   list -- |                  PRISM.doc 3100KB            94
   list -- |                  PRISM.ppt 1583KB            71
   list -- | RadicalPornEnthusiasts.doc 1132KB            19
   list -- | RadicalPornEnthusiasts.ppt 3032KB            91
   list -- |                 SIGINT.doc 2670KB            93
   list -- |                 SIGINT.ppt 1834KB            31
   list -- |              TorStinks.doc 1811KB            57
   list -- |              TorStinks.ppt 2318KB            92`

func TestParseList(t *testing.T) {
	bandwidth, files, _ := parseList(listFixture)
	if len(files) != 15 {
		t.Error("Length of files is wrong, should be 15, was:", len(files))
	}
	if bandwidth != 44254 {
		t.Error("Bandwidth is wrong, should be 44254, was:", bandwidth)
	}
}

func TestPack(t *testing.T) {
	// make agents
	agents := make([]*Agent, 0)
	for i := 0; i < 3; i++ {
		agent := &Agent{
			Name:      fmt.Sprintf("Gopher%d", i+1),
			Bandwidth: (i + 1) * 100,
		}
		agents = append(agents, agent)
		if *debug {
			log.Printf("%s: %d", agent.Name, agent.Bandwidth)
		}
	}

	// make files
	files := make([]*File, 0)
	for i := 0; i < 5; i++ {
		weight := (i + 1) * 10
		size := (5 - i) * 55
		file := &File{
			Name:   fmt.Sprintf("File%d", i+1),
			Size:   size,
			Weight: weight,
			Owner:  agents[0],
			Score:  float64(weight) / float64(size),
			Packed: false,
		}
		files = append(files, file)
		if *debug {
			log.Printf("%s (%d, %d, %f)\n",
				file.Name,
				file.Size,
				file.Weight,
				file.Score,
			)
		}
	}

	delta, push_files := pack(agents, files)
	if len(delta) != 3 {
		t.Error("len(delta) expected: 3, got:", len(delta))
	}
	if len(push_files) != 3 {
		t.Error("len(push_files) expected: 3, got:", len(push_files))
	}
	if delta[0].File.Name != "File5" {
		t.Error("delta error, expected: File5, got:", delta[0].File.Name)
	}
	if delta[0].File.Owner.Name != "Gopher3" {
		t.Error("delta error, expected: Gopher3, got:", delta[0].File.Owner.Name)
	}
	if delta[0].From.Name != "Gopher1" {
		t.Error("delta error, expected: Gopher1, got:", delta[0].From.Name)
	}
	if delta[0].To.Name != "Gopher3" {
		t.Error("delta error, expected: Gopher3, got:", delta[0].To.Name)
	}
}
