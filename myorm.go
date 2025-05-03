package myorm
import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"myorm/dialect"
	"myorm/log"
	"myorm/session"
	"strings"
)

type Engine struct {
	db *sql.DB
	dialectSQL dialect.Dialect
	defaultSession *session.Session //一个引擎可以产生多个会话，此处保存一个默认会话
	sessionQueue []*session.Session //引擎产生的多个会话都保存到这个切片里
}
/*
func (engine *Engine)DB() *sql.DB {
	return engine.db
}*/

func NewEngine(driver, source string) (e *Engine, err error) {
	db, err := sql.Open(driver, source)
	if err != nil {
		log.Error(err)
		return
	}
	// Send a ping to make sure the database connection is alive.
	//使用 sql.Open 来连接数据库，但是这个只是返回一个数据库的抽象实例，
	//并没有真正的连接到数据库中，在后续的对数据库的操作中才会真正去网络连接，
	//如果要马上验证，可以用 db.ping().
	if err = db.Ping(); err != nil {
		log.Error(err)
		return
	}
	dial,ok:=dialect.GetDialect(driver)
	if !ok{
		log.Errorf("dialect %s Not Found", driver)
		return
	}
	sessionQueue0:=make([]*session.Session,0)
	e = &Engine{db: db,dialectSQL:dial,sessionQueue:sessionQueue0}
	log.Info("Connect database success")
	return
}

func (engine *Engine) Close() {
	if err := engine.db.Close(); err != nil {
		log.Error("Failed to close database")
	}
	log.Info("Close database success")
}

func (engine *Engine) DefaultSession() *session.Session {
	return engine.defaultSession
}

func (engine *Engine) SessionQueue() []*session.Session {
	return engine.sessionQueue
}

func (engine *Engine) NewSession() *session.Session {
	result:=session.New(engine.db,engine.dialectSQL)
	engine.sessionQueue=append(engine.sessionQueue,result)
	if engine.defaultSession==nil{
		engine.defaultSession=result
	}
	return result
}

type TxFunc func(*session.Session) (interface{}, error)
//用户只需要将所有的操作放到一个回调函数中，作为入参传递给 engine.Transaction()，发生任何错误，
//自动回滚，如果没有错误发生，则提交。
//例子：
//_, err := engine.Transaction(func(s *session.Session) (result interface{}, err error) {
//		_ = s.Model(&User{}).CreateTable()
//		_, err = s.Insert(&User{"Tom", 18})
//		return nil, errors.New("Error")
//	})
func (engine *Engine) Transaction(f TxFunc) (result interface{}, err error) {
	var s *session.Session
	if engine.defaultSession!=nil{
		s = engine.defaultSession
	}else{
		s = engine.NewSession()
	}

	if err := s.Begin(); err != nil { //开启一个事务。s.Begin()表示新建一个事务并将其指针保存到s中。
		return nil, err
	}
	//Recover 是一个Go语言的内建函数，可以让进入宕机流程中的 goroutine 恢复过来，
	//recover 仅在延迟函数 defer 中有效，在正常的执行过程中，
	//调用 recover 会返回 nil 并且没有其他任何效果，如果当前的 goroutine 陷入恐慌，
	//调用 recover 可以捕获到 panic 的输入值，并且恢复正常的执行。
	defer func() {
		if p := recover(); p != nil { // 发生宕机时，获取panic传递的上下文并打印
			_ = s.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			_ = s.Rollback() // err is non-nil; don't change it
		} else {
			err = s.Commit() // err is nil; if Commit returns error update err
		}
	}()

	return f(s)
	//先执行f(s)、后执行defer func()，最后再返回。
}


/*
func (s *session.Session) Transaction(f TxFunc) (result interface{}, err error) {
	if err := s.Begin(); err != nil { //开启一个事务。s.Begin()表示新建一个事务并将其指针保存到s中。
		return nil, err
	}
	//Recover 是一个Go语言的内建函数，可以让进入宕机流程中的 goroutine 恢复过来，
	//recover 仅在延迟函数 defer 中有效，在正常的执行过程中，
	//调用 recover 会返回 nil 并且没有其他任何效果，如果当前的 goroutine 陷入恐慌，
	//调用 recover 可以捕获到 panic 的输入值，并且恢复正常的执行。
	defer func() {
		if p := recover(); p != nil { // 发生宕机时，获取panic传递的上下文并打印
			_ = s.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			_ = s.Rollback() // err is non-nil; don't change it
		} else {
			err = s.Commit() // err is nil; if Commit returns error update err
		}
	}()

	return f(s)
	//先执行f(s)、后执行defer func()，最后再返回。
}
 */

func difference(a []string, b []string) (diff []string) {
	mapB := make(map[string]bool)
	for _, v := range b {
		mapB[v] = true
	}
	for _, v := range a {
		if _, ok := mapB[v]; !ok {
			diff = append(diff, v)
		}
	}
	return
}

// Migrate table
func (engine *Engine) Migrate(value interface{}) error {
	//value是新的结构体，本函数须根据value更新表框架
	_, err := engine.Transaction(func(s *session.Session) (result interface{}, err error) {
		// s.Model(value)表示将根据value建立一个表框架并定为该会话的refTable（更新了refTable）
		// 先更新表框架（s.Model(value)），再更新表。
		if !s.Model(value).HasTable() { // 如果本来就没有表框架，新建一个即可
			log.Infof("table %s doesn't exist", s.RefTable().Name)
			return nil, s.CreateTable()
		}
		// 下面才是更新表的部分。
		table := s.RefTable()
		rows, _ := s.Raw(fmt.Sprintf("SELECT * FROM %s LIMIT 1", table.Name)).QueryRows()
		columns, _ := rows.Columns() //根据查询结果，调用系统SQL库，获得属性名的集合
		addCols := difference(table.FieldNames, columns)
		delCols := difference(columns, table.FieldNames)
		log.Infof("added cols %v, deleted cols %v", addCols, delCols)

		for _, col := range addCols {
			f := table.GetField(col)
			sqlStr := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s;", table.Name, f.Name, f.Type)
			if _, err = s.Raw(sqlStr).Exec(); err != nil {
				return
			}
		}

		if len(delCols) == 0 {
			return
		}
		tmp := "tmp_" + table.Name
		fieldStr := strings.Join(table.FieldNames, ", ")
		//创建一个新表并从旧表中选取需要的属性
		s.Raw(fmt.Sprintf("CREATE TABLE %s AS SELECT %s from %s;", tmp, fieldStr, table.Name))
		//删除旧表
		s.Raw(fmt.Sprintf("DROP TABLE %s;", table.Name))
		//将新表重命名
		s.Raw(fmt.Sprintf("ALTER TABLE %s RENAME TO %s;", tmp, table.Name))
		_, err = s.Exec()
		return
	})
	return err
}