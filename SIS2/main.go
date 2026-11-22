package main

import (
	"bufio"
	"fmt"
	"github.com/BekZhor/Golang_SIS/taskservice"
	"github.com/BekZhor/Golang_SIS/taskstruct"
	"os"
	"strconv"
	"strings"
	"time"
)

var inputReader = bufio.NewReader(os.Stdin)

func readChoice() string {
	fmt.Print("Choose: ")

	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')

	return strings.TrimSpace(choice)
}

/*
func readChoice() string {
	fmt.Print("Choose: ")
	value, _ := inputReader.ReadString('\n')
	return strings.TrimSpace(value)
}
*/

func readInput(prompt string) string {
	fmt.Print(prompt)
	value, _ := inputReader.ReadString('\n')
	return strings.TrimSpace(value)
}

func readTaskID() int {
	id, err := strconv.Atoi(readInput("Enter task ID: "))
	if err != nil {
		return -1
	}
	return id
}

func main() {
	textService := taskservice.TextService{
		Tasks: make(map[int]taskstruct.TextTask),
	}

	digitService := taskservice.DigitService{
		Tasks: make(map[int]taskstruct.DigitTask),
	}

	defer textService.Shutdown()
	defer digitService.Shutdown()

	for {
		fmt.Println("\n========== TASK MANAGEMENT ==========")
		fmt.Println("1. Text tasks")
		fmt.Println("2. Digit tasks")
		fmt.Println("0. Exit")
		fmt.Println("=====================================")

		var service taskservice.TaskService

		switch readChoice() {
		case "1":
			service = &textService
		case "2":
			service = &digitService
		case "0":
			fmt.Println("Application closed")
			return
		default:
			fmt.Println("Invalid choice")
			continue
		}

		serviceMenu(service)
	}
}

func serviceMenu(service taskservice.TaskService) {
	for {
		fmt.Println("\n========== TASK MENU ==========")
		fmt.Println("1. Create task")
		fmt.Println("2. List tasks")
		fmt.Println("3. Search task")
		fmt.Println("4. Remove task")
		fmt.Println("5. Queue management")
		fmt.Println("0. Back")
		fmt.Println("===============================")

		switch readChoice() {
		case "1":
			service.CreateTask()
		case "2":
			service.ListTasks()
		case "3":
			service.SearchTask()
		case "4":
			service.RemoveTask()
		case "5":
			queueMenu(service)
		case "0":
			return
		default:
			fmt.Println("Invalid choice")
		}
	}
}

func queueMenu(service taskservice.TaskService) {
	for {
		time.Sleep(500 * time.Millisecond)
		fmt.Println("\n========== QUEUE MANAGEMENT ==========")
		fmt.Println("1. Add task to queue")
		fmt.Println("2. List queue")
		fmt.Println("3. Search task in queue")
		fmt.Println("4. Remove task from queue")
		fmt.Println("5. Start Worker Pool")
		fmt.Println("6. Stop Worker Pool")
		fmt.Println("0. Back")
		fmt.Println("======================================")

		switch readChoice() {
		case "1":
			submitAnalysis(service)

		case "2":
			listQueue(service)

		case "3":
			searchQueue(service)

		case "4":
			removeFromQueue(service)

		case "5":
			startWorkerPool(service)

		case "6":
			stopWorkerPool(service)

		case "0":
			return

		default:
			fmt.Println("Invalid choice")
		}
	}
}

func submitAnalysis(service taskservice.TaskService) {
	id := readTaskID()

	switch s := service.(type) {
	case *taskservice.TextService:
		submitTextAnalysis(s, id)

	case *taskservice.DigitService:
		submitDigitAnalysis(s, id)

	default:
		fmt.Println("Unknown service")
	}
}

func submitTextAnalysis(
	service *taskservice.TextService,
	id int,
) {
	fmt.Println("\n========== TEXT ANALYSIS ==========")
	fmt.Println("1. Show text statistics")
	fmt.Println("2. Replace word or symbol")
	fmt.Println("3. Reverse text")
	fmt.Println("===================================")

	choice := readChoice()

	var operation int
	var oldText, newText string

	switch choice {
	case "1":
		operation = taskservice.TextStats

	case "2":
		operation = taskservice.TextReplace
		oldText = readInput("Enter word or symbol to replace: ")
		newText = readInput("Enter replacement: ")

	case "3":
		operation = taskservice.TextReverse

	default:
		fmt.Println("Invalid choice")
		return
	}

	if err := service.SubmitAnalyze(
		id,
		operation,
		oldText,
		newText,
	); err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Task %d added to queue. Status: Queued\n", id)
}

func submitDigitAnalysis(
	service *taskservice.DigitService,
	id int,
) {
	fmt.Println("\n========== DIGIT ANALYSIS ==========")
	fmt.Println("1. Show number statistics")
	fmt.Println("2. Merge two slices and sort")
	fmt.Println("3. Find local peaks")
	fmt.Println("====================================")

	choice := readChoice()

	var operation int
	var second []float64
	var order string

	switch choice {
	case "1":
		operation = taskservice.DigitStats

	case "2":
		operation = taskservice.DigitMergeSort

		input := readInput(
			"Enter another slice of numbers separated by spaces: ",
		)

		for _, part := range strings.Fields(input) {
			value, err := strconv.ParseFloat(part, 64)
			if err != nil {
				fmt.Printf("Invalid number: %s\n", part)
				return
			}

			second = append(second, value)
		}

		order = readInput("Sort ascending or descending? (asc/desc): ")

	case "3":
		operation = taskservice.DigitPeaks

	default:
		fmt.Println("Invalid choice")
		return
	}

	if err := service.SubmitAnalyze(
		id,
		operation,
		second,
		order,
	); err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Task %d added to queue. Status: Queued\n", id)
}

func listQueue(service taskservice.TaskService) {
	switch s := service.(type) {
	case *taskservice.TextService:
		items := s.ListQueue()

		if len(items) == 0 {
			fmt.Println("Queue is empty")
			return
		}

		fmt.Println("\nText queue:")
		for _, item := range items {
			fmt.Printf(
				"Task ID: %d | Operation: %d | Status: Queued\n",
				item.TaskID,
				item.Operation,
			)
		}

	case *taskservice.DigitService:
		items := s.ListQueue()

		if len(items) == 0 {
			fmt.Println("Queue is empty")
			return
		}

		fmt.Println("\nDigit queue:")
		for _, item := range items {
			fmt.Printf(
				"Task ID: %d | Operation: %d | Status: Queued\n",
				item.TaskID,
				item.Operation,
			)
		}
	}
}

func searchQueue(service taskservice.TaskService) {
	id := readTaskID()

	switch s := service.(type) {
	case *taskservice.TextService:
		item, exists := s.SearchQueue(id)
		if !exists {
			fmt.Println("Task not found in queue")
			return
		}

		fmt.Printf(
			"Task ID: %d | Operation: %d | Status: Queued\n",
			item.TaskID,
			item.Operation,
		)

	case *taskservice.DigitService:
		item, exists := s.SearchQueue(id)
		if !exists {
			fmt.Println("Task not found in queue")
			return
		}

		fmt.Printf(
			"Task ID: %d | Operation: %d | Status: Queued\n",
			item.TaskID,
			item.Operation,
		)
	}
}

func removeFromQueue(service taskservice.TaskService) {
	id := readTaskID()

	var err error

	switch s := service.(type) {
	case *taskservice.TextService:
		err = s.RemoveFromQueue(id)

	case *taskservice.DigitService:
		err = s.RemoveFromQueue(id)
	}

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Task %d removed from queue\n", id)
}

func startWorkerPool(service taskservice.TaskService) {
	var err error

	switch s := service.(type) {
	case *taskservice.TextService:
		err = s.StartWorkerPool()

	case *taskservice.DigitService:
		err = s.StartWorkerPool()
	}

	if err != nil {
		fmt.Println("Error:", err)
	}
}

func stopWorkerPool(service taskservice.TaskService) {
	switch s := service.(type) {
	case *taskservice.TextService:
		s.StopWorkerPool()

	case *taskservice.DigitService:
		s.StopWorkerPool()
	}
}
