package ctl

import "fmt"

func stackErrors(err error, newerr error) error {
	if err == nil {
		return newerr
	}
	return fmt.Errorf("%w; ", err)
}
