package session

import (
	"database/sql"
	"myorm/clause"
	"myorm/dialect"
	"myorm/log"
	"myorm/schema"
	"strings"
)

//当 tx 不为空时，则使用 tx 执行 SQL 语句，否则使用 db 执行 SQL 语句。
type Session struct {
	db *sql.DB //数据库引擎
	dialectSQL dialect.Dialect //SQL软件的方言
	tx       *sql.Tx //事务
	refTable *schema.Schema //表框架
	clause   clause.Clause //分句生成器
	sql strings.Builder //SQL语句
	sqlVars []interface{} //SQL语句的参数
}
//会话里面只有表框架，并没有数据表。一个表框架对应一个数据表。
//会话必须通过调用HasTable()才能知道数据库中有没有其表框架对应的数据表。

// CommonDB is a minimal function set of db
type CommonDB interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
}

//检查一下sql.DB和sql.Tx有没有实现CommonDB
var _ CommonDB = (*sql.DB)(nil)
var _ CommonDB = (*sql.Tx)(nil)

// 如果事务已经被定义，则执行事务。否则返回原数据库指针
func (s *Session) DB() CommonDB {
	if s.tx != nil {
		return s.tx
	}
	return s.db
}

func (s *Session) SQLDB() CommonDB {
	return s.db
}

func New(db0 *sql.DB,d0 dialect.Dialect) *Session {
	return &Session{db: db0,dialectSQL:d0}
}

func (s *Session)Clear()  {
	s.sql.Reset()
	s.sqlVars=nil
	s.clause = clause.Clause{}
}

//将SQL语句及其参数写入会话中
func (s *Session)Raw(sql0 string,values ...interface{}) *Session {
	s.sql.WriteString(sql0)
	s.sql.WriteString(" ")
	s.sqlVars=append(s.sqlVars,values...)
	return s
}

//执行会话中的SQL语句及其参数
func (s *Session)Exec() (sql.Result,error) {
	defer s.Clear()
	log.Info(s.sql.String(),s.sqlVars)
	result,err:=s.DB().Exec(s.sql.String(),s.sqlVars...)
	if err!=nil{
		log.Error(err)
	}
	return result,err
}

//返回多条记录
func (s *Session)QueryRows() (*sql.Rows,error) {
	defer s.Clear()
	log.Info(s.sql.String(),s.sqlVars)
	result,err:= s.DB().Query(s.sql.String(),s.sqlVars...)
	if err!=nil{
		log.Error(err)
	}
	return result,err
}

//返回1条记录
func (s *Session)QueryRow() *sql.Row {
	defer s.Clear()
	log.Info(s.sql.String(),s.sqlVars)
	result:=s.DB().QueryRow(s.sql.String(),s.sqlVars...)
	return result
}

//直接执行Clear()的函数：Exec()、QueryRows()、QueryRow()
//间接执行Clear()的函数：Insert()、Find()、Update()、Delete()、Count()、First()
//不会执行Clear()的函数：Limit()、Where()、OrderBy()


type TxFunc2 func(s *Session) (*Session, interface{}, error)
func (s *Session) Transaction(f TxFunc2) (resultSession *Session, result interface{}, err error) {
	if err := s.Begin(); err != nil { //开启一个事务。s.Begin()表示新建一个事务并将其指针保存到s中。
		return s,nil, err
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