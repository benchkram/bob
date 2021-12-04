package tui

type Quit struct{}

type Started struct{}

type Restarted struct{}

type Update struct {
	tab int
}

type EnableScroll struct{}

type DisableScroll struct{}
