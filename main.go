package main

import (
	"bufio"
	"fmt"
	"github.com/BekZhor/Golang_SIS/taskservice"
	"os"
	"strings"
)

/*
func ChooseInputService(text_input *string, digits_input *[]float64) {
	var service_type int

	fmt.Scanln(&service_type)

	switch service_type {
	case 1:
		fmt.Println("Input your text to process")

		reader := bufio.NewReader(os.Stdin)
		text, err := reader.ReadString('\n')

		if err != nil && len(text) == 0 {
			fmt.Println("Error reading text:", err)
			fmt.Println("Please reinput service type")
			ChooseInputService(text_input, digits_input)
			return
		}

		text = strings.TrimSpace(text)
		if len(text) == 0 {
			fmt.Println("Error: text cannot be empty.Please reinput service type")
			ChooseInputService(text_input, digits_input)
			return
		}

		*text_input = text
	case 2:
		fmt.Println("Input your data to analyse")

		reader := bufio.NewReader(os.Stdin)
		text, err := reader.ReadString('\n')

		if err != nil && len(text) == 0 {
			fmt.Println("Error reading data:", err)
			fmt.Println("Please reinput service type")
			ChooseInputService(text_input, digits_input)
			return
		}

		text = strings.TrimSpace(text)
		parts := strings.Fields(text)

		if len(parts) == 0 {
			fmt.Println("Error: no numbers were entered.Please reinput service type")
			ChooseInputService(text_input, digits_input)
			return
		}

		var temp []float64
		for _, part := range parts {
			val, err := strconv.ParseFloat(part, 64)

			if err != nil {
				fmt.Printf("Invalid number %q: %v\nPlease reinput service type\n", part, err)
				ChooseInputService(text_input, digits_input)
				return
			}

			temp = append(temp, val)
		}
		*digits_input = temp
	default:
		fmt.Println("Text Processing - 1\nFinancial Calculation - 2\nPlease choose what service you want")
		ChooseInputService(text_input, digits_input)
		return
	}
}
*/

var reader = bufio.NewReader(os.Stdin)

func readChoice() string {
	fmt.Print("Choose an option: ")
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func main() {
	textService := taskservice.TextService{}
	digitService := taskservice.DigitService{}

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
		fmt.Println("5. Analyze task")
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
			service.ApplyAnalyze()
		case "0":
			return
		default:
			fmt.Println("Invalid choice")
		}
	}
}
