package mysql

import (
	"database/sql"
	"math/rand"
	"strings"
	"time"

	. "../../base"

	"github.com/golang/glog"
)

const (
	categoryTable = "category"
)

func (p *Mysql) createCategorieTable() {
	table := categoryTable
	sql := "CREATE TABLE IF NOT EXISTS `" + table + "` (" +
		"`id` INT(11) NOT NULL AUTO_INCREMENT," +
		"`pid` INT(11) NOT NULL DEFAULT 0," +
		"`factor` INT(11) NOT NULL DEFAULT 0," +
		"`leaf` TINYINT(1) NOT NULL DEFAULT 0," +
		"`updateAt` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP," +
		"`name` VARCHAR(128) NOT NULL DEFAULT ''," +
		"PRIMARY KEY (`id`)," +
		"UNIQUE KEY (`pid`, `name`)," +
		"KEY (`updateAt`)" +
		")"
	_, err := p.db.Exec(sql)
	if err != nil {
		glog.Warningln("create table err", table, err)
	}
}

func (p *Mysql) LoadCategories() (res []CategoryItemInfo, err error) {
	table := categoryTable
	cols := "`id`,`name`,`pid`,`leaf`,`factor`"
	var rows *sql.Rows
	rows, err = p.db.Query("SELECT " + cols + " FROM `" + table + "`")
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

func (p *Mysql) GetOrInsertCategoryItem(info *CategoryItemInfo) (id int, err error) {
	table := categoryTable
	for i := 0; i < 3; i++ {
		err = p.db.QueryRow("SELECT `id` FROM `"+table+"` WHERE pid=? AND name=?",
			info.Pid, info.Name).Scan(&id)

		if err == nil {
			break
		}

		if strings.Index(err.Error(), "Error 1146:") > -1 {
			p.createCategorieTable()
			continue
		}
		if err == sql.ErrNoRows {
			_, err = p.db.Exec("INSERT INTO `"+table+"`(`pid`,`name`,`leaf`) values(?,?,?)",
				info.Pid, info.Name, info.Leaf)
			continue
		}
		glog.Warningln(err)
	}
	return
}

func (p *Mysql) SaveCategoryItemWithPid(c CategoryItem, pid int) (err error) {
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

	p.SaveCategories(c.Sub, id)
	return
}

func (p *Mysql) SaveCategories(c Category, pid int) (err error) {
	if c == nil {
		return
	}
	for _, cate := range c {
		err = p.SaveCategoryItemWithPid(cate, pid)
	}
	return
}

func (p *Mysql) SaveCategoryItemInfoFactor(datas []CategoryItemInfo) {
	table := categoryTable
	stmt, err := p.db.Prepare("UPDATE `" + table + "` SET `factor`=? WHERE `id`=?")
	if err != nil {
		glog.Warningln(err)
		return
	}
	defer stmt.Close()
	for i, c := 0, len(datas); i < c; i++ {
		if datas[i].Factor < 1 {
			continue
		}
		_, e := stmt.Exec(datas[i].Factor, datas[i].Id)
		if e != nil {
			msg := e.Error()
			if strings.Index(msg, "Error 1062:") > -1 {
				// duplicate
				continue
			}
			if strings.Index(msg, "Error 1615:") > -1 {
				// Prepared statement needs to be re-prepared
				i--
				continue
			}
		}
	}
}

func (p *Mysql) changeStar(pid int, symbol string, star bool) {
	table := categoryTable
	info := CategoryItemInfo{Name: "star", Pid: pid, Leaf: false}
	starId, _ := p.GetOrInsertCategoryItem(&info)
	if starId < 1 {
		return
	}
	info.Name = "unstar"
	unstarId, _ := p.GetOrInsertCategoryItem(&info)
	if unstarId < 1 {
		return
	}

	id := 0
	err := p.db.QueryRow("SELECT `id`,`pid` FROM `"+table+"` WHERE `pid` in (?,?) AND name=?",
		starId, unstarId, symbol).Scan(&id, &pid)
	if err == sql.ErrNoRows {
		info.Name = symbol
		info.Pid = unstarId
		if star {
			info.Pid = starId
		}
		info.Leaf = true
		p.GetOrInsertCategoryItem(&info)
		return
	}
	expPid := unstarId
	if star {
		expPid = starId
	}
	if pid != expPid {
		p.db.Exec("UPDATE `"+table+"` SET `pid`=? WHERE `id`=?", expPid, id)
	}
}

func (p *Mysql) Star(pid int, symbol string) {
	p.changeStar(pid, symbol, true)
}

func (p *Mysql) UnStar(pid int, symbol string) {
	p.changeStar(pid, symbol, false)
}

func (p *Mysql) IsStar(pid int, symbol string) bool {
	table := categoryTable
	count := 0
	starSql := "SELECT `id` FROM `" + table + "` WHERE pid=? AND name='star'"
	err := p.db.QueryRow("SELECT COUNT(1) FROM `"+table+"` WHERE `pid`=("+starSql+") AND name=?",
		pid, symbol).Scan(&count)
	return err == nil && count > 0
}

func (p *Mysql) Lucky(uid int, symbol string) string {
	table := categoryTable

	pidSql := "SELECT `id` FROM `" + table + "` WHERE `pid`=-1 AND `name`='star'"
	nameSql := "SELECT ? UNION ALL SELECT `name` FROM `" + table + "` WHERE `pid`=(" + pidSql + ")"
	sql := "SELECT `name` FROM `" + table + "` WHERE `factor`>0 AND `leaf`=1 AND `name` NOT IN (" + nameSql + ") ORDER BY `updateAt` LIMIT ?,1"

	name := symbol
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 10; i > 0; i-- {
		offset := rand.Intn(10)
		err := p.db.QueryRow(sql, symbol, offset).Scan(&name)
		if err != nil {
			glog.Warningln(err)
			continue
		}
		if len(name) > 0 {
			break
		}
	}

	return name
}