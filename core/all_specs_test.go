package core_test

import (
	"github.com/orfjackal/gospec/src/gospec"
	"testing"
	"time"
)

func init() {
	go func() {
		time.Sleep(time.Second * 20)
		panic("Late.")
	}()
}

func TestAllSpecs(t *testing.T) {
	r := gospec.NewRunner()
	r.AddSpec(SimpleServerSpec)
	r.AddSpec(BasicTypeRegistrySpec)
	r.AddSpec(UnorderedTypeRegistrySpec)
	r.AddSpec(BlockWriteTypeRegistrySpec)
	r.AddSpec(InterfaceTypeRegistrySpec)
	gospec.MainGoTest(r, t)
}
