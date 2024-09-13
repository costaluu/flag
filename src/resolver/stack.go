package resolver

import (
	"errors"
	"sync"
)

// Using Generics to define Type in Stake to Use Structs, too.
type Stack[T any] struct {
    lock sync.Mutex // Mutex for Thread safety
    S    []T // Slice
}

func NewStack[T any]() *Stack[T] {
    return &Stack[T]{lock: sync.Mutex{}, S: []T{}}
}

func (stack *Stack[T]) Push(element T) {
    stack.lock.Lock()
    defer stack.lock.Unlock()
    stack.S = append(stack.S, element)
}

func (stack *Stack[T]) Pop() (T, error) {
    stack.lock.Lock()
    defer stack.lock.Unlock()
    l := len(stack.S)
    if l == 0 {
        var empty T
        return empty, errors.New("empty Stack")
    }
    element := stack.S[l-1]
    stack.S = stack.S[:l-1]
    return element, nil
}

func (stack *Stack[T]) Peek() (T, error) {
    stack.lock.Lock()
    defer stack.lock.Unlock()
    l := len(stack.S)
    if l == 0 {
        var empty T
        return empty, errors.New("empty Stack")
    }
    return stack.S[l-1], nil
}