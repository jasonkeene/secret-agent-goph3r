package exfiltrate

type ByBandwidth []*Agent

func (a ByBandwidth) Len() int      { return len(a) }
func (a ByBandwidth) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByBandwidth) Less(i, j int) bool {
	return a[i].Bandwidth > a[j].Bandwidth
}

type ByScore []*File

func (f ByScore) Len() int           { return len(f) }
func (f ByScore) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f ByScore) Less(i, j int) bool { return f[i].Score > f[j].Score }
