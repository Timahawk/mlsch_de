package chat

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
	for _, rune := range res {
		if !strings.ContainsRune(letterBytes, rune) {
			t.Errorf("A Letter of the final result is not in LetterBytes.")
		}
	}
}

func Test_newHub(t *testing.T) {

}
