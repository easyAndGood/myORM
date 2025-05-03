package main

import (
	"fmt"
	"myorm"
	"myorm/session"
)

//注意结构体里面，首字母大写才能被SQL识别到
type USER struct {
	Name string `myorm:"PRIMARY KEY"`
	Age  int
	School  string
}

func (u *USER)BeforeInsert(s *session.Session) error { //钩子函数：插入前
	fmt.Println("调用BeforeInsert函数：即将插入",*u,"。")
	return nil
}

func (u *USER)AfterInsert(s *session.Session) error { //钩子函数：插入后
	fmt.Println("调用AfterInsert函数：插入即将完成。")
	return nil
}


func main() {

	// 1、声明并赋值几个实例
	var (
		user1 = &USER{"ou", 18,"CAU"}
		user2 = &USER{"Sam", 25,"GU"}
		user3 = &USER{"Amy", 21,"HU"}
		user4 = &USER{"su", 21,"HU"}
	)

	// 2、新建一个连接数据库的引擎，数据库文件是同一目录下的newDB.db
	en,_:=myorm.NewEngine("sqlite3","newDB.db")

	// 3、新建一个会话，并根据一个空的USER对象创建表框架
	s := en.NewSession().Model(&USER{})

	// 4、如果已经存在USER表，则先删除
	_ = s.DropTable()

	// 5、建立USER表
	_ = s.CreateTable()

	// 6、向USER表插入2个实例
	_, _ = s.Insert(user1, user2)

	// 7、查询一共有几条记录
	num,_ := s.Count()
	fmt.Println("事务前记录个数：",num) //输出："事务前记录个数： 2"

	// 8、新建一个事务（向USER表插入2个实例）
	s,_,_ =s.Transaction(func(s0 *session.Session) (*session.Session,interface{}, error) {
		_, _ = s0.Insert(user3,user4)
		return s0, nil, nil
		//return nil, errors.New("err")
	})
	num,_=s.Count()
	fmt.Println("事务后记录个数：",num)  //输出："事务后记录个数： 4"

	// 9、将Amy的年龄修改为30
	_, _ = s.Where("Name = ?", "Amy").Update("Age", 30)

	// 10、删除学校为"CAU"的学生并输出成功删除记录数
	deleteNum,_:=s.Where("School=?","CAU").Delete()
	fmt.Println("已删除",deleteNum,"条记录")  //输出："事务后记录个数： 4"

	// 11、将名为"Amy"的记录追加到[]USER切片，且最多不超过三条记录（实际上只有一条，因为Name是主键）
	var users []USER
	_ = s.Where("Name = ?", "Amy").Limit(3).Find(&users)
	fmt.Println(users) //输出："[{Amy 30 HU}]"

	// 12、将年纪大于等于18的记录追加到[]USER切片，且最多不超过2条记录（符合条件的其实有三条）
	_ = s.Where("Age>=?",18).Limit(2).Find(&users)
	fmt.Println(users) //输出："[{Amy 30 HU} {Sam 25 GU} {Amy 30 HU}]"

	// 13、查询年纪小于22的有几条记录（链式操作）
	num2,_:=s.Where("Age<?",22).Count()

	// 14、查询年纪大于等于18的有几条记录
	s.Where("Age>=?",18)
	num3,_:=s.Count()

	// 15、查看查询结果
	fmt.Println(num2,num3) //输出："1 3"

}

/*
说明：
直接执行Clear()的函数：Exec()、QueryRows()、QueryRow()
间接执行Clear()的函数：Insert()、Find()、Update()、Delete()、Count()、First()
不会执行Clear()的函数：Limit()、Where()、OrderBy()
执行Clear()后，会话的SQL语句及其参数都会被清空。
链式操作时，须注意使不会执行Clear()的函数在前面，其他在后面。


 */