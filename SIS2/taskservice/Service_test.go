package taskservice

import (
	"bufio"
	"reflect"
	"strings"
	"testing"

	"github.com/BekZhor/Golang_SIS/analyzers"
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
