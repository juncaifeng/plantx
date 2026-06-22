package cmd

import (
	"fmt"
	"os"
	"os/exec"
)

func Test(args []string) {
	cmd := exec.Command("go", "test", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "tests failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("tests passed")
}
