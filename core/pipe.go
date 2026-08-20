package core

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

type Pipe struct {
	reader 		*os.File
	writer 		*os.File

	output 		 io.Writer

	mutex 		 sync.Mutex
	waitgroup    sync.WaitGroup

	started 	 bool
	closed  	 bool
}

type Pipes struct {
	Stdout *Pipe
	Stderr *Pipe
}

func NewPipe() (*Pipe, error) {
	reader, writer, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("create pipe: %w", err)
	}

	return &Pipe{
		reader: reader,
		writer: writer,
	}, nil
}

func (pipe *Pipe) ReaderFile() *os.File {
	pipe.mutex.Lock()
	defer pipe.mutex.Unlock()

	return pipe.reader
}

func (pipe *Pipe) WriterFile() *os.File {
	pipe.mutex.Lock()
	defer pipe.mutex.Unlock()

	return pipe.writer
}

func (pipe *Pipe) SetOutput(output io.Writer) {
	pipe.mutex.Lock()
	defer pipe.mutex.Unlock()

	pipe.output = output
}

func (pipe *Pipe) Output() io.Writer {
	pipe.mutex.Lock()
	defer pipe.mutex.Unlock()

	return pipe.output
}

func (pipe *Pipe) Start() error {
	pipe.mutex.Lock()

	if pipe.closed {
		pipe.mutex.Unlock()

		return fmt.Errorf("pipe is closed")
	}

	if pipe.started {
		pipe.mutex.Unlock()

		return fmt.Errorf("pipe has already been started")
	}

	if pipe.reader == nil {
		pipe.mutex.Unlock()

		return fmt.Errorf("pipe has no reader")
	}

	if pipe.output == nil {
		pipe.mutex.Unlock()

		return fmt.Errorf("pipe has no output")
	}

	reader := pipe.reader
	output := pipe.output

	pipe.started = true
	pipe.waitgroup.Add(1)
	pipe.mutex.Unlock()

	go func() {
		defer pipe.waitgroup.Done()
		_, _ = io.Copy(output, reader)
	}()

	return nil
}

func (pipe *Pipe) Wait() {
	pipe.waitgroup.Wait()
}

func (pipe *Pipe) Close() error {
	pipe.mutex.Lock()

	if pipe.closed {
		pipe.mutex.Unlock()
		return nil
	}

	pipe.closed = true

	reader := pipe.reader
	writer := pipe.writer

	pipe.reader = nil
	pipe.writer = nil
	pipe.output = nil

	pipe.mutex.Unlock()

	var _errors []error

	if reader != nil {
		if err := reader.Close(); err != nil {
			_errors = append(_errors, err)
		}
	}

	if writer != nil {
		if err := writer.Close(); err != nil {
			_errors = append(_errors, err)
		}
	}

	return errors.Join(_errors...)
}

func NewPipes() (*Pipes, error) {
	stdout, err := NewPipe()
	if err != nil {
		return nil, fmt.Errorf("create stdout pipe: %w", err)
	}

	stderr, err := NewPipe()
	if err != nil {
		_ = stdout.Close()

		return nil, fmt.Errorf("create stderr pipe: %w", err)
	}

	return &Pipes{
		Stdout: stdout,
		Stderr: stderr,
	}, nil
}

func (pipes *Pipes) Start() error {
	if pipes == nil {
		return fmt.Errorf("pipes is nil")
	}

	if pipes.Stdout == nil {
		return fmt.Errorf("stdout pipe is nil")
	}

	if pipes.Stderr == nil {
		return fmt.Errorf("stderr pipe is nil")
	}

	if err := pipes.Stdout.Start(); err != nil {
		return fmt.Errorf("start stdout pipe: %w", err)
	}

	if err := pipes.Stderr.Start(); err != nil {
		_ = pipes.Stdout.Close()

		return fmt.Errorf("start stderr pipe: %w", err)
	}

	return nil
}

func (pipes *Pipes) Wait() {
	if pipes == nil {
		return
	}

	if pipes.Stdout != nil {
		pipes.Stdout.Wait()
	}

	if pipes.Stderr != nil {
		pipes.Stderr.Wait()
	}
}

func (pipes *Pipes) Close() error {
	if pipes == nil {
		return nil
	}

	var _errors []error

	if pipes.Stdout != nil {
		if err := pipes.Stdout.Close(); err != nil {
			_errors = append(_errors, err)
		}
	}

	if pipes.Stderr != nil {
		if err := pipes.Stderr.Close(); err != nil {
			_errors = append(_errors, err)
		}
	}

	return errors.Join(_errors...)
}