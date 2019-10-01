package loggy

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

type Log struct {
	Service  string
	Message  []interface{}
	LogLevel int
	Source   string
	Color    string
}

type LogManager struct {
	LogLevel int
	LogQueue chan Log
}

func getFrame(skipFrames int) runtime.Frame {
	// We need the frame at index skipFrames+2, since we never want runtime.Callers and getFrame
	targetFrameIndex := skipFrames + 2

	// Set size to targetFrameIndex+2 to ensure we have room for one more caller than we need
	programCounters := make([]uintptr, targetFrameIndex+2)
	n := runtime.Callers(0, programCounters)

	frame := runtime.Frame{Function: "unknown"}
	if n > 0 {
		frames := runtime.CallersFrames(programCounters[:n])
		for more, frameIndex := true, 0; more && frameIndex <= targetFrameIndex; frameIndex++ {
			var frameCandidate runtime.Frame
			frameCandidate, more = frames.Next()
			if frameIndex == targetFrameIndex {
				frame = frameCandidate
			}
		}
	}

	return frame
}

func (l *LogManager) Start(logLevel int) {
	l.LogLevel = logLevel
	l.LogQueue = make(chan Log, 1000)
	for {
		select {
		case log := <-l.LogQueue:
			if log.LogLevel <= logLevel {
				t := time.Now().Format("060102:15:04.05")
				fmt.Printf("[%v]%v[%v/%v] %v%v\n", t, log.Color, log.Service, log.Source, fmt.Sprint(log.Message...), "\u001b[0m")
			}
		}
	}
}

func (l *LogManager) Spawn(serviceName string) *LogPrinter {
	printer := &LogPrinter{Service: serviceName, Manager: l}
	return printer
}

type LogPrinter struct {
	Service string
	Manager *LogManager
}

func (p *LogPrinter) Info(message ...interface{}) {
	source := getFrame(1).Function
	sourcesplit := strings.Split(source, "/")
	p.Manager.LogQueue <- Log{Service: p.Service, Source: sourcesplit[len(sourcesplit)-1], Message: message, LogLevel: 0, Color: "ℹ️\u001b[1m"}
}

func (p *LogPrinter) Error(message ...interface{}) {
	source := getFrame(1).Function
	sourcesplit := strings.Split(source, "/")
	p.Manager.LogQueue <- Log{Service: p.Service, Source: sourcesplit[len(sourcesplit)-1], Message: message, LogLevel: 1, Color: "❌\u001b[31;1m"}
}

func (p *LogPrinter) Verbose(message ...interface{}) {
	source := getFrame(1).Function
	sourcesplit := strings.Split(source, "/")
	p.Manager.LogQueue <- Log{Service: p.Service, Source: sourcesplit[len(sourcesplit)-1], Message: message, LogLevel: 2, Color: "☁️\u001b[33;1m"}
}

func (p *LogPrinter) Debug(message ...interface{}) {
	source := getFrame(1).Function
	sourcesplit := strings.Split(source, "/")
	p.Manager.LogQueue <- Log{Service: p.Service, Source: sourcesplit[len(sourcesplit)-1], Message: message, LogLevel: 3, Color: "🐛\u001b[36;1m"}
}
