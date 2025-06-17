package output

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

// Color functions for different message types
var (
	Success = color.New(color.FgGreen).SprintFunc()
	Error   = color.New(color.FgRed).SprintFunc()
	Warning = color.New(color.FgYellow).SprintFunc()
	Info    = color.New(color.FgBlue).SprintFunc()
	Bold    = color.New(color.Bold).SprintFunc()
)

// Print functions for different message types
func PrintSuccess(format string, a ...interface{}) {
	fmt.Fprintf(os.Stdout, Success(format)+"\n", a...)
}

func PrintError(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, Error(format)+"\n", a...)
}

func PrintWarning(format string, a ...interface{}) {
	fmt.Fprintf(os.Stdout, Warning(format)+"\n", a...)
}

func PrintInfo(format string, a ...interface{}) {
	fmt.Fprintf(os.Stdout, Info(format)+"\n", a...)
}

func Print(format string, a ...interface{}) {
	fmt.Fprintf(os.Stdout, format+"\n", a...)
}

// Fatal prints an error message and exits with code 1
func Fatal(format string, a ...interface{}) {
	PrintError(format, a...)
	os.Exit(1)
}

// PrintErrorf prints a formatted error message without exiting
func PrintErrorf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, Error(format), a...)
}

// Prompt asks the user for input with a colored prompt
func Prompt(prompt string) string {
	fmt.Print(Info(prompt + ": "))
	var input string
	_, _ = fmt.Scanln(&input)
	return input
}

// Debug prints a debug message if debug mode is enabled
func Debug(format string, a ...interface{}) {
	if os.Getenv("CCMD_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[DEBUG] "+format+"\n", a...)
	}
}
