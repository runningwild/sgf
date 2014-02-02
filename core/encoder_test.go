package core_test

import (
	"bytes"
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

func (encodable1) id() int {
	return 1
}
func (encodable2) id() int {
	return 2
}
func (encodable3) id() int {
	return 3
}
func (encodable4) id() int {
	return 4
}
func (encodable5) id() int {
	return 5
}

type ider interface {
	id() int
}

func BasicEncoderSpec(c gospec.Context) {
	c.Specify("Test the coder.", func() {
		var er core.EncoderRegistry
		er.Register(encodable1{})
		er.Register(encodable2{})
		er.Register(encodable3([]byte{}))
		er.Register(encodable4{})
		er.Register(encodable5{})
		er.Complete()
		e1 := encodable1{1, 2}
		buf := bytes.NewBuffer(nil)
		err := er.Encode(e1, buf)
		c.Assume(err, gospec.Equals, error(nil))
		dec1, err := er.Decode(buf)
		c.Assume(err, gospec.Equals, error(nil))
		d1, ok := dec1.(encodable1)
		c.Assume(ok, gospec.Equals, true)
		c.Expect(dec1.(ider).id(), gospec.Equals, 1)
		c.Expect(d1, gospec.Equals, e1)

		e2 := encodable2{14}
		buf = bytes.NewBuffer(nil)
		err = er.Encode(e2, buf)
		c.Assume(err, gospec.Equals, error(nil))
		dec2, err := er.Decode(buf)
		c.Assume(err, gospec.Equals, error(nil))
		d2, ok := dec2.(encodable2)
		c.Assume(ok, gospec.Equals, true)
		c.Expect(dec2.(ider).id(), gospec.Equals, 2)
		c.Expect(d2, gospec.Equals, e2)

		e3 := encodable3("fudgeball")
		buf = bytes.NewBuffer(nil)
		err = er.Encode(e3, buf)
		c.Assume(err, gospec.Equals, error(nil))
		dec3, err := er.Decode(buf)
		c.Assume(err, gospec.Equals, error(nil))
		d3, ok := dec3.(encodable3)
		c.Assume(ok, gospec.Equals, true)
		c.Expect(dec3.(ider).id(), gospec.Equals, 3)
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
		buf = bytes.NewBuffer(nil)
		err = er.Encode(e4, buf)
		c.Assume(err, gospec.Equals, error(nil))
		dec4, err := er.Decode(buf)
		c.Assume(err, gospec.Equals, error(nil))
		d4, ok := dec4.(encodable4)
		c.Assume(ok, gospec.Equals, true)
		c.Expect(dec4.(ider).id(), gospec.Equals, 4)
		c.Expect(d4, gospec.Equals, e4)

		e5 := encodable5{
			A: []byte("Monkeyball"),
		}
		buf = bytes.NewBuffer(nil)
		err = er.Encode(e5, buf)
		c.Assume(err, gospec.Equals, error(nil))
		dec5, err := er.Decode(buf)
		c.Assume(err, gospec.Equals, error(nil))
		d5, ok := dec5.(encodable5)
		c.Assume(ok, gospec.Equals, true)
		c.Expect(dec5.(ider).id(), gospec.Equals, 5)
		c.Expect(string(d5.A), gospec.Equals, string(e5.A))
	})
}

func UnorderedEncoderSpec(c gospec.Context) {
	c.Specify("Test that the coder can decode packets out of order.", func() {
		var er core.EncoderRegistry
		er.Register(encodable1{})
		er.Register(encodable2{})
		er.Complete()

		v1 := encodable1{1, 2}
		v2 := encodable2{3}
		v3 := encodable2{4}
		v4 := encodable1{5, 6}

		b1 := bytes.NewBuffer(nil)
		err := er.Encode(v1, b1)
		c.Assume(err, gospec.Equals, error(nil))

		b2 := bytes.NewBuffer(nil)
		err = er.Encode(v2, b2)
		c.Assume(err, gospec.Equals, error(nil))

		b3 := bytes.NewBuffer(nil)
		err = er.Encode(v3, b3)
		c.Assume(err, gospec.Equals, error(nil))

		b4 := bytes.NewBuffer(nil)
		err = er.Encode(v4, b4)
		c.Assume(err, gospec.Equals, error(nil))

		buf := bytes.NewBuffer(nil)
		buf.Write(b1.Bytes())
		buf.Write(b3.Bytes())
		buf.Write(b4.Bytes())
		buf.Write(b2.Bytes())

		d1, err := er.Decode(buf)
		c.Assume(err, gospec.Equals, error(nil))
		d3, err := er.Decode(buf)
		c.Assume(err, gospec.Equals, error(nil))
		d4, err := er.Decode(buf)
		c.Assume(err, gospec.Equals, error(nil))
		d2, err := er.Decode(buf)
		c.Assume(err, gospec.Equals, error(nil))

		c.Expect(d1, gospec.Equals, v1)
		c.Expect(d2, gospec.Equals, v2)
		c.Expect(d3, gospec.Equals, v3)
		c.Expect(d4, gospec.Equals, v4)
	})
}
