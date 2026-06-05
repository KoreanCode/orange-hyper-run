package app

import "strings"

type commandOutput struct {
	Stdout string
	Stderr string
}

func stdout(value string) commandOutput {
	if strings.TrimSpace(value) != "" && !strings.HasSuffix(value, "\n") {
		value += "\n"
	}
	return commandOutput{Stdout: value}
}

type hyperError struct {
	Message string
	Code    int
}

func newError(message string, code int) *hyperError {
	return &hyperError{Message: message, Code: code}
}

func ioError(err error) *hyperError {
	return newError(err.Error(), 1)
}

func dbError(err error) *hyperError {
	return newError(err.Error(), 1)
}
