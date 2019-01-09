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
	if actualVal != "12" || actualVal != "12-SP1" || actualVal != "12-SP2" || actualVal != "12-SP3" || actualVal != "12-SP4" || actualVal != "15" || actualVal != "15-SP1" {
		t.Logf("unexpected OS version '%s'\n", actualVal)
	}
}
