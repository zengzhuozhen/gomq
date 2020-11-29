package common

import "os"

func IsRunningInDocker() bool {
	_, err := os.Stat("/.dockerenv")
	if err != nil {
		return false
	}
	return true
}

