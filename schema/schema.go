package schema

import (
	"myorm/dialect"
	"go/ast"
	"reflect"
)

// Field represents a column of database
//Field:字段，对应数据库中的一个属性（一列），包含列名、类型和注解
type Field struct {
	Name string
	Type string
	Tag  string
}

// Schema represents a table of database
//Schema：模式，即数据库的组织和结构。对应数据库的一个表格
//包含程序中的对应模型、表名、各字段信息、全体列名和列名→字段的映射
//Fields包含了所有字段的所有信息，FieldNames和fieldMap是冗余的。
type Schema struct {
	Model      interface{}
	Name       string
	Fields     []*Field
	FieldNames []string
	fieldMap   map[string]*Field
}

//根据名字获得字段
func (schema *Schema) GetField(name string) *Field {
	return schema.fieldMap[name]
}

//传入一个结构体的实例和方言，建立一个与该结构体对应的表框架（Schema）
func Parse(dest interface{}, d dialect.Dialect) *Schema {
	modelType := reflect.Indirect(reflect.ValueOf(dest)).Type()
	schema := &Schema{
		Model:    dest,
		Name:     modelType.Name(),
		fieldMap: make(map[string]*Field),
	}

	for i := 0; i < modelType.NumField(); i++ {
		p := modelType.Field(i)
		if !p.Anonymous && ast.IsExported(p.Name) {
			field := &Field{
				Name: p.Name,
				Type: d.DataTypeOf(reflect.Indirect(reflect.New(p.Type))),
			}
			if v, ok := p.Tag.Lookup("myorm"); ok {
				field.Tag = v
			}
			schema.Fields = append(schema.Fields, field)
			schema.FieldNames = append(schema.FieldNames, p.Name)
			schema.fieldMap[p.Name] = field
		}
	}
	return schema
}

//{"amy",19}转化成["amy",19]
func (schema *Schema) RecordValues(dest interface{}) []interface{} {
	destValue := reflect.Indirect(reflect.ValueOf(dest))
	var fieldValues []interface{}
	for _, field := range schema.Fields {
		fieldValues = append(fieldValues, destValue.FieldByName(field.Name).Interface())
	}
	return fieldValues
}
/*
func Parse(data interface{},d dialect.Dialect) *Schema {
	modelType := reflect.Indirect(reflect.ValueOf(data)).Type()
	var(
		fields=make([]*Field,0)
		fieldNames=make([]string,0)
		fieldmap=make(map[string]*Field)
	)
	for i:=0;i<modelType.NumField();i++{
		p:=modelType.Field(i)
		if !p.Anonymous{
			fieldNames=append(fieldNames,p.Name)
			temp:=&Field{p.Name,d.DataTypeOf(reflect.Indirect(reflect.New(p.Type))),nil}
			if v,ok:=p.Tag.Lookup("myorm");ok{
				temp.Tag=v
			}
			fields=append(fields,temp)
			fieldNames=append(fieldNames,p.Name)
			fieldmap[p.Name]=temp
		}
	}
	ans:=&Schema{Model: data,Name: modelType.Name(),Fields:fields,FieldNames: fieldNames,fieldMap: fieldmap}
	return ans
}*/