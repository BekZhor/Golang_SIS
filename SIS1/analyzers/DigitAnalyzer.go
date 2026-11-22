package analyzers

import (
	"fmt"
	"github.com/BekZhor/Golang_SIS/taskstruct"
	"sort"
)

func ShowDigitStats(task taskstruct.DigitTask) {
	numbers := task.DigitInput

	fmt.Println("========== DIGIT STATISTICS ==========")
	fmt.Printf("Numbers: %v\n", numbers)
	fmt.Printf("Number of elements: %d\n", len(numbers))

	sum := 0.0
	min := numbers[0]
	max := numbers[0]

	for _, value := range numbers {
		sum += value

		if value < min {
			min = value
		}
		if value > max {
			max = value
		}
	}

	fmt.Printf("Sum: %.2f\n", sum)
	fmt.Printf("Average: %.2f\n", sum/float64(len(numbers)))
	fmt.Printf("Minimum: %.2f\n", min)
	fmt.Printf("Maximum: %.2f\n", max)
	fmt.Println("======================================")
}

type DiditTask interface {
	~[]float64
}

func MergeAndSort(first, second []float64, order string) (taskstruct.DigitTask, error) {
	if order != "asc" && order != "desc" {
		return taskstruct.DigitTask{}, fmt.Errorf("invalid order %q: use asc or desc", order)
	}

	merged := make([]float64, 0, len(first)+len(second))
	merged = append(merged, first...)
	merged = append(merged, second...)

	sort.Float64s(merged)

	if order == "desc" {
		for i, j := 0, len(merged)-1; i < j; i, j = i+1, j-1 {
			merged[i], merged[j] = merged[j], merged[i]
		}
	}

	result := taskstruct.DigitTask{
		DigitInput: merged,
	}

	fmt.Println("Merged and sorted numbers:", merged)

	return result, nil
}

func FindLocalPeaks(task *taskstruct.DigitTask) {
	numbers := task.DigitInput
	var peaks []float64

	for i := 1; i < len(numbers)-1; i++ {
		if numbers[i] > numbers[i-1] && numbers[i] > numbers[i+1] {
			peaks = append(peaks, numbers[i])
		}
	}

	task.Result = peaks

	fmt.Println("Local peaks:", task.Result)
}
