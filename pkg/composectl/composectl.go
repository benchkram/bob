package composectl

import (
	"bytes"
	"context"
	"fmt"

	"github.com/Benchkram/errz"

	"github.com/compose-spec/compose-go/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose-cli/pkg/api"
	"github.com/docker/compose-cli/pkg/compose"
)

var (
	ErrInvalidProject = fmt.Errorf("invalid project")
	ErrComposeError   = fmt.Errorf("compose error")
)

type ComposeController struct {
	ctx     context.Context
	project *types.Project
	service api.Service
	stdout  *bytes.Buffer
	stderr  *bytes.Buffer
	logger  *logger
}

func New(ctx context.Context) (*ComposeController, error) {
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)

	cli, err := command.NewDockerCli(
		command.WithOutputStream(stdout),
		command.WithErrorStream(stderr),
	)
	if err != nil {
		return nil, err
	}

	err = cli.Initialize(flags.NewClientOptions())
	if err != nil {
		return nil, err
	}

	service := compose.NewComposeService(cli.Client(), &configfile.ConfigFile{})

	return &ComposeController{
		ctx:     ctx,
		service: service,
		stdout:  stdout,
		stderr:  stderr,
	}, nil
}

func (ctl *ComposeController) Up(project *types.Project) error {
	project.Name = "bob"
	ctl.project = project

	ctl.logger = new(logger)
	logger, err := NewLogger()
	if err != nil {
		return err
	}
	ctl.logger = logger

	err = ctl.service.Up(ctl.ctx, project, api.UpOptions{})
	if err != nil {
		return err
	}

	go func() {
		err := ctl.service.Logs(ctl.ctx, ctl.project.Name, ctl.logger, api.LogOptions{
			Services:   nil,
			Tail:       "",
			Since:      "",
			Until:      "",
			Follow:     true,
			Timestamps: true,
		})
		if err != nil {
			errz.Log(err)
		}
	}()

	return nil
}

func (ctl *ComposeController) Down() error {
	if ctl.project == nil {
		return ErrInvalidProject
	}

	err := ctl.service.Down(ctl.ctx, ctl.project.Name, api.DownOptions{
		Project: ctl.project,
	})
	if err != nil {
		return err
	}

	if ctl.stderr.String() != "" {
		return ErrComposeError
	}

	return nil
}

func (ctl *ComposeController) Stdout() string {
	return ctl.stdout.String()
}

func (ctl *ComposeController) Stderr() string {
	return ctl.stderr.String()
}
