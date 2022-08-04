package global

// EnvWhitelist keep the whitelisted env variables when hermetic mode is on
var EnvWhitelist = []string{"HOME", "XDG_CACHE_HOME", "NIX_PATH"}
