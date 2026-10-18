package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
	//"errors"
)

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

func main() {
	fmt.Println("Text Processing - 1\nFinancial Calculation - 2")

	var text_input string
	var digits_input []float64

	ChooseInputService(&text_input, &digits_input)


	if text_input != "" {
		fmt.Println("\n========== TEXT ANALYSIS ==========")
		fmt.Printf("Value type: %T\n", text_input)
		fmt.Printf("Original text: %s\n", text_input)
		fmt.Printf("Character count: %d\n", utf8.RuneCountInString(text_input))
		fmt.Printf("Word count: %d\n", len(strings.Fields(text_input)))
		fmt.Printf("Byte size: %d\n", len(text_input))
		fmt.Println("====================================")
	}

	if len(digits_input) > 0 {
		fmt.Println("\n======= FINANCIAL DATA ANALYSIS =======")
		fmt.Printf("Variable type: %T\n", digits_input)
		fmt.Printf("Element type: %T\n", digits_input[0])
		fmt.Printf("Input values: %v\n", digits_input)
		fmt.Printf("Number of elements: %d\n", len(digits_input))
		fmt.Printf("Slice capacity: %d\n", cap(digits_input))

		var sum float64
		min := digits_input[0]
		max := digits_input[0]

		for i, value := range digits_input {
			fmt.Printf("Element [%d]: value = %v, type = %T\n", i, value, value)

			sum += value

			if value < min {
				min = value
			}
			if value > max {
				max = value
			}
		}

		fmt.Printf("Sum: %.2f\n", sum)
		fmt.Printf("Average: %.2f\n", sum/float64(len(digits_input)))
		fmt.Printf("Minimum: %.2f\n", min)
		fmt.Printf("Maximum: %.2f\n", max)
		fmt.Println("=======================================")
	}

}
