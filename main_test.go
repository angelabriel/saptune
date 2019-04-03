package main

import (
	"os"
	"os/exec"
	"syscall"
	"testing"
)

func TestMain(m *testing.M) {
  log.SetOutput(ioutil.Discard)
  m.Run()
}

func TestPrintHelpAndExit(t *testing.T) {
	exitCode := 0
	if os.Getenv("DO_EXIT") == "1" {
		PrintHelpAndExit(9)
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestPrintHelpAndExit")
	cmd.Env = append(os.Environ(), "DO_EXIT=1")
	err := cmd.Run()
	e, ok := err.(*exec.ExitError)
	if ok {
		ws := e.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
		if exitCode != 9 {
			t.Fatalf("process ran with err %v, want exit status 9", err)
		}
		if !e.Success() {
			return
		}
	}
	t.Fatalf("process ran with err %v, want exit status 9", err)
}
