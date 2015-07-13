package exfiltrate

import (
	"reflect"
	"sort"
	"testing"
)

func TestAgentSort(t *testing.T) {
	agents := []*Agent{
		&Agent{
			Bandwidth: 123,
		},
		&Agent{
			Bandwidth: 55,
		},
		&Agent{
			Bandwidth: 300,
		},
	}
	expected_agents := []*Agent{
		agents[2],
		agents[0],
		agents[1],
	}
	sort.Sort(ByBandwidth(agents))
	if !reflect.DeepEqual(agents, expected_agents) {
		t.Error("sorted agents do not match")
	}
}

func TestFileSort(t *testing.T) {
	files := []*File{
		&File{
			Score: 11.1,
		},
		&File{
			Score: 55.5,
		},
		&File{
			Score: 33.3,
		},
	}
	expected_files := []*File{
		files[1],
		files[2],
		files[0],
	}
	sort.Sort(ByScore(files))
	if !reflect.DeepEqual(files, expected_files) {
		t.Error("sorted files do not match")
	}
}
