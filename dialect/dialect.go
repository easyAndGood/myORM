package dialect
import "reflect"

var dialectsMap = map[string]Dialect{}

//从golang映射到SQL

type Dialect interface {
	DataTypeOf(typ reflect.Value) string //golang类型→SQL类型
	TableExistSQL(tableName string) (string, []interface{})
}

//注册一个方言
func RegisterDialect(name string, dialect Dialect) {
	dialectsMap[name] = dialect
}

//根据数据库软件的名字（如MySQL）获得一个方言
func GetDialect(name string) (dialect Dialect, ok bool) {
	dialect, ok = dialectsMap[name]
	return
}