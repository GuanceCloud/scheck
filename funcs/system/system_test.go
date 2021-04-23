package system

import (
	"log"
	"os"
	"testing"
)

func TestSystem(t *testing.T) {

	st, _ := os.Stat("/etc")
	log.Printf("%s", st.Mode().String())

}
