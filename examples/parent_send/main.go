package main

import (
	axy "github.com/rnkv/axy-go"
	"github.com/rnkv/axy-go/examples/parent_send/actors/supervisors"
)

func main() {
	_ = axy.Spawn(supervisors.NewSupervisor())
	axy.Wait()
}
