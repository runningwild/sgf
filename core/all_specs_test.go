package core_test

import (
	"github.com/orfjackal/gospec/src/gospec"
	"testing"
)

func TestAllSpecs(t *testing.T) {
	r := gospec.NewRunner()
	r.AddSpec(SimpleSpec)
	r.AddSpec(BasicTypeRegistrySpec)
	r.AddSpec(UnorderedTypeRegistrySpec)
	r.AddSpec(BlockWriteTypeRegistrySpec)
	gospec.MainGoTest(r, t)
}
