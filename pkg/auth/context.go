package auth

type Context struct {
	// Name is the name of the context.
	Name string

	// Token identifies the user.
	Token string

	// Current is set to true if this should be the currently active context.
	Current bool
}
