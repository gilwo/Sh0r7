//go:build !prod

package common

func init() {
	devBuild = true
}
