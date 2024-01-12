package tt

import (
	"fmt"
	"os"
	"testing"
)

func TestValidateWithFfprobe(t *testing.T) {
	s, e := os.Getwd()
	fmt.Printf("%v %v", s, e)

	good, err := ValidateWithFfprobe()("../downloads/broken.mp4")
	if err != nil {
		fmt.Printf("good: %v, err %v", good, err)
	}

	good, err = ValidateWithFfprobe()("../downloads/fine.mp4")
	if err != nil {
		fmt.Printf("good: %v, err %v", good, err)
	}
}
