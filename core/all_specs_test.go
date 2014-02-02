package core_test

import (
	"github.com/orfjackal/gospec/src/gospec"
	"testing"
)

func TestAllSpecs(t *testing.T) {
	r := gospec.NewRunner()
	r.AddSpec(SimpleSpec)
	r.AddSpec(BasicEncoderSpec)
	r.AddSpec(UnorderedEncoderSpec)
	gospec.MainGoTest(r, t)
}
