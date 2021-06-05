package golog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"runtime"

	"github.com/getlantern/hidden"
)

// JsonOutput creates an output that writes JSON structured log to different io.Writers for errors and debug
func JsonOutput(errorWriter io.Writer, debugWriter io.Writer) Output {
	return &jsonOutput{
		E:  errorWriter,
		D:  debugWriter,
		pc: make([]uintptr, 10),
	}
}

type jsonOutput struct {
	// E is the error writer
	E io.Writer
	// D is the debug writer
	D  io.Writer
	pc []uintptr
}

type Event struct {
	Message   string                 `json:"msg,omitempty"`
	Component string                 `json:"component,omitempty"`
	Caller    string                 `json:"caller,omitempty"`
	Ops       map[string]interface{} `json:"ops,omitempty"`
	Severity  string                 `json:"level,omitempty"`
	Stack     string                 `json:"stack,omitempty"`
}

func (o *jsonOutput) Error(prefix string, skipFrames int, printStack bool, severity string, arg interface{}, values map[string]interface{}) {
	o.print(o.E, prefix, skipFrames, printStack, severity, arg, values)
}

func (o *jsonOutput) Debug(prefix string, skipFrames int, printStack bool, severity string, arg interface{}, values map[string]interface{}) {
	o.print(o.D, prefix, skipFrames, printStack, severity, arg, values)
}

func (o *jsonOutput) print(writer io.Writer, prefix string, skipFrames int, printStack bool, severity string, arg interface{}, values map[string]interface{}) {
	buf := bufferPool.Get()
	defer bufferPool.Put(buf)

	cleanPrefix := prefix[0 : len(prefix)-2] // prefix contains ': ' at the end, strip it
	event := Event{Component: cleanPrefix, Severity: severity, Caller: o.caller(skipFrames), Ops: values}
	if printStack {
		var b []byte
		buf := bytes.NewBuffer(b)
		getStack(buf, o.pc)
		event.Stack = buf.String()
	}
	encoder := json.NewEncoder(writer)
	if arg != nil {
		if ml, isMultiline := arg.(MultiLine); !isMultiline {
			event.Message = fmt.Sprintf("%v", arg)
		} else {
			mlp := ml.MultiLinePrinter()
			for {
				more := mlp(buf)
				buf.WriteByte('\n')
				if !more {
					break
				}
			}
			event.Message = hidden.Clean(buf.String())
		}
	}

	if err := encoder.Encode(event); err != nil {
		errorOnLogging(err)
	}
}

// returns the file and line number corresponding to the log message
func (o *jsonOutput) caller(skipFrames int) string {
	runtime.Callers(skipFrames, o.pc)
	funcForPc := runtime.FuncForPC(o.pc[0])
	file, line := funcForPc.FileLine(o.pc[0] - 1)
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}
