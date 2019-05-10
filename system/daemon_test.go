package system

import (
	"testing"
)


func TestSystemctl(t *testing.T) {
	if !CmdIsAvailable("/usr/bin/systemctl") {
		t.Skip("command '/usr/bin/systemctl' not available. Skip tests")
	}

	testService := "sshd.service"
	if err := SystemctlEnable(testService); err != nil {
		t.Fatal(err)
	}
	if err := SystemctlDisable(testService); err != nil {
		t.Fatal(err)
	}
	if err := SystemctlStart(testService); err != nil {
		t.Fatal(err)
	}
	if !SystemctlIsRunning(testService) {
		t.Fatalf("service '%s' not running\n", testService)
	}
	if err := SystemctlRestart(testService); err != nil {
		t.Fatal(err)
	}
	if !SystemctlIsRunning(testService) {
		t.Fatalf("service '%s' not running\n", testService)
	}
	if err := SystemctlStop(testService); err != nil {
		t.Fatal(err)
	}
	if SystemctlIsRunning(testService) {
		t.Fatalf("service '%s' still running\n", testService)
	}
	if err := SystemctlEnableStart(testService); err != nil {
		t.Fatal(err)
	}
	if !SystemctlIsRunning(testService) {
		t.Fatalf("service '%s' not running\n", testService)
	}
	if err := SystemctlDisableStop(testService); err != nil {
		t.Fatal(err)
	}
	if SystemctlIsRunning(testService) {
		t.Fatalf("service '%s' still running\n", testService)
	}
}

func TestSystemctlIsRunning(t *testing.T) {
	// check, if command is available
	if !CmdIsAvailable("/usr/bin/systemctl") {
		t.Skip("command '/usr/bin/systemctl' not available. Skip tests")
	}
	if !SystemctlIsRunning("dbus.service") {
		t.Fatal("'dbus.service' not running")
	}
	if !SystemctlIsRunning("tuned.service") {
		t.Log("'tuned.service' not running")
	}
}

func TestWriteTunedAdmProfile(t *testing.T) {
	profileName := "balanced"
	if err := WriteTunedAdmProfile(profileName); err != nil {
		t.Fatal(err)
	}
	if !CheckForPattern("/etc/tuned/active_profile", profileName) {
		t.Fatal("wrong profile in '/etc/tuned/active_profile'")
	}
	actProfile := GetTunedProfile()
	if actProfile != profileName {
		t.Fatalf("expected profile '%s', current profile '%s'\n", profileName, actProfile)
	}
	profileName = ""
	if err := WriteTunedAdmProfile(profileName); err != nil {
		t.Fatal(err)
	}
	actProfile = GetTunedProfile()
	if actProfile != "" {
		t.Fatalf("expected profile '%s', current profile '%s'\n", profileName, actProfile)
	}
}

func TestGetTunedProfile(t *testing.T) {
	actVal := GetTunedProfile()
	if actVal == "" {
		t.Log("seams there is no tuned profile, skip test")
	}
}

func TestTunedAdmOff(t *testing.T) {
	if !CmdIsAvailable("/usr/sbin/tuned-adm") {
		t.Skip("command '/usr/sbin/tuned-adm' not available. Skip tests")
	}
	if err := TunedAdmOff(); err != nil {
		//ANGI TODO - switch to t.Fatal(err), if tuned.service is running in the docker container
		t.Logf("seams 'tuned-adm off' does not work: '%v'\n", err)
	}
	actProfile := GetTunedProfile()
	if actProfile != "" {
		t.Fatalf("expected profile '%s', current profile '%s'\n", "", actProfile)
	}
	if err := SystemctlStop("tuned"); err != nil {
		t.Fatal(err)
	}
}

func TestTunedAdmProfile(t *testing.T) {
	profileName := "balanced"
	if !CmdIsAvailable("/usr/sbin/tuned-adm") {
		t.Skip("command '/usr/sbin/tuned-adm' not available. Skip tests")
	}
	if err := TunedAdmProfile(profileName); err != nil {
		//ANGI TODO - switch to t.Fatal(err), if tuned.service is running in the docker container
		t.Logf("seams 'tuned-adm profile balanced' does not work: '%v'\n", err)
	}
	actProfile := GetTunedProfile()
	if actProfile != profileName {
		//ANGI TODO - switch to t.Fatalf(txt), if tuned.service is running in the docker container
		t.Logf("expected profile '%s', current profile '%s'\n", profileName, actProfile)
	}
	if err := TunedAdmOff(); err != nil {
		//ANGI TODO - switch to t.Fatal(err), if tuned.service is running in the docker container
		t.Logf("seams 'tuned-adm off' does not work: '%v'\n", err)
	}
	if err := SystemctlStop("tuned"); err != nil {
		t.Fatal(err)
	}
}

func TestGetTunedAdmProfile(t *testing.T) {
	// check, if command is available
	if !CmdIsAvailable("/usr/sbin/tuned-adm") {
		t.Skip("command '/usr/sbin/tuned-adm' not available. Skip tests")
	}
	actVal := GetTunedAdmProfile()
	if actVal == "" {
		t.Log("seams there is no tuned profile")
	}
}
