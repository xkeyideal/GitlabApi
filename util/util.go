package util

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"runtime"
	"time"
)

func NewError(format string, a ...interface{}) error {

	err := fmt.Sprintf(format, a...)
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		return errors.New(err)
	}

	function := runtime.FuncForPC(pc).Name()
	msg := fmt.Sprintf("%s func:%s file:%s line:%d",
		err, function, file, line)
	return errors.New(msg)
}

/*
判断给定的filename是否是一个文件
*/
func IsFile(filename string) (bool, error) {

	if len(filename) <= 0 {
		return false, NewError("invalid filename")
	}

	stat, err := os.Stat(filename)
	if err != nil {
		return false, NewError("invalid path:" + filename)
	}

	if stat.IsDir() {
		return false, nil
	}

	return true, nil
}

/*
判断给定的filename是否是一个目录
*/
func IsDir(filename string) (bool, error) {

	if len(filename) <= 0 {
		return false, NewError("invalid dir")
	}

	stat, err := os.Stat(filename)
	if err != nil {
		return false, NewError("invalid path:" + filename)
	}

	if !stat.IsDir() {
		return false, nil
	}

	return true, nil
}

func IsExist(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func RunCmd(cmd string, arg ...string) (stdout, stderr string, err error) {
	c := exec.Command(cmd, arg...)
	var out bytes.Buffer
	var errStd bytes.Buffer
	c.Stdout = &out
	c.Stderr = &errStd
	err = c.Run()
	stdout = out.String()
	stderr = errStd.String()
	return
}

func RunCmdWithTimer(seconds int, cmd string, arg ...string) (stdout, stderr string, err error) {
	c := exec.Command(cmd, arg...)

	var out bytes.Buffer
	var errStd bytes.Buffer
	c.Stdout = &out
	c.Stderr = &errStd
	err = c.Start()

	done := make(chan error, 1)
	go func() {
		done <- c.Wait()
		close(done)
	}()

	select {
	case <-time.After(time.Duration(seconds) * time.Second):
		err = NewError("TIMEOUT. KILLED!")
		if e := c.Process.Kill(); e != nil {
			err = NewError("TIMEOUT and Kill CMD failed")
		} else {
			<-done
		}
	case err = <-done:
		stdout = out.String()
		stderr = errStd.String()
	}
	return
}

// Default Timeout: 60 seconds
func RunScript(scpt string, seconds ...int) (stdout, stderr string, err error) {
	var timeout int
	if len(seconds) > 0 {
		timeout = seconds[0]
	}
	if timeout == 0 {
		timeout = 60
	}
	runfile, err := ioutil.TempFile("", "scptempfile")
	if err != nil {
		return "", "Mk RunFile Fail.", err
	}
	defer os.Remove(runfile.Name())
	content := `#!/bin/bash
set -e
` + scpt
	//trap 'kill -s INT 0' EXIT
	if ok, err := WriteToFile(runfile.Name(), []byte(content)); ok {
		return RunCmdWithTimer(timeout, "/bin/bash", runfile.Name())
	} else {
		return "", "Run Script Fail.", err
	}
}

func PingRemote(remote string) (err error) {
	for tri := 5; tri > 0; tri-- {
		_, _, err = RunCmdWithTimer(12, "ssh", remote, "-o", "StrictHostKeyChecking=no")
		if err == nil {
			return
		}
		time.Sleep(time.Second)
	}
	return
}

// Default File Mode: 0664
func WriteToFile(filePath string, content []byte, perm ...os.FileMode) (bool, error) {
	err := os.MkdirAll(path.Dir(filePath), 0700)
	if err != nil {
		return false, NewError(fmt.Sprintf("MKDIR err: %s", err))
	}
	var fmode os.FileMode
	if len(perm) > 0 {
		fmode = perm[0]
	} else {
		fmode = 0664
	}
	err = ioutil.WriteFile(filePath, content, fmode)
	if err != nil {
		return false, NewError(fmt.Sprintf("Write File err: %s", err))
	}
	return true, nil
}
