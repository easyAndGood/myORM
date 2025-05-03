package session

import (
	"errors"
	"myorm/clause"
	"reflect"
)


//直接执行Clear()的函数：Exec()、QueryRows()、QueryRow()
//间接执行Clear()的函数：Insert()、Find()、Update()、Delete()、Count()、First()
//不会执行Clear()的函数：Limit()、Where()、OrderBy()
//执行Clear()后，会话的SQL语句及其参数都会被清空。
//链式操作时，须注意使不会执行Clear()的函数在前面，其他在后面。

//传入多个结构体实例，把每个实例的值改成一条记录并插入数据表中
func (s *Session) Insert(values ...interface{}) (int64, error) {
	recordValues := make([]interface{}, 0)
	for _, value := range values {
		s.CallMethod(BeforeInsert, value)
		table := s.Model(value).RefTable()
		s.clause.Set(clause.INSERT, table.Name, table.FieldNames)
		recordValues = append(recordValues, table.RecordValues(value))
	}
	//recordValues类似于 [[9 "amy"] [92 "john"]]

	//插入语句 pool.Exec("insert into `users` (`name`) values (?)", name)
	s.clause.Set(clause.VALUES, recordValues...)
	sql, vars := s.clause.Build(clause.INSERT, clause.VALUES)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}
	s.CallMethod(AfterInsert, nil)
	return result.RowsAffected()
}
//执行保存到会话中的SQL语句，并将结果保存到传入的结构体切片中。
//使用方法：传入结构体实例的切片的指针，如： var users []User0;s.Find(&users)。结果将追加到users。
func (s *Session) Find(values interface{}) error {
	//fmt.Println("values",values)
	s.CallMethod(BeforeQuery, nil)
	destSlice := reflect.Indirect(reflect.ValueOf(values))
	destType := destSlice.Type().Elem()
	table := s.Model(reflect.New(destType).Elem().Interface()).RefTable()

	s.clause.Set(clause.SELECT, table.Name, table.FieldNames)
	sql, vars := s.clause.Build(clause.SELECT, clause.WHERE, clause.ORDERBY, clause.LIMIT)
	//rows, err := s.Raw(sql, vars...).QueryRows()
	s0:=s.Raw(sql, vars...)
	//s0.QueryRow()
	rows, err := s0.QueryRows()
	if err != nil {
		return err
	}
	//fmt.Println("values",values)
	for rows.Next() {

		dest := reflect.New(destType).Elem()

		var value []interface{}
		//fmt.Println("dest",dest,values)
		for _, name := range table.FieldNames {
			value = append(value, dest.FieldByName(name).Addr().Interface())
		}
		if err := rows.Scan(value...); err != nil {
			return err
		}
		s.CallMethod(AfterQuery, dest.Addr().Interface())

		//是这一句把数据追加到values里面的
		destSlice.Set(reflect.Append(destSlice, dest))
		//fmt.Println("values",values)
	}
	//fmt.Println("values",values)
	return rows.Close()
}

// support map[string]interface{}
// also support kv list: "Name", "Tom", "Age", 18, ....
// Update 方法比较特别的一点在于，Update 接受 2 种入参，平铺开来的键值对和 map 类型的键值对。
// 因为 generator 接受的参数是 map 类型的键值对，因此 Update 方法会动态地判断传入参数的类型，
// 如果是不是 map 类型，则会自动转换。
func (s *Session) Update(kv ...interface{}) (int64, error) {
	s.CallMethod(BeforeUpdate, nil)
	m, ok := kv[0].(map[string]interface{})
	if !ok {
		m = make(map[string]interface{})
		for i := 0; i < len(kv); i += 2 {
			m[kv[i].(string)] = kv[i+1]
		}
	}
	s.clause.Set(clause.UPDATE, s.RefTable().Name, m)
	sql, vars := s.clause.Build(clause.UPDATE, clause.WHERE)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}
	s.CallMethod(AfterUpdate, nil)
	return result.RowsAffected()
}

func (s *Session) Delete() (int64, error) {
	s.CallMethod(BeforeDelete, nil)
	s.clause.Set(clause.DELETE, s.RefTable().Name)
	sql, vars := s.clause.Build(clause.DELETE, clause.WHERE)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}
	s.CallMethod(AfterDelete, nil)
	return result.RowsAffected()
}


// Count records with where clause
func (s *Session) Count() (int64, error) {
	s.clause.Set(clause.COUNT, s.RefTable().Name)
	sql, vars := s.clause.Build(clause.COUNT, clause.WHERE)
	row := s.Raw(sql, vars...).QueryRow()
	var tmp int64
	if err := row.Scan(&tmp); err != nil {
		return 0, err
	}
	return tmp, nil
}

// Limit adds limit condition to clause
func (s *Session) Limit(num int) *Session {
	s.clause.Set(clause.LIMIT, num)
	return s
}

// Where adds limit condition to clause
func (s *Session) Where(desc string, args ...interface{}) *Session {
	//desc是SQL语句，如"Age > ?"
	//args是对应的参数，如args[0]=18
	var vars []interface{}
	s.clause.Set(clause.WHERE, append(append(vars, desc), args...)...)
	return s
}

// OrderBy adds order by condition to clause
func (s *Session) OrderBy(desc string) *Session {
	s.clause.Set(clause.ORDERBY, desc)
	return s
}

//获得第一个记录
//用法： u := &User{}
//_ = s.OrderBy("Age DESC").First(u)
func (s *Session) First(value interface{}) error {
	dest := reflect.Indirect(reflect.ValueOf(value))
	destSlice := reflect.New(reflect.SliceOf(dest.Type())).Elem()
	if err := s.Limit(1).Find(destSlice.Addr().Interface()); err != nil {
		return err
	}
	if destSlice.Len() == 0 {
		return errors.New("NOT FOUND")
	}
	dest.Set(destSlice.Index(0))
	return nil
}

/*
func (s *Session)COUNT1() (int64,error) {
	s.clause.Set(clause.COUNT)
	sql,vars:=s.clause.Build(clause.COUNT,clause.WHERE)
	re,err:=s.Raw(sql,vars...).Exec()
	var ans int64
	if err:=re.Scan(&ans);err!=nil{
		return 0, err
	}
	return ans, err
}

func (s *Session)DELETE() (int64,error) {
	s.clause.Set(clause.DELETE,s.RefTable().Name)
	sql,vars:=s.clause.Build(clause.DELETE,clause.WHERE)
	re,err:=s.Raw(sql,vars).Exec()
	if err!=nil{
		return 0, err
	}
	return re.RowsAffected()
}*/