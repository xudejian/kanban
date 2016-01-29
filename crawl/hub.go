package crawl

import "github.com/golang/glog"

type hub_parser struct {
	typing_parser
}

// 娇注
// 判断盘整延伸结束是产生3买卖。
// 判断趋势延伸结束是同级别走势回拉中枢--3买卖后扩展或者非标准趋势延伸9段成大中枢

func (p *Tdatas) ParseHub() {
	line := p.Segment.Line
	p.Hub.drop_last_5_data()
	start := 0
	if l := len(p.Hub.Data); l > 0 {
		start = p.Hub.Data[l-1].end
	} else {
		start = findLineDir(line, len(line))
	}
	if start == -1 {
		return
	}
	glog.Infoln(p.tag, "for hub start", start)

	for i, l := start, len(line); i+2 < l; i++ {
		if lhub := len(p.Hub.Data); lhub > 0 {
			hub := &p.Hub.Data[lhub-1]
			if hub.end == i {
				gn := g(line[i+2])
				dn := d(line[i+2])
				zg := hub.High
				zd := hub.Low

				// [dn, gn] # [ZD, ZG]
				if dn > zg || gn < zd {
				} else {
					hub.end = i + 2
					hub.ETime = line[hub.end].ETime
				}
				i++
				continue
			} else if hub.end > i {
				continue
			}
		}
		zg := ZG(line[i], line[i+1], line[i+2])
		zd := ZD(line[i], line[i+1], line[i+2])
		if zg-zd < p.min_hub_height {
			i++
			continue
		}
		hub := line[i]
		hub.High = zg
		hub.Low = zd
		hub.begin = i
		hub.end = i + 2
		hub.ETime = line[i+2].ETime
		p.Hub.Data = append(p.Hub.Data, hub)
		i++
	}
}

func g(a Typing) int { return a.High }

func d(a Typing) int { return a.Low }

// ZG=min(g1, g2)
func ZG(a, b, c Typing) int {
	return minInt(a.High, c.High)
}

// ZD=max(d1, d2)
func ZD(a, b, c Typing) int {
	return maxInt(a.Low, c.Low)
}

// G=min(gn)
func G(line []Typing, t Typing) int {
	l := len(line)
	begin, end := t.begin, t.end
	if begin < 0 || begin >= l {
		glog.Fatalf("segment line begin=%d should in range [0, %d)", begin, l)
	}
	if end < 0 || end >= l {
		glog.Fatalf("segment line end=%d should in range [0, %d)", end, l)
	}

	v := line[begin].High
	for i, end := begin+1, end+1; i < end; i++ {
		if v > line[i].High {
			v = line[i].High
		}
	}
	return v
}

// GG=max(gn)
func GG(line []Typing, t Typing) int {
	l := len(line)
	begin, end := t.begin, t.end
	if begin < 0 || begin >= l {
		glog.Fatalf("segment line begin=%d should in range [0, %d)", begin, l)
	}
	if end < 0 || end >= l {
		glog.Fatalf("segment line end=%d should in range [0, %d)", end, l)
	}

	v := line[begin].High
	for i, end := begin+1, end+1; i < end; i++ {
		if v < line[i].High {
			v = line[i].High
		}
	}
	return v
}

// D=max(dn)
func D(line []Typing, t Typing) int {
	l := len(line)
	begin, end := t.begin, t.end
	if begin < 0 || begin >= l {
		glog.Fatalf("segment line begin=%d should in range [0, %d)", begin, l)
	}
	if end < 0 || end >= l {
		glog.Fatalf("segment line end=%d should in range [0, %d)", end, l)
	}

	v := line[begin].Low
	for i, end := begin+1, end+1; i < end; i++ {
		if v < line[i].Low {
			v = line[i].Low
		}
	}
	return v
}

// DD=min(dn)
func DD(line []Typing, t Typing) int {
	l := len(line)
	begin, end := t.begin, t.end
	if begin < 0 || begin >= l {
		glog.Fatalf("segment line begin=%d should in range [0, %d)", begin, l)
	}
	if end < 0 || end >= l {
		glog.Fatalf("segment line end=%d should in range [0, %d)", end, l)
	}

	v := line[begin].Low
	for i, end := begin+1, end+1; i < end; i++ {
		if v > line[i].Low {
			v = line[i].Low
		}
	}
	return v
}

func hasHighHub(t Typing) bool { return t.end-t.begin > 7 }

// 2:
// GG1 < DD0 Down
// DD1 > GG0 Up
// ZG1 < ZD0 && GG1 >= DD0 New Hub
// ZD1 > ZG0 && DD1 <= GG0 New Hub
func (p *Tdatas) LinkHub(next *Tdatas) {
	hub := p.Hub
	hub.drop_last_5_line()
	segline := p.Segment.Line
	start := 0
	ldata := len(hub.Data)
	line := hub.Line
	var prev *Typing

	DD1, GG1, ZD1, ZG1 := 0, 0, 0, 0
	DD0, GG0, ZD0, ZG0 := DD1, GG1, ZD1, ZG1

	if l := len(line); l > 0 {
		start = line[l-1].end + 1
		prev = &line[l-1]
		if start-1 < ldata {
			h := hub.Data[start-1]
			ZD0, ZG0 = h.Low, h.High
			DD0, GG0 = prev.Low, prev.High
		}
	}

	make_prev_line_end := func(end int) {
		l := len(line)
		if l < 1 {
			return
		}
		prev := &line[l-1]
		if prev.Type == UpTyping && segline[end].High > prev.High {
			prev.High = segline[end].High
		} else if prev.Type == DownTyping && segline[end].Low < prev.Low {
			prev.Low = segline[end].Low
		}
		glog.Infoln("make prev line end", prev, end)
	}

	for i := start; i < ldata; i++ {

		t := hub.Data[i]
		ZD1, ZG1 = t.Low, t.High
		begin, end := t.begin, t.end
		width := begin - end
		t.High = GG(segline, t)
		t.Low = DD(segline, t)
		t.begin = i
		t.end = i
		DD1, GG1 = t.Low, t.High

		if prev == nil {
			line = append(line, t)
		} else if width > 7 {
			if l := len(line); l > 0 {
				make_prev_line_end(begin - 1)
				glog.Infoln("found width>7 hub", i, prev, t)
			}
			line = append(line, t)
			t = line[len(line)-1]
		} else if DD1 > GG0 { // DD1 > GG0 Up
			l := len(line)
			line[l-1].High = GG1
			line[l-1].end = t.end
			line[l-1].ETime = t.ETime
			line[l-1].Type = UpTyping
			glog.Infoln("found up", i, prev, t)
			t = line[l-1]
		} else if GG1 < DD0 { // GG1 < DD0 Down
			l := len(line)
			line[l-1].Low = DD1
			line[l-1].end = t.end
			line[l-1].ETime = t.ETime
			line[l-1].Type = DownTyping
			glog.Infoln("found down", i, prev, t)
			t = line[l-1]
		} else if ZG1 < ZD0 && GG1 >= DD0 { // ZG1 < ZD0 && GG1 >= DD0 New Hub
			make_prev_line_end(begin - 1)
			glog.Infoln("found ZG1 < ZD0 && GG1 >= DD0 New Hub")
			t.Type = DullTyping
			line = append(line, t)
		} else if ZD1 > ZG0 && DD1 <= GG0 { // ZD1 > ZG0 && DD1 <= GG0 New Hub
			make_prev_line_end(begin - 1)
			glog.Infoln("found ZD1 > ZG0 && DD1 <= GG0")
			t.Type = DullTyping
			line = append(line, t)
		} else {
			t.Type = DullTyping
			line = append(line, t)
		}

		prev = &t
		DD0, GG0, ZD0, ZG0 = DD1, GG1, ZD1, ZG1
	}

	hub.Line = line
	glog.Infoln(hub.tag, "hub link len(line)=", len(line))
	if next != nil {
		next.Segment.Line = hub.Line
	}
}
