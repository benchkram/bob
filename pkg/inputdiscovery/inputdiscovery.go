package inputdiscovery

var KeywordSeparator = ":"

type InputDiscovery interface {
	DiscoverInputs(string) ([]string, error)
}
