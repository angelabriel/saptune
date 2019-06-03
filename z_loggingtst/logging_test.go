package system

import (
	"github.com/SUSE/saptune/system"
	"testing"
)

func TestLog(t *testing.T) {
	logFile := "/var/log/tuned/tuned.log"
	system.LogInit()
	// Debug will only be written, if 'DEBUG="1"' is set in
	// /etc/sysconfig/saptune
	// the switch is set in LogInit()
	system.DebugLog("TestMessage%s_%s", "1", "Debug")
	if system.CheckForPattern(logFile, "TestMessage1_Debug") {
		t.Fatal("Debug message found in log file")
	}
	system.InfoLog("TestMessage%s_%s", "2", "Info")
	if !system.CheckForPattern(logFile, "TestMessage2_Info") {
		t.Fatal("Info message not found in log file")
	}
	system.WarningLog("TestMessage%s_%s", "3", "Warning")
	if !system.CheckForPattern(logFile, "TestMessage3_Warning") {
		t.Fatal("Warning message not found in log file")
	}
	system.ErrorLog("TestMessage%s_%s", "4", "Error")
	if !system.CheckForPattern(logFile, "TestMessage4_Error") {
		t.Fatal("Error message not found in log file")
	}
}
