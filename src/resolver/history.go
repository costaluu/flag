package resolver

import (
	"fmt"

	"github.com/costaluu/flag/types"
)

type ConflictRecord struct {
	Current 		types.Conflict
	UndoStack       *Stack[types.Conflict]
	RedoStack       *Stack[types.Conflict]
}

func (stack *ConflictRecord) RecordChange(conflict types.Conflict) {
	if stack.Current.Content != conflict.Content {
		stack.UndoStack.Push(stack.Current)
		stack.Current = conflict
	}
}

func (stack *ConflictRecord) Undo() {
	current := stack.Current
	pop, err := stack.UndoStack.Pop()

	if err == nil {
		stack.RedoStack.Push(current)
		stack.Current = pop
	}
}

func (stack *ConflictRecord) Redo() {
	pop, err := stack.RedoStack.Pop()

	if err == nil {
		stack.UndoStack.Push(pop)
		stack.Current = pop
	}
}

func (stack *ConflictRecord) Show() {
	var conflict types.Conflict
	var err error = nil

	fmt.Printf("CURRENT\n\n")

	fmt.Printf("%+v\n", stack.Current)

	fmt.Printf("UNDO\n\n")

	for err == nil {
		conflict, err = stack.UndoStack.Pop()
		
		fmt.Printf("%+v\n", conflict)
	}

	fmt.Printf("\nREDO\n\n")

	err = nil
	
	for err == nil {
		conflict, err = stack.RedoStack.Pop()
		
		fmt.Printf("%+v\n", conflict)
	}
}