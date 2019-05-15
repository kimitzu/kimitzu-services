package servicelogger

import "fmt"

type Log struct {
	Service  string
	Message  interface{}
	LogLevel int
}

type LogManager struct {
	LogLevel int
	LogQueue chan Log
}

func (l *LogManager) Start(logLevel int) {
	l.LogLevel = logLevel
	l.LogQueue = make(chan Log, 100)
	for {
		select {
		case log := <-l.LogQueue:
			if log.LogLevel <= logLevel {
				fmt.Printf("[%v] %v\n", log.Service, log.Message)
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

func (p *LogPrinter) Info(message interface{}) {
	p.Manager.LogQueue <- Log{Service: p.Service, Message: message, LogLevel: 0}
}
