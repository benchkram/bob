package composeutil

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
)

var (
	ErrInvalidProtocol = fmt.Errorf("invalid protocol")
)

// PortConfig holds data that are needed for port conflict resolution
type PortConfig struct {
	Service  *string
	Protocol string
	Port     uint32
}

// GetNewPortMappings returns any modified port mappings
func GetNewPortMappings(resolved map[string][]*PortConfig) string {
	ports := make([]string, 0, len(resolved))
	for port := range resolved {
		ports = append(ports, port)
	}
	sort.Strings(ports)

	mappings := "\nResolved host port mappings:\n"
	for _, port := range ports {
		services := resolved[port]
		for _, srv := range services {
			res := fmt.Sprintf("%d/%s", srv.Port, srv.Protocol)
			if port != res {
				if srv.Service == nil {
					continue
				}

				mappings += fmt.Sprintf(
					" - %-9s -> %-9s\t [ %s ]\n",
					port, res,
					*srv.Service,
				)
			}
		}
	}

	return mappings
}

// ResolvePortConflicts mutates the project and returns the port configs with any conflicting ports
// remapped
func ResolvePortConflicts(project *types.Project, configs map[string][]*PortConfig) (resolved map[string][]*PortConfig, err error) {
	// find in-use ports
	reserved := make([]string, 0, len(configs))
	for port := range configs {
		reserved = append(reserved, port)
	}

	// sort conflicting ports for deterministic resolution
	ports := make([]string, 0, len(configs))
	for port := range configs {
		ports = append(ports, port)
	}
	sort.Strings(ports)

	for _, port := range ports {
		services := configs[port]
		last := port

		for i, srv := range services {
			if srv.Service == nil || i == 0 {
				continue
			}

			var conflict uint32
			var proto string
			_, err := fmt.Sscanf(last, "%d/%s", &conflict, &proto)
			if err != nil {
				return nil, err
			}

			for j := conflict + 1; j < conflict+11; j++ {
				check := fmt.Sprintf("%d/%s", j, proto)

				avail, _ := PortAvailable(check)
				if !avail {
					continue
				}

				isReserved := false
				for _, p := range reserved {
					if p == check {
						isReserved = true
						break
					}
				}
				if isReserved {
					continue
				}

				for _, s := range project.Services {
					if s.Name == *srv.Service {
						for k, p := range s.Ports {
							if p.Published == conflict && p.Protocol == proto {
								p.Published = j
								srv.Port = j
								s.Ports[k] = p
							}
						}

					}
				}

				reserved = append(reserved, check)

				break
			}
		}
	}

	return configs, nil
}

// GetPortConflicts returns all the conflicting ports along with the services they were declared in
func GetPortConflicts(configs map[string][]*PortConfig) string {
	ports := make([]string, 0, len(configs))
	for port := range configs {
		ports = append(ports, port)
	}
	sort.Strings(ports)

	conflicts := "\nConflicting ports detected:"
	for _, port := range ports {
		services := configs[port]

		if len(services) > 1 {
			// port conflict detected

			serviceNames := make([]string, len(services))
			for i, srv := range services {
				if srv.Service != nil {
					serviceNames[i] = *srv.Service
				} else {
					serviceNames[i] = "(local)"
				}
			}

			conflicts += fmt.Sprintf(
				" - %-9s\t [ %s ]\n",
				port,
				strings.Join(serviceNames, ", "),
			)
		}
	}

	return conflicts
}

// HasPortConflicts returns true if there are two services using the same host port
func HasPortConflicts(configs map[string][]*PortConfig) bool {
	for _, services := range configs {
		if len(services) > 1 {
			// port conflict detected
			return true
		}
	}

	return false
}

// PortConfigs returns all configured ports for the given project
func PortConfigs(proj *types.Project) map[string][]*PortConfig {
	configs := make(map[string][]*PortConfig)

	// Services' order is undefined, sort them
	services := proj.Services
	sort.Slice(services, func(i, j int) bool {
		return services[i].Name < services[j].Name
	})

	for _, s := range services {
		for _, portcfg := range s.Ports {
			port := fmt.Sprintf("%d/%s", portcfg.Published, portcfg.Protocol)

			avail, _ := PortAvailable(port)
			if !avail {
				configs[port] = append(configs[port], &PortConfig{
					Service:  nil,
					Protocol: portcfg.Protocol,
					Port:     portcfg.Published,
				})
			}

			name := s.Name
			configs[port] = append(configs[port], &PortConfig{
				Service:  &name,
				Protocol: portcfg.Protocol,
				Port:     portcfg.Published,
			})
		}
	}

	return configs
}

// ProjectFromConfig loads a docker-compose config file into a compose Project
func ProjectFromConfig(path string) (p *types.Project, err error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	p, err = loader.Load(types.ConfigDetails{
		WorkingDir: filepath.Dir(path),
		ConfigFiles: []types.ConfigFile{
			{Filename: path, Content: b},
		},
	})

	if err != nil {
		return nil, err
	}

	if p.Name == "" {
		p.Name = path
	}

	return p, nil
}

// PortAvailable returns true if the port is not currently in use by the host
func PortAvailable(protoport string) (bool, error) {
	var proto string
	var port uint32
	_, err := fmt.Sscanf(protoport, "%d/%s", &port, &proto)
	if err != nil {
		return false, err
	}

	switch proto {
	case "tcp":
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			return false, err
		}

		return true, ln.Close()
	case "udp":
		ln, err := net.ListenPacket("udp", fmt.Sprintf(":%d", port))
		if err != nil {
			return false, err
		}

		return true, ln.Close()
	default:
		return false, ErrInvalidProtocol
	}
}
