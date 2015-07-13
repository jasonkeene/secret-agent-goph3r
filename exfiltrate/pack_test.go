package exfiltrate

import (
	"reflect"
	"testing"
)

func TestPack(t *testing.T) {
	// make agents
	agents := []*Agent{
		&Agent{
			Name:      "Gopher1",
			Bandwidth: 100,
		},
		&Agent{
			Name:      "Gopher2",
			Bandwidth: 200,
		},
		&Agent{
			Name:      "Gopher3",
			Bandwidth: 300,
		},
	}

	// make files
	files := []*File{
		&File{
			Name:   "File1",
			Size:   275,
			Weight: 10,
			Owner:  agents[0],
			Score:  27.5,
			Packed: false,
		},
		&File{
			Name:   "File2",
			Size:   220,
			Weight: 20,
			Owner:  agents[0],
			Score:  11.0,
			Packed: false,
		},
		&File{
			Name:   "File3",
			Size:   165,
			Weight: 30,
			Owner:  agents[0],
			Score:  5.5,
			Packed: false,
		},
		&File{
			Name:   "File4",
			Size:   110,
			Weight: 40,
			Owner:  agents[0],
			Score:  2.75,
			Packed: false,
		},
		&File{
			Name:   "File5",
			Size:   55,
			Weight: 50,
			Owner:  agents[0],
			Score:  1.1,
			Packed: false,
		},
	}

	expected_delta := []FileDelta{
		FileDelta{
			File: files[0],
			From: agents[0],
			To:   agents[2],
		},
		FileDelta{
			File: files[2],
			From: agents[0],
			To:   agents[1],
		},
	}
	expected_push_files := []*File{
		files[0],
		files[2],
		files[4],
	}
	delta, push_files := pack(agents, files)

	if !reflect.DeepEqual(delta, expected_delta) {
		t.Error("Packing delta is wrong.")
	}
	if !reflect.DeepEqual(push_files, expected_push_files) {
		t.Error("Packing push files is wrong.")
	}
}
