package utils

import (
	"fmt"
	"testing"
)

func TestSalt(t *testing.T) {
	fmt.Println(Salt(18))
}
