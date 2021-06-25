package common

import "os"

func IsRunningInDocker() bool {
	_, err := os.Stat("/.dockerenv")
	if err != nil {
		return false
	}
	return true
}


func FindInt64(target int64, container []int64) bool {
	for _, i := range container{
		if i == target{
			return true
		}
	}
	return false
}
