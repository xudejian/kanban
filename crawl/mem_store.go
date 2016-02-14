package crawl

func init() {
	RegisterStore("mem", &MemStore{})
}

type MemStore struct {
}

func (p *MemStore) Open() error { return nil }

func (p *MemStore) Close() {
}

func (p *MemStore) LoadTDatas(table string) (res []Tdata, err error) {
	return
}

func (p *MemStore) SaveTData(table string, data *Tdata) (err error) {
	return
}

func (p *MemStore) LoadTicks(table string) (res []Tick, err error) {
	return
}

func (p *MemStore) SaveTick(table string, tick *Tick) (err error) {
	return
}

func (p *MemStore) LoadCategories() (res TopCategory, err error) { return }
func (p *MemStore) SaveCategories(c TopCategory) (err error)     { return }
