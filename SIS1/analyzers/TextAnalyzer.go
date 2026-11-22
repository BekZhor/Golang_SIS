package analyzers

import (
	"fmt"
	"github.com/BekZhor/Golang_SIS/taskstruct"
	"strings"
	"unicode/utf8"
)

func ShowTextStats(task taskstruct.TextTask) {
	fmt.Println("========== TEXT STATISTICS ==========")
	fmt.Printf("Text: %s\n", task.TextInput)
	fmt.Printf("Characters: %d\n", utf8.RuneCountInString(task.TextInput))
	fmt.Printf("Words: %d\n", len(strings.Fields(task.TextInput)))
	fmt.Printf("Bytes: %d\n", len(task.TextInput))
	fmt.Println("=====================================")
}

func ReplaceText(task *taskstruct.TextTask, old string, new string) {
	task.Result = strings.ReplaceAll(task.TextInput, old, new)
	fmt.Println("Result:", task.Result)
}

func ReverseText(task *taskstruct.TextTask) {
	runes := []rune(task.TextInput)

	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	task.Result = string(runes)
	fmt.Println("Result:", task.Result)
}
