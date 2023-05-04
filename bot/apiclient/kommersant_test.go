package apiclient

import "testing"

func TestCallKommersantAPI(t *testing.T) {
	result, err := CallKommersantAPI()
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal()
	}
}
