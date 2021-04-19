package common

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	// "sync"
)

// StreamingExec models a generic command execution that consumers can use to
// execute commands and stream their output to an io.Writer. For example
// compute commands can use this to standardize the flow control for each
// compiler toolchain.
type StreamingExec struct {
	command string
	args    []string
	env     []string
	verbose bool
	output  io.Writer
	process *os.Process
}

// NewStreamingExec constructs a new StreamingExec instance.
func NewStreamingExec(cmd string, args, env []string, verbose bool, out io.Writer) *StreamingExec {
	return &StreamingExec{
		cmd,
		args,
		env,
		verbose,
		out,
		nil,
	}
}

// Exec executes the compiler command and pipes the child process stdout and
// stderr output to the supplied io.Writer, it waits for the command to exit
// cleanly or returns an error.
func (s StreamingExec) Exec() error {
	// Construct the command with given arguments and environment.
	//
	// gosec flagged this:
	// G204 (CWE-78): Subprocess launched with variable
	// Disabling as the variables come from trusted sources.
	/* #nosec */
	cmd := exec.Command(s.command, s.args...)
	cmd.Env = append(os.Environ(), s.env...)

	// Store off Process so it can be killed by signals
	s.process = cmd.Process

	// Pipe the child process stdout and stderr to our own output writer.
	var stderrBuf bytes.Buffer
	cmd.Stdout = s.output
	cmd.Stderr = io.MultiWriter(s.output, &stderrBuf)

	if err := cmd.Run(); err != nil {
		if !s.verbose && stderrBuf.Len() > 0 {
			return fmt.Errorf("error during execution process:\n%s", strings.TrimSpace(stderrBuf.String()))
		}
		return fmt.Errorf("error during execution process")
	}

	return nil
}

// Signal enables spawned subprocess to accept given signal.
func (s StreamingExec) Signal(signal os.Signal) error {
	if s.process != nil {
		err := s.process.Signal(signal)
		if err != nil {
			return err
		}
	}
	return nil
}
