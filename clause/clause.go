package clause
import "strings"

//子句
//一个Clause就是一条SQL语句的各行的集合

type Clause struct {
	sql     map[Type]string
	sqlVars map[Type][]interface{}
}

type Type int
const (
	INSERT Type = iota
	VALUES
	SELECT
	LIMIT
	WHERE
	ORDERBY
	UPDATE
	DELETE
	COUNT
)

//设置SQL语句的各行分句及其参数
//调用方法：
//Set(clause.COUNT, s.RefTable().Name 表名)
//Set(clause.LIMIT, num 整数)
//Set(clause.ORDERBY, desc 字符串)
//Set(clause.WHERE, "Age > 18")
func (c *Clause) Set(name Type, vars ...interface{}) {
	if c.sql == nil {
		c.sql = make(map[Type]string)
		c.sqlVars = make(map[Type][]interface{})
	}
	sql, vars := generators[name](vars...)
	c.sql[name] = sql
	c.sqlVars[name] = vars
}

//把分句合并起来
func (c *Clause) Build(orders ...Type) (string, []interface{}) {
	var sqls []string
	var vars []interface{}
	for _, order := range orders {
		if sql, ok := c.sql[order]; ok {
			sqls = append(sqls, sql)
			vars = append(vars, c.sqlVars[order]...)
		}
	}
	return strings.Join(sqls, " "), vars
}