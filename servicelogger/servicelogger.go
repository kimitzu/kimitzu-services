package servicelogger

import "fmt"

type Log struct {
	Service  string
	Message  interface{}
	LogLevel int
	Color    string
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
				fmt.Printf("[%v] %v%v%v\n", log.Service, log.Color, log.Message, "\u001b[0m")
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
	p.Manager.LogQueue <- Log{Service: p.Service, Message: message, LogLevel: 0, Color: "\u001b[1m"}
}

func (p *LogPrinter) Error(message interface{}) {
	p.Manager.LogQueue <- Log{Service: p.Service, Message: message, LogLevel: 1, Color: "\u001b[31;1m"}
}

func (p *LogPrinter) Verbose(message interface{}) {
	p.Manager.LogQueue <- Log{Service: p.Service, Message: message, LogLevel: 2, Color: "\u001b[33;1m"}
}

func (p *LogPrinter) Debug(message interface{}) {
	p.Manager.LogQueue <- Log{Service: p.Service, Message: message, LogLevel: 3, Color: "\u001b[36;1m"}
}
