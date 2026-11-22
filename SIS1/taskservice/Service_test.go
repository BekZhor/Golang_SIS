package taskservice

import (
	"bufio"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/BekZhor/Golang_SIS/analyzers"
	"github.com/BekZhor/Golang_SIS/taskstruct"
)

func setTestInput(input string) {
	reader = bufio.NewReader(strings.NewReader(input))
}

func TestTextValidation(t *testing.T) {
	service := &TextService{}

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid text", "Hello world", false},
		{"empty text", "", true},
		{"whitespace only", "   \t ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidationTask(tt.input)

			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidationTask(%q): error = %v, wantErr = %v",
					tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestTextCreateTask(t *testing.T) {
	service := &TextService{}

	setTestInput("Hello Go\n")
	service.CreateTask()

	task, exists := service.Tasks[1]
	if !exists {
		t.Fatal("expected task with ID 1 to be created")
	}

	if task.TextInput != "Hello Go" {
		t.Errorf("TextInput = %q, want %q", task.TextInput, "Hello Go")
	}

	if task.Status != "Created" {
		t.Errorf("Status = %q, want %q", task.Status, "Created")
	}
}

func TestTextCreateTaskInvalidInput(t *testing.T) {
	service := &TextService{}

	setTestInput("   \n")
	service.CreateTask()

	if len(service.Tasks) != 0 {
		t.Fatalf("expected no tasks for invalid input, got %d", len(service.Tasks))
	}
}

func TestTextCreateMultipleTasks(t *testing.T) {
	service := &TextService{}

	setTestInput("First task\nSecond task\n")
	service.CreateTask()
	service.CreateTask()

	if len(service.Tasks) != 2 {
		t.Fatalf("got %d tasks, want 2", len(service.Tasks))
	}

	if service.Tasks[1].TextInput != "First task" {
		t.Errorf("task 1 input = %q, want %q",
			service.Tasks[1].TextInput, "First task")
	}

	if service.Tasks[2].TextInput != "Second task" {
		t.Errorf("task 2 input = %q, want %q",
			service.Tasks[2].TextInput, "Second task")
	}
}

func TestTextApplyAnalyzeReverse(t *testing.T) {
	service := &TextService{
		Tasks: map[int]taskstruct.TextTask{
			1: {
				ID:        1,
				TextInput: "hello",
				Status:    "Created",
			},
		},
	}

	setTestInput("1\n3\n")
	service.ApplyAnalyze()

	task, exists := service.Tasks[1]
	if !exists {
		t.Fatal("expected task to remain in the service")
	}

	if task.Result != "olleh" {
		t.Errorf("Result = %q, want %q", task.Result, "olleh")
	}

	if task.Status != "Completed" {
		t.Errorf("Status = %q, want %q", task.Status, "Completed")
	}

	if task.Operation != "Reverse text" {
		t.Errorf("Operation = %q, want %q",
			task.Operation, "Reverse text")
	}
}

func TestTextApplyAnalyzeInvalidOperation(t *testing.T) {
	service := &TextService{
		Tasks: map[int]taskstruct.TextTask{
			1: {
				ID:        1,
				TextInput: "hello",
				Status:    "Created",
			},
		},
	}

	setTestInput("1\n9\n")
	service.ApplyAnalyze()

	task := service.Tasks[1]

	if task.Status != "Invalid" {
		t.Errorf("Status = %q, want %q", task.Status, "Invalid")
	}

	if task.Result != "" {
		t.Errorf("Result = %q, want an empty result", task.Result)
	}
}

func TestDigitValidation(t *testing.T) {
	service := &DigitService{}

	tests := []struct {
		name    string
		value   float64
		wantErr bool
	}{
		{"ordinary number", 12.5, false},
		{"NaN", math.NaN(), true},
		{"positive infinity", math.Inf(1), true},
		{"negative infinity", math.Inf(-1), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidationTask(tt.value)

			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidationTask(%v): error = %v, wantErr = %v",
					tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestDigitCreateTask(t *testing.T) {
	service := &DigitService{}

	setTestInput("3 1 2.5\n")
	service.CreateTask()

	task, exists := service.Tasks[1]
	if !exists {
		t.Fatal("expected digit task with ID 1 to be created")
	}

	want := []float64{3, 1, 2.5}
	if len(task.DigitInput) != len(want) {
		t.Fatalf("DigitInput = %v, want %v", task.DigitInput, want)
	}

	for i := range want {
		if task.DigitInput[i] != want[i] {
			t.Errorf("DigitInput[%d] = %v, want %v",
				i, task.DigitInput[i], want[i])
		}
	}

	if task.Status != "Created" {
		t.Errorf("Status = %q, want %q", task.Status, "Created")
	}
}

func TestDigitCreateTaskInvalidInput(t *testing.T) {
	service := &DigitService{}

	setTestInput("1 abc 3\n")
	service.CreateTask()

	if len(service.Tasks) != 0 {
		t.Fatalf("expected no tasks for invalid input, got %d", len(service.Tasks))
	}
}

func TestMergeAndSort(t *testing.T) {
	tests := []struct {
		name   string
		first  []float64
		second []float64
		order  string
		want   []float64
	}{
		{
			name:   "ascending",
			first:  []float64{3, 1},
			second: []float64{4, 2},
			order:  "asc",
			want:   []float64{1, 2, 3, 4},
		},
		{
			name:   "descending",
			first:  []float64{3, 1},
			second: []float64{4, 2},
			order:  "desc",
			want:   []float64{4, 3, 2, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := analyzers.MergeAndSort(
				tt.first, tt.second, tt.order,
			)
			if err != nil {
				t.Fatalf("MergeAndSort returned error: %v", err)
			}

			if !reflect.DeepEqual(result.DigitInput, tt.want) {
				t.Errorf("result = %v, want %v",
					result.DigitInput, tt.want)
			}
		})
	}
}

func TestMergeAndSortInvalidOrder(t *testing.T) {
	_, err := analyzers.MergeAndSort(
		[]float64{1, 2},
		[]float64{3, 4},
		"sideways",
	)

	if err == nil {
		t.Fatal("expected an error for invalid sorting order")
	}
}

func TestDigitApplyAnalyzeMerge(t *testing.T) {
	service := &DigitService{
		Tasks: map[int]taskstruct.DigitTask{
			1: {
				ID:         1,
				DigitInput: []float64{3, 1},
				Status:     "Created",
			},
		},
	}

	setTestInput("1\n2\n4 2\nasc\n")
	service.ApplyAnalyze()

	original := service.Tasks[1]
	if !reflect.DeepEqual(original.DigitInput, []float64{3, 1}) {
		t.Errorf("original input changed: %v", original.DigitInput)
	}

	if original.Status != "Completed" {
		t.Errorf("original status = %q, want Completed", original.Status)
	}

	var resultTask taskstruct.DigitTask
	found := false

	for id, task := range service.Tasks {
		if id != 1 && reflect.DeepEqual(
			task.DigitInput, []float64{1, 2, 3, 4},
		) {
			resultTask = task
			found = true
			break
		}
	}

	if !found {
		t.Fatal("merged result was not saved as a separate task")
	}

	if resultTask.Status != "Completed" {
		t.Errorf("result status = %q, want Completed", resultTask.Status)
	}

	if resultTask.Operation != "Merge and sort" {
		t.Errorf("result operation = %q, want Merge and sort", resultTask.Operation)
	}
}

func TestDigitApplyAnalyzeInvalidOperation(t *testing.T) {
	service := &DigitService{
		Tasks: map[int]taskstruct.DigitTask{
			1: {
				ID:         1,
				DigitInput: []float64{3, 1},
				Status:     "Created",
			},
		},
	}

	setTestInput("1\n9\n")
	service.ApplyAnalyze()

	task := service.Tasks[1]

	if task.Status != "Invalid" {
		t.Errorf("status = %q, want Invalid", task.Status)
	}
}

func TestDigitTasksAreIndependent(t *testing.T) {
	service := &DigitService{
		Tasks: map[int]taskstruct.DigitTask{
			1: {
				ID:         1,
				DigitInput: []float64{3, 1},
				Status:     "Created",
			},
			2: {
				ID:         2,
				DigitInput: []float64{8, 6},
				Status:     "Created",
			},
		},
	}

	setTestInput("1\n2\n4 2\nasc\n")
	service.ApplyAnalyze()

	first := service.Tasks[1]
	second := service.Tasks[2]

	if !reflect.DeepEqual(first.DigitInput, []float64{3, 1}) {
		t.Errorf("task 1 input changed: %v", first.DigitInput)
	}

	if !reflect.DeepEqual(second.DigitInput, []float64{8, 6}) {
		t.Errorf("task 2 input changed: %v", second.DigitInput)
	}

	if first.Status != "Completed" {
		t.Errorf("task 1 status = %q, want Completed", first.Status)
	}

	if second.Status != "Created" {
		t.Errorf("task 2 status = %q, want Created", second.Status)
	}

	if len(service.Tasks) != 3 {
		t.Errorf("got %d tasks, want 3 after saving merged result",
			len(service.Tasks))
	}
}
