package messager

import (
	"fmt"
	"io"
	"os"
	"encoding/json"
	"github.com/fatih/color"
)

type ResourceMessager struct {
	logWriter      io.Writer
	responseWriter io.Writer
}

var logger *ResourceMessager

func NewMessager(logWriter, responseWriter io.Writer) (*ResourceMessager) {
	return &ResourceMessager{logWriter, responseWriter}
}
func (rl *ResourceMessager) LogIt(args ...interface{}) {
	var text string
	text, ok := args[0].(string);
	if !ok {
		panic("Firt argument should be a string")
	}
	if len(args) > 1 {
		newArgs := rl.addColorToArgs(args[1:])
		fmt.Fprintf(rl.logWriter, text, newArgs...)
	} else {
		fmt.Fprint(rl.logWriter, text)
	}
}
func (rl *ResourceMessager) addColorToArgs(args []interface{}) ([]interface{}) {
	cyan := color.New(color.FgCyan).SprintFunc()
	var newArgs []interface{} = make([]interface{}, len(args))
	for index, arg := range args {
		newArgs[index] = cyan(arg)
	}
	return newArgs
}
func (rl *ResourceMessager) LogItLn(args ...interface{}) {
	var text string
	text, ok := args[0].(string);
	if !ok {
		panic("Firt argument should be a string")
	}
	args[0] = text + "\n"
	rl.LogIt(args...)
}
func (rl *ResourceMessager) SendJsonResponse(v interface{}) {
	json.NewEncoder(rl.responseWriter).Encode(v)
}
func (rl *ResourceMessager) GetLogWriter() (io.Writer) {
	return rl.logWriter
}
func (rl *ResourceMessager) GetResponseWriter() (io.Writer) {
	return rl.responseWriter
}
func GetMessager() (*ResourceMessager) {
	if logger == nil {
		logger = NewMessager(os.Stderr, os.Stdout)
	}
	return logger
}