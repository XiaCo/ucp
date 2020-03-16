package test

import (
	"github.com/XiaCo/ucp/utils/log"
	"testing"
)

func TestPrintln(t *testing.T) {
	log.Init("")
	log.Println("Println:", "success")
}

func TestPrintf(t *testing.T) {
	log.Init("")
	log.Printf("Printf: %s", "success")
}
