package inputdiscovery

var KeywordSeparator = ":"

type InputDiscovery interface {
	GetInputs(string) ([]string, error)
}
