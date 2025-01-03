package ch13

import (
	"bytes"
	"fmt"
	"log"
	"os"
)

func Example_logMultiWriter() {
	logFile := new(bytes.Buffer)
	w := SustainedMultiWriter(os.Stdout, logFile)
	l := log.New(w, "example: ", log.Lshortfile|log.Lmsgprefix)

	fmt.Println("standard output:")
	l.Print("Canada is south of Detroit")
	fmt.Print("\nlog file contents:\n", logFile.String())
}
