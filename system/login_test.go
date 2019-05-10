package system

import (
	"testing"
)

func TestGetCurrentLogins(t *testing.T) {
	val := ""
	for _, userID := range GetCurrentLogins() {
		val = userID
	}
	if val == "" {
		t.Logf("no users currently logged in")
	} else {
		t.Logf("at least user '%s' is logged in\n", val)
	}
}

func TestSetTasksMax(t *testing.T) {
	userID := "1"
	val := "18446744073709"
	err := SetTasksMax(userID, val)
	if err != nil {
		t.Fatal(err)
	}
	value := GetTasksMax(userID)
	if value != val {
		t.Logf("expected '%s', actual '%s'\n", val, value)
	}
	val = "infinity"
	err = SetTasksMax(userID, val)
	if err != nil {
		t.Fatal(err)
	}
	value = GetTasksMax(userID)
	if value != val {
		t.Logf("expected '%s', actual '%s'\n", val, value)
	}
}
