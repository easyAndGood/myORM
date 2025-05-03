package session

import (
	"fmt"
	"myorm/log"
	"myorm/schema"
	"reflect"
	"strings"
)

//操作数据库表相关的函数

//传入一个结构体的实例，建立一个表框架并定为 会话的refTable
func (s *Session) Model(value interface{}) *Session {
	// nil or different model, update refTable
	//fmt.Println(reflect.TypeOf(value))
	//fmt.Println(reflect.TypeOf(s.refTable.Model))
	if s.refTable == nil || reflect.TypeOf(value) != reflect.TypeOf(s.refTable.Model) {
		s.refTable = schema.Parse(value, s.dialectSQL)
	}
	return s
}

//获得会话对应的表框架
func (s *Session) RefTable() *schema.Schema {
	if s.refTable == nil {
		log.Error("Model is not set")
	}
	return s.refTable
}

func (s *Session)CreateTable() error {
	table:=s.RefTable()
	col:=make([]string,0)
	for _,value:=range table.Fields{
		//col=append(col,value.Name+" "+value.Type+" "+value.Tag)
		col = append(col, fmt.Sprintf("%s %s %s", value.Name, value.Type, value.Tag))
	}
	s1:=strings.Join(col,",")
	s2:=fmt.Sprintf("CREATE TABLE %s (%s);",table.Name,s1)
	_,err:=s.Raw(s2).Exec()
	return err
}

//HasTable()是根据结构体的名称（string）来判断的
func (s *Session)HasTable() bool {
	d0:=s.dialectSQL
	a1,a2:=d0.TableExistSQL(s.RefTable().Name)
	result:=s.Raw(a1,a2...).QueryRow()
	name1:=""
	_ = result.Scan(&name1)
	return  s.RefTable().Name==name1
}
func (s *Session) DropTable() error {
	_, err := s.Raw(fmt.Sprintf("DROP TABLE IF EXISTS %s", s.RefTable().Name)).Exec()
	return err
}
/*


func (s *Session)Model(model interface{}) *Session {
	if s.refTable==nil || reflect.TypeOf(model)!=reflect.TypeOf(s.refTable.Model){
		s.refTable=schema.Parse(model,s.dialectSQL)
	}
	return s
}
func (s *Session)RefTable() *schema.Schema {
	return s.refTable
}


*/