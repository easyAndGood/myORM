package session

import (
	"myorm/log"
	"reflect"
)

// Hooks constants
const (
	BeforeQuery  = "BeforeQuery"
	AfterQuery   = "AfterQuery"
	BeforeUpdate = "BeforeUpdate"
	AfterUpdate  = "AfterUpdate"
	BeforeDelete = "BeforeDelete"
	AfterDelete  = "AfterDelete"
	BeforeInsert = "BeforeInsert"
	AfterInsert  = "AfterInsert"
)
/*
钩子的实现
Hook 的意思是钩住，也就是在消息过去之前，先把消息钩住，不让其传递，使用户可以优先处理。
执行这种操作的函数也称为钩子函数。
“先钩住再处理”，执行某操作之前，优先处理一下，再决定后面的执行走向。
本框架可以让用户自定义8个钩子函数，分别在增删改查操作的前后发生。
例子：
type Account struct {
	ID       int `geeorm:"PRIMARY KEY"`
	Password string
}
func (account *Account) BeforeInsert(s *Session) error {
	log.Info("before inert", account)
	account.ID += 1000
	return nil
}
func (account *Account) AfterQuery(s *Session) error {
	log.Info("after query", account)
	account.Password = "******"
	return nil
}
声明一个账号类，类有一个BeforeInsert函数，则在每次插入记录时均会调用该函数。
有一个AfterQuery函数，则在每次查询前均会调用该函数。

说明：
这些钩子函数必须是结构体（如Account）的对应method。
且函数格式必须是：
func (指针名 *类名) 函数名(s *Session) error {
	操作
	return nil
}
其中函数名必须是"BeforeQuery"、"AfterQuery"、"BeforeUpdate"、"AfterUpdate"、"BeforeDelete"、
"AfterDelete"、"BeforeInsert"、"AfterInsert"之一。
接收且仅接收一个参数：对应的会话的指针。
用户可以根据自己的需要，利用获得的对应会话的指针，设置自己需要的操作（当然也可以不使用该会话）。
用户只能改变 account *Account 或 s *Session ，不能返回值。
函数仅返回一个error。

 */

// CallMethod calls the registered hooks
func (s *Session) CallMethod(method string, value interface{}) {
	//传入value仅仅是为了得到value的类型（进而调用对应结构体的method），value的值不会被使用。
	fm := reflect.ValueOf(s.RefTable().Model).MethodByName(method)
	if value != nil {
		fm = reflect.ValueOf(value).MethodByName(method)
	}
	//如果value不为空则fm=value对应的值的方法，否则fm为会话的默认模型的方法。

	//使用reflect的Method，参数必须是reflect.Value而不能是string或int等等原生的类型。
	//不管原生函数的返回值是什么，reflect.Call的返回值就是reflect.Value的切片，可以计算长度。
	param := []reflect.Value{reflect.ValueOf(s)}
	if fm.IsValid() {
		if v := fm.Call(param); len(v) > 0 {
			if err, ok := v[0].Interface().(error); ok {
				log.Error(err)
			}
		}
	}
	return
}