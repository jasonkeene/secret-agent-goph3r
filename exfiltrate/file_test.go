package exfiltrate

import (
	"reflect"
	"testing"
)

// TODO: test MoveFiles

func TestParseList(t *testing.T) {
	const listFixture = `
	   list -- | Remaining Bandwidth: 44254 KB
	   list -- |                       Name   Size Secrecy Value
	   list -- |                   641A.ppt  754KB            10
	   list -- |     BoundlessInformant.doc 2206KB            87
	   list -- |             EgoGiraffe.doc 2925KB            85
	   list -- |                   GCHQ.doc 2892KB            44
	   list -- |                  PRISM.doc 3100KB            94
	`

	bandwidth, files, err := parseList(listFixture)
	expected_files := []*File{
		&File{
			Name:   "641A.ppt",
			Size:   754,
			Weight: 10,
			Score:  0.013262599469496022,
			Packed: false,
		},
		&File{
			Name:   "BoundlessInformant.doc",
			Size:   2206,
			Weight: 87,
			Score:  0.03943789664551224,
			Packed: false,
		},
		&File{
			Name:   "EgoGiraffe.doc",
			Size:   2925,
			Weight: 85,
			Score:  0.02905982905982906,
			Packed: false,
		},
		&File{
			Name:   "GCHQ.doc",
			Size:   2892,
			Weight: 44,
			Score:  0.015214384508990318,
			Packed: false,
		},
		&File{
			Name:   "PRISM.doc",
			Size:   3100,
			Weight: 94,
			Score:  0.03032258064516129,
			Packed: false,
		},
	}
	if err != nil {
		t.Error("error occured in parseList")
	}
	if !reflect.DeepEqual(files, expected_files) {
		t.Error("error with expected files")
	}
	if bandwidth != 44254 {
		t.Error("Bandwidth is wrong, should be 44254, was:", bandwidth)
	}
}
