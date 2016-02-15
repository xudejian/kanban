package crawl

import (
	"database/sql"
	"flag"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	"github.com/golang/glog"
)

const (
	categoryTable = "category"
)

var mysql string

func init() {
	flag.StringVar(&mysql, "mysql", "root@/stock", "mysql uri")
	RegisterStore("mysql", &MysqlStore{})
}

func (p *MysqlStore) Open() (err error) {
	if p.db != nil {
		p.Close()
	}

	dsn := mysql
	if !strings.Contains(dsn, "?") {
		dsn = dsn + "?"
	}
	if !strings.Contains(dsn, "parseTime=") {
		dsn = dsn + "&parseTime=true"
	}
	if !strings.Contains(dsn, "loc=") {
		dsn = dsn + "&loc=UTC"
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return
	}
	p.db = db
	return
}

type MysqlStore struct {
	db *sql.DB
}

func (p *MysqlStore) Close() {
	if p.db != nil {
		p.db.Close()
		p.db = nil
	}
}

func (p *MysqlStore) createDayTdataTable(table string) {
	sql := "CREATE TABLE IF NOT EXISTS `" + table + "` (" +
		"`time` DATETIME NOT NULL," +
		"`open` INT(11) NOT NULL DEFAULT 0," +
		"`high` INT(11) NOT NULL DEFAULT 0," +
		"`low` INT(11) NOT NULL DEFAULT 0," +
		"`close` INT(11) NOT NULL DEFAULT 0," +
		"`volume` INT(11) NOT NULL DEFAULT 0," +
		"PRIMARY KEY (`time`)" +
		")"
	_, err := p.db.Exec(sql)
	if err != nil {
		glog.Warningln("create table err", table, err)
	}
}

func (p *MysqlStore) LoadTDatas(table string) (res []Tdata, err error) {
	rows, err := p.db.Query("SELECT `time`,`open`,`high`,`low`,`close`,`volume` FROM `" + table + "` ORDER BY time")
	if err != nil {
		glog.Warningln(err)
		return
	}
	defer rows.Close()
	d := Tdata{}
	for rows.Next() {
		if err := rows.Scan(&d.Time, &d.Open, &d.High, &d.Low, &d.Close, &d.Volume); err != nil {
			glog.Warningln(err)
		}
		res = append(res, d)
	}
	if err := rows.Err(); err != nil {
		glog.Warningln(err)
	}
	return
}

func (p *MysqlStore) SaveTData(table string, data *Tdata) (err error) {
	for i := 0; i < 2; i++ {
		_, err = p.db.Exec("INSERT INTO `"+table+"`(`time`,`open`,`high`,`low`,`close`,`volume`) values(?,?,?,?,?,?)",
			data.Time, data.Open, data.High, data.Low, data.Close, data.Volume)
		if err == nil {
			break
		}
		p.createDayTdataTable(table)
	}
	if err != nil {
		glog.Warningln("insert tdata error", err, *data)
	}
	return
}

func (p *MysqlStore) createTickTable(table string) {
	sql := "CREATE TABLE IF NOT EXISTS `" + table + "` (" +
		"`time` DATETIME(3) NOT NULL," +
		"`price` INT(11) NOT NULL DEFAULT 0," +
		"`change` INT(11) NOT NULL DEFAULT 0," +
		"`volume` INT(11) NOT NULL DEFAULT 0," +
		"`turnover` INT(11) NOT NULL DEFAULT 0," +
		"`type` INT(11) NOT NULL DEFAULT 0," +
		"PRIMARY KEY (`time`)" +
		")"
	_, err := p.db.Exec(sql)
	if err != nil {
		glog.Warningln("create table err", table, err)
	}
}

func (p *MysqlStore) LoadTicks(table string) (res []Tick, err error) {
	d := Tick{}
	rows, err := p.db.Query("SELECT `time`,`price`,`change`,`volume`,`turnover`,`type` FROM `" + table + "` ORDER BY time")
	if err != nil {
		glog.Warningln(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&d.Time, &d.Price, &d.Change, &d.Volume, &d.Turnover, &d.Type); err != nil {
			glog.Warningln(err)
		}
		res = append(res, d)
	}
	if err := rows.Err(); err != nil {
		glog.Warningln(err)
	}
	return
}

func (p *MysqlStore) SaveTick(table string, tick *Tick) (err error) {
	for i := 0; i < 2; i++ {
		_, err = p.db.Exec("INSERT INTO `"+table+"`(`time`,`price`,`change`,`volume`,`turnover`,`type`) values(?,?,?,?,?,?)",
			tick.Time, tick.Price, tick.Change, tick.Volume, tick.Turnover, tick.Type)
		if err == nil {
			break
		}
		p.createTickTable(table)
	}
	if err != nil {
		glog.Warningf("insert tick error %+v %+v", err, *tick)
	}
	return
}

func (p *MysqlStore) createCategorieTable() {
	table := categoryTable
	sql := "CREATE TABLE IF NOT EXISTS `" + table + "` (" +
		"`id` INT(11) NOT NULL AUTO_INCREMENT," +
		"`pid` INT(11) NOT NULL DEFAULT 0," +
		"`factor` INT(11) NOT NULL DEFAULT 0," +
		"`leaf` TINYINT(1) NOT NULL DEFAULT 0," +
		"`name` VARCHAR(128) NOT NULL DEFAULT ''," +
		"PRIMARY KEY (`id`)," +
		"UNIQUE KEY (`pid`, `name`)" +
		")"
	_, err := p.db.Exec(sql)
	if err != nil {
		glog.Warningln("create table err", table, err)
	}
}

func (p *MysqlStore) LoadCategories() (res []CategoryItemInfo, err error) {
	table := categoryTable
	cols := "`id`,`name`,`pid`,`leaf`,`factor`"
	rows, err := p.db.Query("SELECT " + cols + " FROM `" + table + "`")
	if err != nil {
		glog.Warningln(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		d := CategoryItemInfo{}
		if err = rows.Scan(&d.Id, &d.Name, &d.Pid, &d.Leaf, &d.Factor); err != nil {
			glog.Warningln(err)
			continue
		}
		res = append(res, d)
	}
	if err = rows.Err(); err != nil {
		glog.Warningln(err)
	}
	return
}

func (p *MysqlStore) GetOrInsertCategoryItem(info *CategoryItemInfo) (id int, err error) {
	table := categoryTable
	for i := 0; i < 2; i++ {
		err = p.db.QueryRow("SELECT `id` FROM `"+table+"` WHERE pid=? AND name=?",
			info.Pid, info.Name).Scan(&id)

		if err == sql.ErrNoRows {
			_, err = p.db.Exec("INSERT INTO `"+table+"`(`pid`,`name`,`leaf`) values(?,?,?)",
				info.Pid, info.Name, info.Leaf)
			continue
		}
		if err == nil {
			break
		}
	}
	return
}

func (p *MysqlStore) SaveCategoryItemWithPid(c CategoryItem, pid int) (err error) {
	id := 0
	info := CategoryItemInfo{Name: c.Name, Pid: pid, Leaf: false}
	id, err = p.GetOrInsertCategoryItem(&info)
	if err != nil {
		glog.Warningln(err)
		return
	}

	for _, info := range c.Info {
		info.Pid = id
		info.Leaf = true
		p.GetOrInsertCategoryItem(&info)
	}

	p.SaveCategoryWithPid(c.Sub, id)
	return
}

func (p *MysqlStore) SaveCategoryWithPid(c Category, pid int) (err error) {
	if c == nil {
		return
	}
	for _, cate := range c {
		err = p.SaveCategoryItemWithPid(cate, pid)
	}
	return
}

func (p *MysqlStore) SaveCategories(c Category) (err error) {
	if c == nil {
		return
	}

	p.createCategorieTable()
	p.SaveCategoryWithPid(c, 0)
	return
}