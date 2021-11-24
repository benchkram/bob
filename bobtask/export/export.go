package export

type E string

func (e *E) String() string {
	return string(*e)
}
