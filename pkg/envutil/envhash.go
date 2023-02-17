package envutil

type Hash string

func (h *Hash) String() string {
	return string(*h)
}
