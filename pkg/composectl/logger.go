package composectl

import (
	"fmt"
	"io"

	"github.com/docker/compose/v2/pkg/api"
)

type logger struct {
	writer io.Writer
}

var _ api.LogConsumer = (*logger)(nil)

func NewLogger(w io.Writer) (*logger, error) {
	return &logger{
		writer: w,
	}, nil
}

func (l *logger) Log(service, container, message string) {
	_, err := l.writer.Write([]byte(fmt.Sprintln("Log", service, container, message)))
	if err != nil {
		panic(err)
	}
}

func (l *logger) Status(container, msg string) {
	_, err := l.writer.Write([]byte(fmt.Sprintln("Status", container, msg)))
	if err != nil {
		panic(err)
	}
}

func (l *logger) Register(container string) {
	_, err := l.writer.Write([]byte(fmt.Sprintln("Register", container)))
	if err != nil {
		panic(err)
	}

}
