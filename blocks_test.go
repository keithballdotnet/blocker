package blocks

import (
	"fmt"
	"testing"
)

// A test to create a new file
func TestCreateNewFile(t *testing.T) {

	// NOTE:  Change this path
	err, Id := CreateFile("/home/keithball/Projects/src/blocks/src/github.com/Inflatablewoman/blocks/Resources/tempest.txt")

	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("Created new file with ID: ", Id)
}
