package fileutil

import "os"

// Exist check if the file with the given path exits.
func Exist(path string) bool {
	_, err := os.Lstat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}

	return true
}

func ExistDir(path string) bool {
	fi, err := os.Lstat(path)
	if err != nil {
		if os.IsExist(err) && fi != nil && fi.IsDir() {
			return true
		}
		return false
	}
	if fi != nil && fi.IsDir() {
		return true
	}

	return false
}
