package session
import "myorm/log"

//数据库事务(transaction)是访问并可能操作各种数据项的一个数据库操作序列，
//这些操作要么全部执行,要么全部不执行，是一个不可分割的工作单位。
//事务由事务开始与事务结束之间执行的全部数据库操作组成。
//分三个阶段：开始；读写；提交或回滚。
//事务开始之后不断进行读写操作，但写操作仅仅将数据写入磁盘缓冲区，而非真正写入磁盘内。
//顺利完成所有操作则提交，数据保存到磁盘；否则回滚。
//本框架中事务的实现有两种，分别是Session的method和Engine的的method。
//调用 s.db.Begin() 得到 *sql.Tx 对象，赋值给 s.tx。
func (s *Session) Begin() (err error) {
	log.Info("transaction begin")
	if s.tx, err = s.db.Begin(); err != nil {
		log.Error(err)
		return
	}
	return
}

func (s *Session) Commit() (err error) {
	log.Info("transaction commit")
	if err = s.tx.Commit(); err != nil {
		log.Error(err)
	}
	s.tx=nil //注意Commit()或Rollback()需要将s.tx设置为空
	return
}

func (s *Session) Rollback() (err error) {
	log.Info("transaction rollback")
	if err = s.tx.Rollback(); err != nil {
		log.Error(err)
	}
	s.tx=nil
	return
}

/*
Session.DB()中，如果s.tx非空则返回s.tx。
而Exec()是执行Session.DB()返回的函数。
如果不在Commit()或Rollback()中将s.tx设置为空，则不能调用Session.DB()中的s.db。
 */