package bj

import (
	"fmt"
	"github.com/pkg/term"
	"os"
)

func GetChar() []byte {
	t, _ := term.Open("/dev/tty")

	err := term.RawMode(t)
	if err != nil {
		fmt.Println("Cannot set term raw mode")
		os.Exit(1)
	}

	bytes := make([]byte, 3)
	numRead, err := t.Read(bytes)

	err2 := t.Restore()
	if err2 != nil {
		fmt.Println("Cannot restore term mode")
		os.Exit(1)
	}

	err3 := t.Close()
	if err3 != nil {
		fmt.Println("Cannot close term")
		os.Exit(1)
	}

	if err != nil {
		return nil
	}

	return bytes[0:numRead]
}
