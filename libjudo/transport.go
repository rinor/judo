package libjudo

import (
	"fmt"
	"os"
	"time"
)

func (host *Host) pushFiles(job *Job,
	fname_local string, fname_remote string) (err error) {
	var remote = fmt.Sprintf("[%s]:%s", host.Name, fname_remote)
	proc, err := NewProc("scp", "-r", fname_local, remote)
	if err != nil {
		return
	}
	close(proc.Stdin)
	for {
		select {
		case line := <-proc.Stdout:
			host.Log(line)
		case line := <-proc.Stderr:
			host.Log(line)
		case err = <-proc.Done:
			return err
		case <-time.After(job.Timeout):
			return TimeoutError{}
		case <-host.cancel:
			proc.Signal(os.Interrupt)
			return CancelError{}
		}
	}
}

func (host *Host) Ssh(job *Job,
	command string, args ...string) (err error) {
	proc, err := NewProc(
		"ssh",
		append([]string{host.Name, command}, args...)...,
	)
	if err != nil {
		return
	}
	close(proc.Stdin)
	for {
		select {
		case line := <-proc.Stdout:
			host.Log(line)
		case line := <-proc.Stderr:
			host.Log(line)
		case err = <-proc.Done:
			return err
		case <-time.After(job.Timeout):
			return TimeoutError{}
		case <-host.cancel:
			proc.Signal(os.Interrupt)
			return CancelError{}
		}
	}
}

func (host *Host) SshRead(job *Job,
	command string, args ...string) (out string, err error) {
	proc, err := NewProc(
		"ssh",
		append([]string{host.Name, command}, args...)...,
	)
	if err != nil {
		return
	}
	close(proc.Stdin)
	for {
		select {
		case line := <-proc.Stdout:
			out = line
		case line := <-proc.Stderr:
			host.Log(line)
		case err = <-proc.Done:
			return
		case <-time.After(job.Timeout):
			return "", TimeoutError{}
		case <-host.cancel:
			proc.Signal(os.Interrupt)
			return "", CancelError{}
		}
	}
}
