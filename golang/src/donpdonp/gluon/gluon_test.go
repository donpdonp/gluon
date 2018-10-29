package main

import "testing"

func TestpickLang(t *testing.T) {
	url := "https://"
	contentType := "text/plain"
	if pickLang(url, contentType) != "javascript" {
		t.Error("")
	}
}
