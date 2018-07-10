package stack

import "errors"
//声明Stack为空接口类型的切片
type Stack []interface {}

// Length 栈的元素个数
//实现Len()和Cap()方法用来获取其长度和容量
func (stack Stack) Len() int {
	return len(stack)
}

func (stack Stack) Cap() int {
	return cap(stack)
}

// Empty 是否为空栈
func (stack Stack) IsEmpty() bool  {
	return len(stack) == 0
}

//希望更改原始值 使用指针传递
//可以接受任意类型作为参数。尾部添加新值
func (stack *Stack) Push(value interface{})  {
	*stack = append(*stack, value)
}

//弹出一个元素
func (stack *Stack) Pop() (interface{}, error)  {
	theStack := *stack
	if len(theStack) == 0 {
		return nil, errors.New("Out of index, len is 0")
	}
	value := theStack[len(theStack) - 1]
	*stack = theStack[:len(theStack) - 1]
	return value, nil
}

//如果栈空 返回错误
//如果有值 返回最后一个值
func (stack Stack) Top() (interface{}, error)  {
	if len(stack) == 0 {
		return nil, errors.New("Out of index, len is 0")
	}
	return stack[len(stack) - 1], nil
}

