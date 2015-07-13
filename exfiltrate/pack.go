package exfiltrate

import "sort"

type FileDelta struct {
	File *File
	From *Agent
	To   *Agent
}

// compute files to send to Glenda and other users
func pack(agents []*Agent, files []*File) ([]FileDelta, []*File) {
	delta := make([]FileDelta, 0)
	push_files := make([]*File, 0)

	sort.Sort(ByScore(files))
	sort.Sort(ByBandwidth(agents))

	for _, agent := range agents {
		remaining_bandwidth := agent.Bandwidth
		for _, file := range files {
			if !file.Packed && file.Size <= remaining_bandwidth {
				file.Packed = true
				remaining_bandwidth -= file.Size
				if file.Owner != agent {
					delta = append(delta, FileDelta{
						File: file,
						From: file.Owner,
						To:   agent,
					})
					file.Owner = agent
				}
				push_files = append(push_files, file)
			}
		}
	}
	return delta, push_files
}
