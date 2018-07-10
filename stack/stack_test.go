package stack

import (
	"testing"
)

//测试存值 len长度
func TestStack_Len(t *testing.T) {
	var myStack Stack
	myStack.Push(1)
	myStack.Push("test")
	if myStack.Len() == 2 {
		t.Log("长度校验通过")
	}
}
//测试非空
func TestStack_IsEmpty(t *testing.T) {
	var mStack Stack
	if mStack.IsEmpty() {
		t.Log("栈空")
	}
}
//测试容量
func TestStack_Cap(t *testing.T) {
	myStack := make(Stack, 3)
	if myStack.Cap() == 3 {
		t.Log("通过 容量为3")
	}
}

//测试 取值
func TestStack_Top(t *testing.T) {
	var mStack Stack
	if _, err := mStack.Top(); err == nil {
		t.Error("返回失败")
	}
	mStack.Push(3)
	if value, _ := mStack.Top(); value == 3 {
		t.Log("获得值成功")
	} else {
		t.Errorf("获得值失败 value is %d", value)
	}
}