package composeutil

import (
	"fmt"
	"math"
	"net"
	"sort"
	"strconv"
	"strings"

	"github.com/compose-spec/compose-go/cli"
	"github.com/compose-spec/compose-go/types"
	"github.com/docker/compose/v2/pkg/api"

	"github.com/benchkram/bob/pkg/usererror"
)

type PortConfigs []PortConfig

func (c PortConfigs) String() string {
	s := ""

	for _, cfg := range c {
		s += fmt.Sprintf(
			" %s/%s\t%s\n",
			cfg.Port,
			cfg.Protocol,
			strings.Join(cfg.Services, ", "),
		)
	}

	return s
}

// ResolvePortConflicts mutates the project and returns the port configs with any conflicting ports
// remapped
func ResolvePortConflicts(conflicts PortConfigs) (PortConfigs, error) {
	// indexes reserved ports for easier lookup
	protoPortCfgs := map[string]bool{}
	for _, cfg := range conflicts {
		protoPortCfgs[protoPort(cfg.Port, cfg.Protocol)] = true
	}

	resolved := []PortConfig{}

	for _, cfg := range conflicts {
		for i, service := range cfg.Services {
			// skip the first service, as it's either the host or we want to keep this service's ports and change
			// the ports of the other services
			if i == 0 {
				resolved = append(resolved, PortConfig{
					Port:         cfg.Port,
					OriginalPort: cfg.Port,
					Protocol:     cfg.Protocol,
					Services:     []string{service},
				})
			}

			// check the next port
			intPort, _ := strconv.Atoi(cfg.Port)

			port := intPort + 1
			for {
				pp := protoPort(fmt.Sprint(port), cfg.Protocol)

				if _, ok := protoPortCfgs[pp]; !ok && PortAvailable(fmt.Sprint(port), cfg.Protocol) {
					// we found an available port that is not already reserved
					protoPortCfgs[pp] = true
					break
				}

				port += 1
				if port == math.MaxUint16 {
					return nil, fmt.Errorf("no ports available")
				}
			}

			resolved = append(resolved, PortConfig{
				Port:         fmt.Sprint(port),
				OriginalPort: cfg.Port,
				Protocol:     cfg.Protocol,
				Services:     []string{service},
			})
		}
	}

	return resolved, nil
}

func protoPort(port string, proto string) string {
	return fmt.Sprintf("%s/%s", port, proto)
}

func ApplyPortMapping(p *types.Project, mapping PortConfigs) {
	// index that associates service with its port configs
	servicePorts := map[string]PortConfigs{}
	for _, cfg := range mapping {
		service := cfg.Services[0]
		servicePorts[service] = append(servicePorts[service], cfg)
	}

	for i, service := range p.Services {
		servicePortCfgs := servicePorts[service.Name]

		for j, port := range service.Ports {
			for _, cfg := range servicePortCfgs {
				if port.Published == cfg.OriginalPort && port.Protocol == cfg.Protocol {
					port.Published = cfg.Port
					service.Ports[j] = port
					p.Services[i] = service
				}
			}
		}
	}
}

// PortConflicts returns all PortConfigs that have a port conflict
func PortConflicts(cfgs PortConfigs) PortConfigs {
	conflicts := []PortConfig{}

	for _, cfg := range cfgs {
		if len(cfg.Services) > 1 {
			// port conflict detected
			conflicts = append(conflicts, cfg)
		}
	}

	return conflicts
}

// HasPortConflicts returns true if there are two services using the same host port
func HasPortConflicts(cfgs PortConfigs) bool {
	for _, cfg := range cfgs {
		if len(cfg.Services) > 1 {
			// port conflict detected
			return true
		}
	}

	return false
}

type PortConfig struct {
	Port         string
	OriginalPort string
	Protocol     string
	Services     []string
}

// ProjectPortConfigs returns a slice of associations of port/proto to services, for a specific protocol, sorted by port
func ProjectPortConfigs(p *types.Project) PortConfigs {
	tcpCfg := portConfigs(p, "tcp")
	udpCfg := portConfigs(p, "udp")
	cfgs := append(tcpCfg, udpCfg...)

	// sort ports for consistent ordering
	sort.Slice(cfgs, func(i, j int) bool {
		return protoPort(cfgs[i].Port, cfgs[i].Protocol) > protoPort(cfgs[j].Port, cfgs[j].Protocol)
	})

	return cfgs
}

func portConfigs(proj *types.Project, typ string) PortConfigs {
	portServices := map[string][]string{}

	// services' order is undefined, sort them
	services := proj.Services
	sort.Slice(services, func(i, j int) bool {
		return services[i].Name < services[j].Name
	})

	for _, s := range services {
		for _, spCfg := range s.Ports {
			port := spCfg.Published

			if spCfg.Protocol == typ {
				if !PortAvailable(port, typ) && len(portServices[port]) == 0 {
					portServices[port] = append(portServices[port], "host")
				}

				portServices[port] = append(portServices[port], s.Name)
			}
		}
	}

	portCfgs := []PortConfig{}
	for port, services := range portServices {
		portCfgs = append(portCfgs, PortConfig{
			Port:         port,
			OriginalPort: port,
			Protocol:     typ,
			Services:     services,
		})
	}

	return portCfgs
}

// ProjectFromConfig loads a docker-compose config file into a compose Project
func ProjectFromConfig(composePath string) (p *types.Project, err error) {
	opts, err := cli.NewProjectOptions([]string{composePath})
	if err != nil {
		return nil, usererror.Wrapm(err, "error ")
	}

	p, err = cli.ProjectFromOptions(opts)
	if err != nil {
		return nil, usererror.Wrapm(err, "error loading docker-compose file")
	}

	for i, s := range p.Services {
		s.CustomLabels = map[string]string{
			api.ProjectLabel:     p.Name,
			api.ServiceLabel:     s.Name,
			api.VersionLabel:     api.ComposeVersion,
			api.WorkingDirLabel:  p.WorkingDir,
			api.ConfigFilesLabel: strings.Join(p.ComposeFiles, ","),
			api.OneoffLabel:      "False",
		}
		p.Services[i] = s
	}

	return p, nil
}

// PortAvailable returns true if the port is not currently in use by the host
func PortAvailable(port string, proto string) bool {
	switch proto {
	case "tcp":
		ln, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
		if err != nil {
			return false
		}

		_ = ln.Close()

		return true
	case "udp":
		ln, err := net.ListenPacket("udp", fmt.Sprintf(":%s", port))
		if err != nil {
			return false
		}

		_ = ln.Close()

		return true
	default:
		return false
	}
}
