package sonar_helpers

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func Check_sha1(filepath string, sha1hash string) (bool, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	hasher := sha1.New()

	if _, err := io.Copy(hasher, file); err != nil {
		return false, err
	}

	hashbytes := hasher.Sum(nil)
	sha1sum := hex.EncodeToString(hashbytes)
	fmt.Printf("We have sha1sum of %v \n", sha1sum)
	if sha1hash == sha1sum {
		return true, nil
	} else {
		return false, nil
	}
}

func Check_downloaded(filepath string, sha1hash string) (bool, error) {
	_, err := os.Open(filepath)
	if err != nil {
		return false, err
	}
	return true, nil
}
