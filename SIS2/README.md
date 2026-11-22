# Golang SIS 2

## Overview

SIS 2 is a Go application that extends the original SIS project with concurrent task processing.

The application supports text and digit processing tasks. Instead of executing every operation immediately, tasks are submitted to bounded queues and processed asynchronously by a fixed number of workers.

## Features

- **Task Management** — create, retrieve, list, and remove tasks.
- **Input Validation** — validate task data before submission.
- **Text Processing** — perform text analysis and transformations.
- **Digit Processing** — perform operations on numeric data.
- **Asynchronous Execution** — submit tasks without waiting for their completion.
- **Worker Pools** — process multiple tasks concurrently using reusable workers.
- **Bounded Queues** — limit the number of waiting tasks.
- **Synchronization** — protect shared task data and results from concurrent access.
- **Error Handling** — handle invalid submissions, duplicate tasks, and full queues.
- **Graceful Shutdown** — stop workers and wait for active processing to finish.

## Architecture

The project separates task management, domain operations, and concurrent execution.

```text
Golang_SIS
│
├── analyzers/
│   ├── text processing
│   └── digit processing
│
├── taskservice/
│   ├── TextService
│   └── DigitService
│
├── taskstruct/
│   └── task definitions
│
└── concurrency/
    ├── TaskQueue
    ├── Worker
    └── WorkerPool
```

### Main Components

**Analyzers**

Contain the actual processing operations. They are responsible for performing domain-specific work rather than managing task execution.

**Task Services**

`TextService` and `DigitService` manage their respective tasks, including validation, task state, and results.

Each service has its own worker pool and queue.

**Concurrency Package**

Provides reusable concurrency components:

- `TaskQueue` — stores pending tasks in a buffered channel.
- `Worker` — retrieves tasks from the queue and executes the assigned handler.
- `WorkerPool` — manages a fixed number of workers and their lifecycle.

## Task Processing Workflow

A task follows this general processing lifecycle:

```text
Created
   |
   v
Validated
   |
   v
Queued
   |
   v
Processing
   |
   v
Completed / Failed
```

1. A task is created and validated.
2. The corresponding service submits it to its queue.
3. An available worker retrieves the task.
4. The worker executes the corresponding analyzer operation.
5. The service updates the task state and stores the result.

## Concurrency

The application uses Go concurrency primitives:

| Primitive | Purpose |
|---|---|
| Goroutines | Run workers concurrently |
| Channels | Transfer tasks between the queue and workers |
| `sync.Mutex` | Protect shared data |
| `sync.WaitGroup` | Wait for workers to finish |
| `select` | Handle channel communication and shutdown signals |

The worker pool uses a fixed number of workers instead of creating a new goroutine for every submitted task.

## Requirements

- Go installed on your system.
- A Go module initialized for the project.

## Running the Project

Clone the repository:

```bash
git clone https://github.com/BekZhor/Golang_SIS.git
cd Golang_SIS
```

Build the project:

```bash
go build ./...
```

Run tests:

```bash
go test ./...
```

Run tests with the race detector:

```bash
go test -race ./...
```

Run static analysis:

```bash
go vet ./...
```

## Testing

The project tests cover:

- Task creation and validation.
- Queue insertion and removal.
- Duplicate task handling.
- Queue capacity limits.
- Concurrent task processing.
- Worker pool startup and shutdown.
- Error handling.
- Concurrent access to shared data.

The race detector can be used to identify data races:

```bash
go test -race ./...
```

## Technologies

- **Go**
- Goroutines and Channels
- `sync` package
- Go Testing Package

## Project Goal

The goal of SIS 2 is to extend the original task management system with asynchronous and concurrent execution while maintaining separation of concerns, controlled resource usage, and safe access to shared data.