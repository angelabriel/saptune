package system

import (
	"testing"
)

func TestIsUserRoot(t *testing.T) {
	if !IsUserRoot() {
		t.Log("the test requires root access")
	}
}

func TestGetOsName(t *testing.T) {
	actualVal := GetOsName()
	if actualVal != "SLES" {
		t.Logf("OS is '%s' and not 'SLES'\n", actualVal)
	}
	if actualVal == "" {
		t.Logf("empty value returned for the os Name")
	}
}

func TestGetOsVers(t *testing.T) {
	actualVal := GetOsVers()
	switch actualVal {
	case "12", "12-SP1", "12-SP2", "12-SP3", "12-SP4", "15", "15-SP1":
		t.Logf("expected OS version '%s' found\n", actualVal)
	default:
		t.Logf("unexpected OS version '%s'\n", actualVal)
	}
}

func TestCmdIsAvailable(t *testing.T) {
	if !CmdIsAvailable("/usr/bin/go") {
		t.Fatal("'/usr/bin/go' not found")
	}
	if CmdIsAvailable("/cmd_not_avail.CnA") {
		t.Fatal("found '/cmd_not_avail.CnA'")
	}
}

func TestCheckForPattern(t *testing.T) {
	if CheckForPattern("/file_not_available", "Test_Text") {
		t.Fatal("found '/file_not_available'")
	}
}
