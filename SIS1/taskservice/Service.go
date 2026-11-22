package taskservice

import (
	"bufio"
	"fmt"
	"github.com/BekZhor/Golang_SIS/analyzers"
	"github.com/BekZhor/Golang_SIS/taskstruct"
	"math"
	"os"
	"strconv"
	"strings"
)

var reader = bufio.NewReader(os.Stdin)

func readLine(prompt string) string {
	fmt.Print(prompt)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

func readID() int {
	id, err := strconv.Atoi(readLine("Enter task ID: "))
	if err != nil {
		return -1
	}
	return id
}

type TaskService interface {
	CreateTask()
	RemoveTask()
	ListTasks()
	SearchTask()
	ApplyAnalyze()
}

type TextService struct {
	Tasks map[int]taskstruct.TextTask
}

type DigitService struct {
	Tasks map[int]taskstruct.DigitTask
}

func (s *TextService) CreateTask() {
	text := readLine("Enter text: ")

	if err := s.ValidationTask(text); err != nil {
		fmt.Println("Error:", err)
		return
	}

	if s.Tasks == nil {
		s.Tasks = make(map[int]taskstruct.TextTask)
	}

	id := 1
	for existingID := range s.Tasks {
		if existingID >= id {
			id = existingID + 1
		}
	}

	s.Tasks[id] = taskstruct.TextTask{
		ID:        id,
		TextInput: text,
		Status:    "Created",
	}

	fmt.Printf("Text task created. ID: %d\n", id)
}

func (s *TextService) ValidationTask(text string) error {
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("text cannot be empty")
	}

	return nil
}

func (s *TextService) RemoveTask() {
	id := readID()

	if _, exists := s.Tasks[id]; !exists {
		fmt.Println("Task not found")
		return
	}

	delete(s.Tasks, id)
	fmt.Println("Task removed")
}

func (s *TextService) ListTasks() {
	if len(s.Tasks) == 0 {
		fmt.Println("No text tasks")
		return
	}

	for _, task := range s.Tasks {
		fmt.Printf(
			"ID: %d | Input: %q | Operation: %s | Status: %s | Result: %s\n",
			task.ID, task.TextInput, task.Operation, task.Status, task.Result,
		)
	}
}

func (s *TextService) SearchTask() {
	id := readID()

	task, exists := s.Tasks[id]
	if !exists {
		fmt.Println("Task not found")
		return
	}

	fmt.Printf("%+v\n", task)
}

func (s *TextService) ApplyAnalyze() {
	if len(s.Tasks) == 0 {
		fmt.Println("No text tasks")
		return
	}

	fmt.Println("Choose ID of task you want to analyze:")
	id := readID()

	task, exists := s.Tasks[id]
	if !exists {
		fmt.Println("Task not found")
		return
	}

	fmt.Println("Select text analysis:")
	fmt.Println("1. Show text statistics")
	fmt.Println("2. Replace word or symbol")
	fmt.Println("3. Reverse text")

	choice := readLine("Choose 1-3: ")

	if err := s.ValidationTask(task.TextInput); err != nil {
		task.Status = "Invalid"
		s.Tasks[id] = task
		fmt.Println("Validation failed:", err)
		return
	}

	task.Status = "Ready for analysis"
	s.Tasks[id] = task
	fmt.Println("Status:", task.Status)

	task.Status = "Processing"
	s.Tasks[id] = task
	fmt.Println("Status:", task.Status)

	switch choice {
	case "1":
		task.Operation = "Show text statistics"
		analyzers.ShowTextStats(task)
	case "2":
		task.Operation = "Replace word or symbol"

		oldText := readLine("Enter word or symbol to replace: ")
		newText := readLine("Enter replacement: ")

		if oldText == "" {
			task.Status = "Invalid"
			s.Tasks[id] = task
			fmt.Println("Text to replace cannot be empty")
			return
		}

		analyzers.ReplaceText(&task, oldText, newText)

	case "3":
		task.Operation = "Reverse text"
		analyzers.ReverseText(&task)

	default:
		task.Status = "Invalid"
		s.Tasks[id] = task
		fmt.Println("Invalid choice")
		return
	}

	task.Status = "Completed"
	s.Tasks[id] = task
	fmt.Println("Status:", task.Status)

	if choice != "1" {
		fmt.Println("Analysis result:", task.Result)
	}
}

func (s *DigitService) CreateTask() {
	input := readLine("Enter numbers separated by spaces: ")
	parts := strings.Fields(input)

	if len(parts) == 0 {
		fmt.Println("Error: no numbers provided")
		return
	}

	var numbers []float64

	for _, part := range parts {
		value, err := strconv.ParseFloat(part, 64)
		if err != nil {
			fmt.Printf("Invalid number %q. Task was not created.\n", part)
			return
		}

		if err := s.ValidationTask(value); err != nil {
			fmt.Printf("Invalid number %q: %v\n", part, err)
			return
		}

		numbers = append(numbers, value)
	}

	if s.Tasks == nil {
		s.Tasks = make(map[int]taskstruct.DigitTask)
	}

	id := 1
	for existingID := range s.Tasks {
		if existingID >= id {
			id = existingID + 1
		}
	}

	s.Tasks[id] = taskstruct.DigitTask{
		ID:         id,
		DigitInput: numbers,
		Status:     "Created",
	}

	fmt.Printf("Digit task created. ID: %d\n", id)
}

func (s *DigitService) ValidationTask(value float64) error {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return fmt.Errorf("number must be finite")
	}

	return nil
}

func (s *DigitService) RemoveTask() {
	id := readID()

	if _, exists := s.Tasks[id]; !exists {
		fmt.Println("Task not found")
		return
	}

	delete(s.Tasks, id)
	fmt.Println("Task removed")
}

func (s *DigitService) ListTasks() {
	if len(s.Tasks) == 0 {
		fmt.Println("No digit tasks")
		return
	}

	for _, task := range s.Tasks {
		fmt.Printf(
			"ID: %d | Input: %v | Operation: %s | Status: %s | Result: %v\n",
			task.ID, task.DigitInput, task.Operation, task.Status, task.Result,
		)
	}
}

func (s *DigitService) SearchTask() {
	id := readID()

	task, exists := s.Tasks[id]
	if !exists {
		fmt.Println("Task not found")
		return
	}

	fmt.Printf("%+v\n", task)
}

func (s *DigitService) ApplyAnalyze() {
	if len(s.Tasks) == 0 {
		fmt.Println("No digit tasks")
		return
	}

	fmt.Println("Choose ID of task you want to analyze:")
	id := readID()

	task, exists := s.Tasks[id]
	if !exists {
		fmt.Println("Task not found")
		return
	}

	if len(task.DigitInput) == 0 {
		task.Status = "Invalid"
		s.Tasks[id] = task
		fmt.Println("Validation failed: no numbers provided")
		return
	}

	for _, value := range task.DigitInput {
		if err := s.ValidationTask(value); err != nil {
			task.Status = "Invalid"
			s.Tasks[id] = task
			fmt.Println("Validation failed:", err)
			return
		}
	}

	fmt.Println("Select digit analysis:")
	fmt.Println("1. Show number statistics")
	fmt.Println("2. Merge two slices and sort")
	fmt.Println("3. Find local peaks")

	choice := readLine("Choose 1-3: ")

	switch choice {
	case "1":
		task.Operation = "Show number statistics"
	case "2":
		task.Operation = "Merge and sort slices"
	case "3":
		task.Operation = "Find local peaks"
	default:
		task.Status = "Invalid"
		s.Tasks[id] = task
		fmt.Println("Invalid choice")
		return
	}

	task.Status = "Ready for analysis"
	s.Tasks[id] = task
	fmt.Println("Status:", task.Status)

	task.Status = "Processing"
	s.Tasks[id] = task
	fmt.Println("Status:", task.Status)

	switch choice {
	case "1":
		analyzers.ShowDigitStats(task)
	case "2":
		input := readLine("Enter another slice of numbers separated by spaces: ")
		parts := strings.Fields(input)

		var second []float64

		if len(parts) == 0 {
			task.Status = "Invalid"
			s.Tasks[id] = task
			fmt.Println("Error: no numbers provided")
			return
		}

		for _, part := range parts {
			value, err := strconv.ParseFloat(part, 64)
			if err != nil {
				task.Status = "Invalid"
				s.Tasks[id] = task
				fmt.Printf("Invalid number %q\n", part)
				return
			}

			if err := s.ValidationTask(value); err != nil {
				task.Status = "Invalid"
				s.Tasks[id] = task
				fmt.Println("Validation failed:", err)
				return
			}

			second = append(second, value)
		}

		order := readLine("Sort ascending or descending? (asc/desc): ")

		resultTask, err := analyzers.MergeAndSort(task.DigitInput, second, order)
		if err != nil {
			task.Status = "Invalid"
			s.Tasks[id] = task
			fmt.Println("Error:", err)
			return
		}

		newID := 1
		for existingID := range s.Tasks {
			if existingID >= newID {
				newID = existingID + 1
			}
		}

		resultTask.ID = newID
		resultTask.Operation = "Merge and sort"
		resultTask.Status = "Completed"

		s.Tasks[newID] = resultTask

		fmt.Printf("Merged result saved as task ID %d: %v\n", newID, resultTask.DigitInput)
	case "3":
		analyzers.FindLocalPeaks(&task)
	}

	task.Status = "Completed"
	s.Tasks[id] = task
	fmt.Println("Status:", task.Status)
}

var _ TaskService = (*TextService)(nil)
var _ TaskService = (*DigitService)(nil)
