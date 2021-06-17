package composectl

import (
	"errors"
	"log"
	"os"

	"github.com/docker/compose-cli/pkg/api"
)

type logger struct {
	*log.Logger
}

var _ api.LogConsumer = (*logger)(nil)

func NewLogger() (*logger, error) {
	// TODO: replace file logger with TUI
	err := os.Mkdir("logs", os.ModePerm)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return nil, err
	}
	logfile, err := os.OpenFile(
		"logs/debug.log",
		os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666,
	)
	if err != nil {
		return nil, err
	}
	return &logger{
		log.New(logfile, "", 0),
	}, nil
}

func (l *logger) Log(service, container, message string) {
	l.Logger.Println("Log", service, container, message)
}

func (l *logger) Status(container, msg string) {
	l.Logger.Println("Status", container, msg)
}

func (l *logger) Register(container string) {
	l.Logger.Println("Register", container)
}
