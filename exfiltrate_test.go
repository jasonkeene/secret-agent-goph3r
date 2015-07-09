package main

import "testing"

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
