package main

import (
	"fmt"
	"testing"
)

func TestUUID(t *testing.T) {
	uuid, err := GetSystemUUID()

	if err != nil {
		panic(err)
	}
	fmt.Println(uuid)
}

func main() {

}
