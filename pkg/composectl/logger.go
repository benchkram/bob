package composectl

import (
	"fmt"
	"io"
	"sync"

	"github.com/docker/compose/v2/pkg/api"
	"github.com/logrusorgru/aurora"
)

var colorPool = []aurora.Color{
	aurora.BlueFg,
	aurora.GreenFg,
	aurora.CyanFg,
	aurora.MagentaFg,
	aurora.YellowFg,
	aurora.RedFg,
}

type logger struct {
	writer          io.Writer
	mutex           sync.Mutex
	containerColors map[string]aurora.Color
}

var _ api.LogConsumer = (*logger)(nil)

func NewLogConsumer(w io.Writer) (*logger, error) {
	return &logger{
		writer:          w,
		containerColors: map[string]aurora.Color{},
	}, nil
}

func (l *logger) colorize(cid string) string {
	color, ok := l.containerColors[cid]
	if ok {
		return aurora.Colorize(cid, color).String()
	}

	color = colorPool[len(l.containerColors)%len(colorPool)]
	l.containerColors[cid] = color

	return aurora.Colorize(cid, color).String()
}

func (l *logger) Log(_, container, msg string) {
	l.mutex.Lock()
	_, _ = l.writer.Write([]byte(fmt.Sprintf("[%s] %s\n", l.colorize(container), msg)))
	l.mutex.Unlock()
}

func (l *logger) Status(container, msg string) {
	l.mutex.Lock()
	_, _ = l.writer.Write([]byte(fmt.Sprintf("[%s] %s\n", l.colorize(container), msg)))
	l.mutex.Unlock()
}

func (l *logger) Register(container string) {
	l.mutex.Lock()
	_, _ = l.writer.Write([]byte(fmt.Sprintf("[%s] registered\n", l.colorize(container))))
	l.mutex.Unlock()
}
