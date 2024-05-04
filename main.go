package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func run() error {
	args := os.Args[1:]
	if len(args) < 1 {
		args = []string{""}
	}

	log.SetPrefix("| ")
	log.SetFlags(log.Flags() | log.Lmsgprefix)

	switch strings.ToLower(args[0]) {
	case "dump":
		if len(args) < 2 {
			return fmt.Errorf("session-cookie is missing")
		}
		session := args[1]
		testName := ""
		if len(args) >= 3 {
			testName = args[2]
		}
		tests, err := dump(session, testName)
		if err != nil {
			return err
		}

		if testName == "" {
			var b strings.Builder

			b.WriteString("test is missing, select one:")
			for _, test := range tests {
				b.WriteRune('\n')
				b.WriteString(strings.ToLower(test.VendorTest))
			}
			return fmt.Errorf("%s", b.String())
		}
		return nil
	case "produce":
		if len(args) < 2 {
			var b strings.Builder

			entries, err := os.ReadDir(filepath.Join("out", "dump"))
			if err != nil {
				return err
			}

			b.WriteString("test is missing, select one:")
			for _, entry := range entries {
				if entry.IsDir() {
					b.WriteRune('\n')
					b.WriteString(strings.ToLower(entry.Name()))
				}
			}
			return fmt.Errorf("%s", b.String())
		}

		testName := args[1]
		return produce(testName)
	default:
		return fmt.Errorf("first argument must be 'dump' or 'produce'")
	}
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
