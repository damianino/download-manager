package downloader

import (
	"fmt"
	"testing"
)

func TestMakeSections(t *testing.T){
	fmt.Println(makeSections(100, 10))
}