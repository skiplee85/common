package utils

import (
	"fmt"
	"testing"
)

func TestSalt(t *testing.T) {
	fmt.Println(Salt(18))
}

func TestRandInt(t *testing.T) {
	fmt.Println(RandInt(100), RandInt(100), RandInt(100), RandInt(100), RandInt(100))
}
