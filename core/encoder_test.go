package core_test

import (
	"github.com/orfjackal/gospec/src/gospec"
	"github.com/runningwild/sgf/core"
)

type encodable1 struct {
	A, B int32
}

type encodable2 struct {
	B byte
}

type encodable3 string

type encodable4 struct {
	A e4sub1
	G uint64
}

type e4sub1 struct {
	B, C int32
	D    e4sub2
}

type e4sub2 struct {
	E, F string
}

type encodable5 struct {
	A []byte
}

func EncoderSpec(c gospec.Context) {
	c.Specify("Test the coder.", func() {
		var er core.EncoderRegistry
		er.Register(encodable1{})
		er.Register(encodable2{})
		er.Register(encodable3([]byte{}))
		er.Register(encodable4{})
		er.Register(encodable5{})
		er.Complete()
		e1 := encodable1{1, 2}
		enc1, err := er.Encode(e1)
		c.Assume(err, gospec.Equals, error(nil))
		dec1, err := er.Decode(enc1)
		c.Assume(err, gospec.Equals, error(nil))
		d1, ok := dec1.(encodable1)
		c.Assume(ok, gospec.Equals, true)
		c.Expect(d1, gospec.Equals, e1)

		e2 := encodable2{14}
		enc2, err := er.Encode(e2)
		c.Assume(err, gospec.Equals, error(nil))
		dec2, err := er.Decode(enc2)
		c.Assume(err, gospec.Equals, error(nil))
		d2, ok := dec2.(encodable2)
		c.Assume(ok, gospec.Equals, true)
		c.Expect(d2, gospec.Equals, e2)

		e3 := encodable3("fudgeball")
		enc3, err := er.Encode(e3)
		c.Assume(err, gospec.Equals, error(nil))
		dec3, err := er.Decode(enc3)
		c.Assume(err, gospec.Equals, error(nil))
		d3, ok := dec3.(encodable3)
		c.Assume(ok, gospec.Equals, true)
		c.Expect(string(d3), gospec.Equals, string(e3))

		e4 := encodable4{
			A: e4sub1{
				B: 10,
				C: 20,
				D: e4sub2{
					E: "foo",
					F: "bar",
				},
			},
			G: 12345,
		}
		enc4, err := er.Encode(e4)
		c.Assume(err, gospec.Equals, error(nil))
		dec4, err := er.Decode(enc4)
		c.Assume(err, gospec.Equals, error(nil))
		d4, ok := dec4.(encodable4)
		c.Assume(ok, gospec.Equals, true)
		c.Expect(d4, gospec.Equals, e4)

		e5 := encodable5{
			A: []byte("Monkeyball"),
		}
		enc5, err := er.Encode(e5)
		c.Assume(err, gospec.Equals, error(nil))
		dec5, err := er.Decode(enc5)
		c.Assume(err, gospec.Equals, error(nil))
		d5, ok := dec5.(encodable5)
		c.Assume(ok, gospec.Equals, true)
		c.Expect(string(d5.A), gospec.Equals, string(e5.A))
	})
}
