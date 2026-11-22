package taskservice

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/BekZhor/Golang_SIS/analyzers"
	"github.com/BekZhor/Golang_SIS/concurrency"
	"github.com/BekZhor/Golang_SIS/taskstruct"
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

const (
	TextStats = iota + 1
	TextReplace
	TextReverse

	DigitStats
	DigitMergeSort
	DigitPeaks
)

type TaskService interface {
	CreateTask()
	RemoveTask()
	ListTasks()
	SearchTask()
}

type textReplaceParams struct {
	OldText string
	NewText string
}

type digitMergeParams struct {
	Second []float64
	Order  string
}

type TextService struct {
	Tasks map[int]taskstruct.TextTask

	mu            sync.RWMutex
	pool          *concurrency.WorkerPool
	replaceParams map[int]textReplaceParams
}

type DigitService struct {
	Tasks map[int]taskstruct.DigitTask

	mu          sync.RWMutex
	pool        *concurrency.WorkerPool
	mergeParams map[int]digitMergeParams
}

func (s *TextService) initWorkerPool() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.pool != nil {
		return nil
	}

	s.replaceParams = make(map[int]textReplaceParams)

	s.pool = concurrency.NewWorkerPool(
		s.processTask,
		s.handleTaskError,
	)

	return nil
}

func (s *TextService) CreateTask() {
	text := readLine("Enter text: ")

	if err := s.ValidationTask(text); err != nil {
		fmt.Println("Error:", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

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

	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.Tasks[id]
	if !exists {
		fmt.Println("Task not found")
		return
	}

	if task.Status == "Queued" || task.Status == "Processing" {
		fmt.Println("Cannot remove a task that is being processed")
		return
	}

	delete(s.Tasks, id)
	fmt.Println("Task removed")
}

func (s *TextService) ListTasks() {
	s.mu.RLock()
	defer s.mu.RUnlock()

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

	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.Tasks[id]
	if !exists {
		fmt.Println("Task not found")
		return
	}

	fmt.Printf("%+v\n", task)
}

func (s *TextService) SubmitAnalyze(
	id int,
	operation int,
	oldText string,
	newText string,
) error {
	if operation < TextStats || operation > TextReverse {
		return fmt.Errorf("invalid text operation")
	}

	if operation == TextReplace && oldText == "" {
		return fmt.Errorf("text to replace cannot be empty")
	}

	if err := s.initWorkerPool(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.Tasks[id]
	if !exists {
		return fmt.Errorf("task %d not found", id)
	}

	if task.Status == "Queued" || task.Status == "Processing" {
		return fmt.Errorf("task is already being processed")
	}

	if err := s.ValidationTask(task.TextInput); err != nil {
		return err
	}

	switch operation {
	case TextStats:
		task.Operation = "Show text statistics"
	case TextReplace:
		task.Operation = "Replace word or symbol"
	case TextReverse:
		task.Operation = "Reverse text"
	}

	task.Status = "Queued"
	s.Tasks[id] = task

	if operation == TextReplace {
		s.replaceParams[id] = textReplaceParams{
			OldText: oldText,
			NewText: newText,
		}
	}

	err := s.pool.Submit(concurrency.QueueItem{
		TaskID:    id,
		Operation: operation,
	})

	if err != nil {
		task.Status = "Created"
		s.Tasks[id] = task
		delete(s.replaceParams, id)
		return err
	}

	return nil
}

func (s *TextService) processTask(item concurrency.QueueItem) error {
	s.mu.Lock()

	task, exists := s.Tasks[item.TaskID]
	if !exists {
		s.mu.Unlock()
		return fmt.Errorf("text task %d not found", item.TaskID)
	}

	task.Status = "Processing"
	s.Tasks[item.TaskID] = task
	fmt.Printf("Task %d: Processing\n", item.TaskID)

	params := s.replaceParams[item.TaskID]

	s.mu.Unlock()

	switch item.Operation {
	case TextStats:
		analyzers.ShowTextStats(task)

	case TextReplace:
		analyzers.ReplaceText(
			&task,
			params.OldText,
			params.NewText,
		)

	case TextReverse:
		analyzers.ReverseText(&task)

	default:
		return fmt.Errorf("unknown text operation: %d", item.Operation)
	}

	s.mu.Lock()
	task.Status = "Completed"
	s.Tasks[item.TaskID] = task
	delete(s.replaceParams, item.TaskID)
	fmt.Printf("Task %d: Completed\n", item.TaskID)
	s.mu.Unlock()

	return nil
}

func (s *TextService) handleTaskError(
	item concurrency.QueueItem,
	err error,
) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.Tasks[item.TaskID]
	if !exists {
		return
	}

	task.Status = "Failed"
	s.Tasks[item.TaskID] = task
	delete(s.replaceParams, item.TaskID)

	fmt.Printf("Text task %d failed: %v\n", item.TaskID, err)
}

func (s *TextService) Shutdown() {
	s.StopWorkerPool()
}

func (s *DigitService) initWorkerPool() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.pool != nil {
		return nil
	}

	s.mergeParams = make(map[int]digitMergeParams)

	s.pool = concurrency.NewWorkerPool(
		s.processTask,
		s.handleTaskError,
	)

	return nil
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

	s.mu.Lock()
	defer s.mu.Unlock()

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

	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.Tasks[id]
	if !exists {
		fmt.Println("Task not found")
		return
	}

	if task.Status == "Queued" || task.Status == "Processing" {
		fmt.Println("Cannot remove a task that is being processed")
		return
	}

	delete(s.Tasks, id)
	fmt.Println("Task removed")
}

func (s *DigitService) ListTasks() {
	s.mu.RLock()
	defer s.mu.RUnlock()

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

	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.Tasks[id]
	if !exists {
		fmt.Println("Task not found")
		return
	}

	fmt.Printf("%+v\n", task)
}

func (s *DigitService) SubmitAnalyze(
	id int,
	operation int,
	second []float64,
	order string,
) error {
	if operation < DigitStats || operation > DigitPeaks {
		return fmt.Errorf("invalid digit operation")
	}

	if operation == DigitMergeSort {
		if len(second) == 0 {
			return fmt.Errorf("second slice cannot be empty")
		}

		if order != "asc" && order != "desc" {
			return fmt.Errorf("order must be asc or desc")
		}

		for _, value := range second {
			if err := s.ValidationTask(value); err != nil {
				return err
			}
		}
	}

	if err := s.initWorkerPool(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.Tasks[id]
	if !exists {
		return fmt.Errorf("task %d not found", id)
	}

	if task.Status == "Queued" || task.Status == "Processing" {
		return fmt.Errorf("task is already being processed")
	}

	if len(task.DigitInput) == 0 {
		return fmt.Errorf("task contains no numbers")
	}

	for _, value := range task.DigitInput {
		if err := s.ValidationTask(value); err != nil {
			return err
		}
	}

	switch operation {
	case DigitStats:
		task.Operation = "Show number statistics"
	case DigitMergeSort:
		task.Operation = "Merge and sort slices"
	case DigitPeaks:
		task.Operation = "Find local peaks"
	}

	task.Status = "Queued"
	s.Tasks[id] = task

	if operation == DigitMergeSort {
		s.mergeParams[id] = digitMergeParams{
			Second: append([]float64(nil), second...),
			Order:  order,
		}
	}

	err := s.pool.Submit(concurrency.QueueItem{
		TaskID:    id,
		Operation: operation,
	})

	if err != nil {
		task.Status = "Created"
		s.Tasks[id] = task
		delete(s.mergeParams, id)
		return err
	}

	return nil
}

func (s *DigitService) processTask(item concurrency.QueueItem) error {
	s.mu.Lock()

	task, exists := s.Tasks[item.TaskID]
	if !exists {
		s.mu.Unlock()
		return fmt.Errorf("digit task %d not found", item.TaskID)
	}

	task.Status = "Processing"
	s.Tasks[item.TaskID] = task
	fmt.Printf("Task %d: Processing\n", item.TaskID)

	params := s.mergeParams[item.TaskID]

	s.mu.Unlock()

	switch item.Operation {
	case DigitStats:
		analyzers.ShowDigitStats(task)

	case DigitMergeSort:
		resultTask, err := analyzers.MergeAndSort(
			task.DigitInput,
			params.Second,
			params.Order,
		)
		if err != nil {
			return err
		}

		s.mu.Lock()

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

		s.mu.Unlock()

	case DigitPeaks:
		analyzers.FindLocalPeaks(&task)

	default:
		return fmt.Errorf("unknown digit operation: %d", item.Operation)
	}

	s.mu.Lock()
	task.Status = "Completed"
	s.Tasks[item.TaskID] = task
	delete(s.mergeParams, item.TaskID)
	fmt.Printf("Task %d: Completed\n", item.TaskID)
	s.mu.Unlock()

	return nil
}

func (s *DigitService) handleTaskError(
	item concurrency.QueueItem,
	err error,
) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.Tasks[item.TaskID]
	if !exists {
		return
	}

	task.Status = "Failed"
	s.Tasks[item.TaskID] = task
	delete(s.mergeParams, item.TaskID)

	fmt.Printf("Digit task %d failed: %v\n", item.TaskID, err)
}

func (s *DigitService) Shutdown() {
	s.StopWorkerPool()
}

func (s *TextService) StartWorkerPool() error {
	if err := s.initWorkerPool(); err != nil {
		return err
	}

	s.mu.RLock()
	pool := s.pool
	s.mu.RUnlock()

	return pool.Start()
}

func (s *TextService) StopWorkerPool() {
	s.mu.RLock()
	pool := s.pool
	s.mu.RUnlock()

	if pool != nil {
		pool.Stop()
	}
}

func (s *TextService) IsWorkerPoolRunning() bool {
	s.mu.RLock()
	pool := s.pool
	s.mu.RUnlock()

	return pool != nil && pool.IsRunning()
}

func (s *TextService) ListQueue() []concurrency.QueueItem {
	s.mu.RLock()
	pool := s.pool
	s.mu.RUnlock()

	if pool == nil {
		return nil
	}

	return pool.ListQueued()
}

func (s *TextService) SearchQueue(id int) (concurrency.QueueItem, bool) {
	s.mu.RLock()
	pool := s.pool
	s.mu.RUnlock()

	if pool == nil {
		return concurrency.QueueItem{}, false
	}

	return pool.GetQueued(id)
}

func (s *TextService) RemoveFromQueue(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.pool == nil {
		return fmt.Errorf("queue is empty")
	}

	if err := s.pool.RemoveQueued(id); err != nil {
		return err
	}

	task, exists := s.Tasks[id]
	if exists {
		task.Status = "Created"
		s.Tasks[id] = task
	}

	delete(s.replaceParams, id)

	return nil
}

func (s *DigitService) StartWorkerPool() error {
	if err := s.initWorkerPool(); err != nil {
		return err
	}

	s.mu.RLock()
	pool := s.pool
	s.mu.RUnlock()

	return pool.Start()
}

func (s *DigitService) StopWorkerPool() {
	s.mu.RLock()
	pool := s.pool
	s.mu.RUnlock()

	if pool != nil {
		pool.Stop()
	}
}

func (s *DigitService) IsWorkerPoolRunning() bool {
	s.mu.RLock()
	pool := s.pool
	s.mu.RUnlock()

	return pool != nil && pool.IsRunning()
}

func (s *DigitService) ListQueue() []concurrency.QueueItem {
	s.mu.RLock()
	pool := s.pool
	s.mu.RUnlock()

	if pool == nil {
		return nil
	}

	return pool.ListQueued()
}

func (s *DigitService) SearchQueue(id int) (concurrency.QueueItem, bool) {
	s.mu.RLock()
	pool := s.pool
	s.mu.RUnlock()

	if pool == nil {
		return concurrency.QueueItem{}, false
	}

	return pool.GetQueued(id)
}

func (s *DigitService) RemoveFromQueue(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.pool == nil {
		return fmt.Errorf("queue is empty")
	}

	if err := s.pool.RemoveQueued(id); err != nil {
		return err
	}

	task, exists := s.Tasks[id]
	if exists {
		task.Status = "Created"
		s.Tasks[id] = task
	}

	delete(s.mergeParams, id)

	return nil
}

var _ TaskService = (*TextService)(nil)
var _ TaskService = (*DigitService)(nil)
