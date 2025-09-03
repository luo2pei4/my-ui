package command

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

type CmdResult struct {
	Code      int
	StdOutput []byte
	ErrOutput []byte
	Err       error
}

// ExecCmd execute command line with /bin/bash
//
//	parameters:
//	  dur:  timeout duration
//	  args: command name/options/flags
//	return code:
//	  -1: call (*exec.Cmd)Start error
//	  -2: command execute timeout error
//	  others: the standard exit code of the command
func ExecCmd(dur time.Duration, args ...string) *CmdResult {

	ctx, cancel := context.WithTimeout(context.Background(), dur)
	defer cancel()

	result := &CmdResult{
		StdOutput: make([]byte, 0),
		ErrOutput: make([]byte, 0),
	}

	if len(args) == 0 {
		result.Code = -1
		result.Err = errors.New("invalid parameters, command is empty")
		return result
	}

	// assemble command line
	commandLine := strings.Join(args, " ")

	// create Cmd instance
	cmd := exec.Command("/bin/bash", "-c", commandLine)

	// set cmd attributes
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	if err := cmd.Start(); err != nil {
		// the Start method does not return exit code
		// if there is an error, set the exit code(-1) manually
		result.Code = -1
		result.Err = err
		return result
	}

	// wait until command exit
	notify := make(chan struct{})
	go func() {
		result.Err = cmd.Wait()
		close(notify)
	}()

	select {
	case <-notify:
	case <-ctx.Done():
		result.Code = -2
		result.Err = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}

	// when execute timeout, the value of code is -2
	// and the content of err is "signal: killed" in normal
	if result.Code == -2 {
		return result
	}

	// get standard exit status code
	result.Code = cmd.ProcessState.ExitCode()
	if stdout.Len() > 0 {
		result.StdOutput = stdout.Bytes()
	}
	if stderr.Len() > 0 {
		result.ErrOutput = stderr.Bytes()
	}

	return result
}

// IsTimeout whether timed out or not
func (cr *CmdResult) IsTimeout() bool {
	return cr.Code == -2
}

func (cr *CmdResult) HasError() bool {
	if cr.Code == 0 {
		return false
	}
	if cr.Err != nil || len(cr.ErrOutput) > 0 {
		return true
	}
	return false
}

// Error return error message
func (cr *CmdResult) Error() string {
	if cr.Code == -2 {
		errMsg := "command execution timed out"
		if cr.Err != nil {
			return errMsg + ", " + cr.Err.Error()
		}
		return errMsg
	}

	if cr.Code == -1 && cr.Err != nil {
		return cr.Err.Error()
	}

	var msg []byte
	if len(cr.ErrOutput) > 0 {
		msg = append(msg, cr.ErrOutput...)
	}
	if cr.Err != nil {
		if len(msg) > 0 {
			msg = append(append(msg, ','), ' ')
		}
		msg = append(msg, []byte(cr.Err.Error())...)
		return string(msg)
	}
	if len(msg) > 0 {
		return string(msg)
	}
	return ""
}
