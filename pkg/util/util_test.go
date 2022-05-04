package util

import (
	"strings"
	"testing"
)

func Test_RandString(t *testing.T) {
	length := 5
	res := RandString(length)

	if len(res) != length {
		t.Errorf("Res Lenght should be equal to lenght")
	}
	// Check if in Possible Letters
	for _, myrune := range res {
		if !strings.ContainsRune(letterBytes, myrune) {
			t.Errorf("A Letter of the final result is not in LetterBytes.")
		}
	}
}
