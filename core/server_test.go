package core_test

import (
	"fmt"
	"github.com/orfjackal/gospec/src/gospec"
	"github.com/runningwild/sgf/core"
)

func SimpleServerSpec(c gospec.Context) {
	c.Specify("Hook up all of the basic parts and make them talk.", func() {
		host, err := core.MakeHost("127.0.0.1", 1234)
		c.Assume(err, gospec.Equals, error(nil))
		client0, err := core.MakeClient("127.0.0.1", 1234)
		c.Assume(err, gospec.Equals, error(nil))
		client1, err := core.MakeClient("127.0.0.1", 1234)
		c.Assume(err, gospec.Equals, error(nil))
		fmt.Printf("%v %v %v\n", host, client0, client1)
	})
}
