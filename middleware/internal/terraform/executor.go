package terraform

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

// LogCallback is a function that receives log lines from terraform execution
type LogCallback func(line string)

// Executor handles running Terraform commands
type Executor struct {
	workingDir  string
	logCallback LogCallback
	logMutex    sync.Mutex
}

// NewExecutor creates a new Terraform executor
func NewExecutor(workingDir string) *Executor {
	return &Executor{
		workingDir: workingDir,
	}
}

// SetLogCallback sets a callback function to receive log lines
func (e *Executor) SetLogCallback(callback LogCallback) {
	e.logMutex.Lock()
	defer e.logMutex.Unlock()
	e.logCallback = callback
}

// log sends a log line to the callback if set
func (e *Executor) log(line string) {
	e.logMutex.Lock()
	callback := e.logCallback
	e.logMutex.Unlock()

	if callback != nil {
		callback(line)
	}
}

// streamOutput reads from a reader and sends lines to both a buffer and the log callback
func (e *Executor) streamOutput(reader io.Reader, buffer *bytes.Buffer) error {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		buffer.WriteString(line)
		buffer.WriteString("\n")
		e.log(line)
	}
	return scanner.Err()
}

// Init initializes Terraform in the working directory
func (e *Executor) Init() error {
	e.log("=== Initializing Terraform ===")

	cmd := exec.Command("terraform", "init")
	cmd.Dir = e.workingDir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start terraform init: %w", err)
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		e.streamOutput(stdout, &stdoutBuf)
	}()
	go func() {
		defer wg.Done()
		e.streamOutput(stderr, &stderrBuf)
	}()

	wg.Wait()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("terraform init failed: %w\n%s", err, stderrBuf.String())
	}

	e.log("=== Terraform initialization complete ===")
	return nil
}

// Plan runs terraform plan
func (e *Executor) Plan() (string, error) {
	cmd := exec.Command("terraform", "plan", "-no-color")
	cmd.Dir = e.workingDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("terraform plan failed: %w\n%s", err, stderr.String())
	}

	return stdout.String(), nil
}

// Apply runs terraform apply
func (e *Executor) Apply() error {
	e.log("=== Applying Terraform configuration ===")

	cmd := exec.Command("terraform", "apply", "-auto-approve", "-no-color")
	cmd.Dir = e.workingDir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start terraform apply: %w", err)
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		e.streamOutput(stdout, &stdoutBuf)
	}()
	go func() {
		defer wg.Done()
		e.streamOutput(stderr, &stderrBuf)
	}()

	wg.Wait()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("terraform apply failed: %w\n%s", err, stderrBuf.String())
	}

	e.log("=== Terraform apply complete ===")
	return nil
}

// Destroy runs terraform destroy
func (e *Executor) Destroy() error {
	e.log("=== Destroying Terraform resources ===")

	cmd := exec.Command("terraform", "destroy", "-auto-approve", "-no-color")
	cmd.Dir = e.workingDir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start terraform destroy: %w", err)
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		e.streamOutput(stdout, &stdoutBuf)
	}()
	go func() {
		defer wg.Done()
		e.streamOutput(stderr, &stderrBuf)
	}()

	wg.Wait()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("terraform destroy failed: %w\n%s", err, stderrBuf.String())
	}

	e.log("=== Terraform destroy complete ===")
	return nil
}

// Output retrieves terraform output values
func (e *Executor) Output(name string) (string, error) {
	cmd := exec.Command("terraform", "output", "-raw", name)
	cmd.Dir = e.workingDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("terraform output failed: %w\n%s", err, stderr.String())
	}

	return stdout.String(), nil
}

// Show runs terraform show in JSON format
func (e *Executor) Show() (string, error) {
	cmd := exec.Command("terraform", "show", "-json")
	cmd.Dir = e.workingDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("terraform show failed: %w\n%s", err, stderr.String())
	}

	return stdout.String(), nil
}

// Validate checks if Terraform is installed and available
func (e *Executor) Validate() error {
	cmd := exec.Command("terraform", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("terraform not found: %w", err)
	}
	return nil
}

// CreateWorkingDir creates a working directory for an environment
func CreateWorkingDir(baseDir, envName string) (string, error) {
	envDir := filepath.Join(baseDir, envName)
	if err := os.MkdirAll(envDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create working directory: %w", err)
	}
	return envDir, nil
}

// RemoveWorkingDir removes the working directory for an environment
func RemoveWorkingDir(baseDir, envName string) error {
	envDir := filepath.Join(baseDir, envName)
	if err := os.RemoveAll(envDir); err != nil {
		return fmt.Errorf("failed to remove working directory: %w", err)
	}
	return nil
}
