package crawl

import (
	"bytes"
	"errors"
	"log"
	"time"

	"gopkg.in/mgo.v2"
)

var market_begin_day time.Time

func init() {
	market_begin_day, _ = time.Parse("2006-01-02", "2000-01-01")
}

type Stock struct {
	Id        string `json:"id"`
	M1s       M1s    `json:"m1s"`
	M5s       M5s    `json:"m5s"`
	M30s      M30s   `json:"m30s"`
	Days      Days   `json:"days"`
	Weeks     Weeks  `json:"weeks"`
	Months    Months `json:"months"`
	Ticks     Ticks  `json:"-"`
	last_tick Tick
}

func (p *Stock) days_download(t time.Time) (bool, error) {
	body := DownloadDaysFromSina(p.Id, t)
	body = bytes.TrimSpace(body)
	lines := bytes.Split(body, []byte("\n"))
	count := len(lines)
	if count < 1 {
		return false, nil
	}

	day := Tdata{}
	for i := 0; i < count; i++ {
		line := bytes.TrimSpace(lines[i])
		infos := bytes.Split(line, []byte(","))
		if len(infos) != 6 {
			err := errors.New("could not parse line " + string(line))
			return false, err
		}

		day.FromBytes(infos[0], infos[1], infos[2], infos[3], infos[4], infos[5])
		p.Days.Add(day)
	}
	return true, nil
}

func (p *Stock) Days_update(db *mgo.Database) int {
	c := Day_collection(db, p.Id)
	p.Days.Load(c)
	t := p.Days.latest_time()
	l := len(p.Days.Data)
	p.days_download(t)
	count := len(p.Days.Data)
	if count > l {
		for i, j := l, count; i < j; i++ {
			p.Days.Data[i].Save(c)
		}
	}
	p.Days.Delta = count - l
	return count - l
}

func (p *Stock) M5s_update(db *mgo.Database) int {
	c := M5_collection(db, p.Id)
	p.M5s.Load(c)
	l := len(p.M5s.Data)
	p.m5s_download()
	count := len(p.M5s.Data)
	if count > l {
		for i, j := l, count; i < j; i++ {
			p.M5s.Data[i].Save(c)
		}
	}
	p.M5s.Delta = count - l
	return count - l
}

func (p *Stock) M30s_update(db *mgo.Database) int {
	c := M30_collection(db, p.Id)
	p.M30s.Load(c)
	l := len(p.M30s.Data)
	p.m30s_download()
	count := len(p.M30s.Data)
	if count > l {
		for i, j := l, count; i < j; i++ {
			p.M30s.Data[i].Save(c)
		}
	}
	p.M30s.Delta = count - l
	return count - l
}

func (p *Stock) Ticks_update(db *mgo.Database) int {
	c := Tick_collection(db, p.Id)
	p.Ticks.Load(c)
	begin_time := p.Ticks.latest_time()
	l := len(p.Ticks.Data)

	now := time.Now().UTC()
	end_time := now.Truncate(time.Hour * 24)
	if now.Hour() > 10 {
		end_time = end_time.AddDate(0, 0, 1)
	}

	if begin_time.Equal(market_begin_day) {
		begin_time = end_time.AddDate(0, -2, -1)
	}
	begin_time = begin_time.AddDate(0, 0, 1).Truncate(time.Hour * 24)

	for t := begin_time; t.Before(end_time); t = t.AddDate(0, 0, 1) {
		if !IsTradeDay(t) {
			log.Println(t, "skip non trading day")
			continue
		}

		if TickHasInDB(t, c) {
			log.Println(t, "already in db, skip")
			continue
		}

		log.Println("prepare download ticks", t)
		if ok, err := p.ticks_download(t); ok {
			log.Println("download ticks succ", t)
		} else if err != nil {
			log.Println(err)
		}
	}

	count := len(p.Ticks.Data)
	if count > l {
		for i, j := l, count; i < j; i++ {
			p.Ticks.Data[i].Save(c)
		}
	}
	p.Ticks.Delta = count - l
	return count - l
}

func (p *Stock) get_latest_time_from_db(c *mgo.Collection) time.Time {
	d := Day{}
	err := c.Find(nil).Sort("-_id").Limit(1).One(&d)
	if err != nil {
		log.Println("find fail", err)
		return market_begin_day
	}
	return ObjectId2Time(d.Id)
}

func (p *Tdata) parse_mins_from_sina(line []byte) error {
	items := [6]string{"day:", "open:", "high:", "close:", "low:", "volume:"}
	v := [6]string{}
	line = bytes.TrimSpace(line)
	line = bytes.Trim(line, "[{}]")
	infos := bytes.Split(line, []byte(","))
	if len(infos) != 6 {
		return errors.New("could not parse line " + string(line))
	}

	for i, item := range items {
		v[i] = ""
		for _, info := range infos {
			if bytes.HasPrefix(info, []byte(item)) {
				info = bytes.TrimPrefix(info, []byte(item))
				info = bytes.Trim(info, "\"")
				v[i] = string(info)
			}
		}
	}

	p.FromString(v[0], v[1], v[2], v[3], v[4], v[5])
	return nil
}

func (p *Stock) m30s_download() (bool, error) {
	body := DownloadM30sFromSina(p.Id)
	body = bytes.TrimSpace(body)
	lines := bytes.Split(body, []byte("},{"))
	count := len(lines)
	if count < 1 {
		return false, nil
	}

	data := Tdata{}

	for i := 0; i < count; i++ {
		err := data.parse_mins_from_sina(lines[i])
		if err != nil {
			return false, err
		}
		p.M30s.Add(data)
	}

	return true, nil
}

func (p *Stock) m5s_download() (bool, error) {
	body := DownloadM5sFromSina(p.Id)
	body = bytes.TrimSpace(body)
	lines := bytes.Split(body, []byte("},{"))
	count := len(lines)
	if count < 1 {
		return false, nil
	}

	data := Tdata{}

	for i := 0; i < count; i++ {
		err := data.parse_mins_from_sina(lines[i])
		if err != nil {
			return false, err
		}

		p.M5s.Add(data)
	}
	return true, nil
}

var UnknowSinaRes error = errors.New("could not find '成交时间' in head line")

func (p *Stock) ticks_download(t time.Time) (bool, error) {
	body := Tick_download_from_sina(p.Id, t)
	if body == nil {
		return false, nil
	}
	body = bytes.TrimSpace(body)
	lines := bytes.Split(body, []byte("\n"))
	count := len(lines) - 1
	if count < 1 {
		return false, nil
	}
	if bytes.Contains(lines[0], []byte("script")) {
		return false, nil
	}
	if !bytes.Contains(lines[0], []byte("成交时间")) {
		return false, UnknowSinaRes
	}

	ticks := make([]Tick, count)
	for i := count; i > 0; i-- {
		line := bytes.TrimSpace(lines[i])
		infos := bytes.Split(line, []byte("\t"))
		if len(infos) != 6 {
			err := errors.New("could not parse line " + string(line))
			return false, err
		}
		ticks[count-i].FromString(t, infos[0], infos[1], infos[2],
			infos[3], infos[4], infos[5])
	}
	FixTickTime(ticks)
	FixTickId(ticks)

	for _, tick := range ticks {
		p.Ticks.Add(tick)
	}
	return true, nil
}

func (p *Stock) Ticks_today_update() int {
	l := len(p.Ticks.Data)

	now := time.Now().UTC()
	if !IsTradeDay(now) {
		return 0
	}

	nhour := now.Hour()
	if nhour < 1 || nhour > 10 {
		return 0
	}

	p.ticks_get_today()

	count := len(p.Ticks.Data)
	p.Ticks.Delta = count - l
	return p.Ticks.Delta
}

func (p *Stock) ticks_get_today() bool {
	t := time.Now().UTC().Truncate(time.Hour * 24)
	body := Tick_download_today_from_sina(p.Id)
	if body == nil {
		return false
	}
	body = bytes.TrimSpace(body)
	lines := bytes.Split(body, []byte("\n"))
	count := len(lines) - 2
	if count < 1 {
		return false
	}

	ticks := make([]Tick, count)
	j := 0
	nul := []byte("")
	for i := len(lines) - 1; i > 0 && j < count; i-- {
		line := bytes.TrimSpace(lines[i])
		line = bytes.Trim(line, ");")
		infos := bytes.Split(line, []byte("] = new Array("))
		if len(infos) != 2 {
			continue
		}
		line = bytes.Replace(infos[1], []byte(" "), nul, -1)
		line = bytes.Replace(line, []byte("'"), nul, -1)
		infos = bytes.Split(line, []byte(","))
		if len(infos) != 4 {
			continue
		}

		ticks[j].FromString(t, infos[0], infos[2], nul, infos[1], nul, infos[3])
		j++
	}
	FixTickTime(ticks)
	FixTickData(ticks)

	for _, tick := range ticks {
		p.Ticks.Insert(tick)
	}
	return true
}

func (p *Stock) ticks_get_real() bool {
	t := time.Now().UTC().Truncate(time.Hour * 24)

	body := Tick_download_real_from_sina(p.Id)
	if body == nil {
		return false
	}
  //var hq_str_sh600000="������,14.95,14.95,15.27,15.30,14.87,15.24,15.25,78862572,1192962812,36900,15.24,7104,15.23,59604,15.22,40200,15.21,25800,15.20,32200,15.25,7600,15.26,70100,15.27,5600,15.28,194400,15.29,2015-09-24,15:03:01,00";
  str := fmt.Sprintf("var hq_str_%s=", id)
  i := bytes.Index(body, []byte(str))
  if i < 0 {
    log.Println("not found", str)
    return false
  }

  line := body[i+len(str):]
  i = bytes.Index(line, []byte("\";"))
  if i < 0 {
    log.Println("not found \";")
    return false
  }
  line = line[:i]
  infos := bytes.Split(line, []byte(","))
  if len(infos) < 32 {
    log.Println("sina hq api, res format changed")
    return false
  }

	nul := []byte("")
  tick := Tick{}
  tick.FromString(t, infos[31], infos[3], nul, infos[8], infos[9], nul)
  if p.last_tick.Volume == 0 {
    p.last_tick = tick
    return
  }
  if tick.Time.Before(p.last_tick.Time) {
    p.last_tick.Volume = 0
  }
  if tick.Volume != p.last_tick.Volume {
  }

  tick.Volume = tick.Volume / 100
  p.Ticks.Insert(tick)
	return true
}
