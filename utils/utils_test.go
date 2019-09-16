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

func TestWeightPick(t *testing.T) {
	ws := []int{20, 20, 60}
	fmt.Println(WeightPick(ws), WeightPick(ws), WeightPick(ws), WeightPick(ws), WeightPick(ws))
}
