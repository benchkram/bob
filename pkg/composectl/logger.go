package composectl

import (
	"fmt"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/logrusorgru/aurora"
	"io"
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
	_, _ = l.writer.Write([]byte(fmt.Sprintf("[%s] %s\n", l.colorize(container), msg)))
}

func (l *logger) Status(container, msg string) {
	_, _ = l.writer.Write([]byte(fmt.Sprintf("[%s] %s\n", l.colorize(container), msg)))
}

func (l *logger) Register(container string) {
	_, _ = l.writer.Write([]byte(fmt.Sprintf("[%s] registered\n", l.colorize(container))))
}
