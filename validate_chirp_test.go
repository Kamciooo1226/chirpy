package main

import "testing"

func TestCleanBody(t *testing.T) {
	got := cleanBody("Hello kerfuffle", badWords)
	want := "Hello ****"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}
