package taskstruct

import ()

type InputType interface {
	~string | []float64
}

type TextTask struct {
	ID        int
	TextInput string
	Operation string
	Status    string
	Result    string
}

type DigitTask struct {
	ID         int
	DigitInput []float64
	Operation  string
	Status     string
	Result     []float64
}
