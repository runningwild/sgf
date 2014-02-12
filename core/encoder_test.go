package core_test

import (
	"bytes"
	"fmt"
	"github.com/orfjackal/gospec/src/gospec"
	"github.com/runningwild/sgf/core"
	"io"
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

type encodable6 struct {
	Interface Barer
}
type Barer interface {
	Bar() string
}
type BarerImpl1 struct {
	A int32
	B string
}

func (b BarerImpl1) Bar() string {
	return fmt.Sprintf("%d:%s", b.A, b.B)
}

type BarerImpl2 struct {
	A, B int32
}

func (b BarerImpl2) Bar() string {
	return fmt.Sprintf("%d:%d", b.A, b.B)
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

func BasicTypeRegistrySpec(c gospec.Context) {
	c.Specify("Test the codtr.", func() {
		var tr core.TypeRegistry
		tr.Register(encodable1{})
		tr.Register(encodable2{})
		tr.Register(encodable3([]byte{}))
		tr.Register(encodable4{})
		tr.Register(encodable5{})
		tr.Complete()
		e1 := encodable1{1, 2}
		buf := bytes.NewBuffer(nil)
		err := tr.Encode(e1, buf)
		c.Assume(err, gospec.Equals, error(nil))
		dec1, err := tr.Decode(buf)
		c.Assume(err, gospec.Equals, error(nil))
		d1, ok := dec1.(encodable1)
		c.Assume(ok, gospec.Equals, true)
		c.Expect(dec1.(ider).id(), gospec.Equals, 1)
		c.Expect(d1, gospec.Equals, e1)

		e2 := encodable2{14}
		buf = bytes.NewBuffer(nil)
		err = tr.Encode(e2, buf)
		c.Assume(err, gospec.Equals, error(nil))
		dec2, err := tr.Decode(buf)
		c.Assume(err, gospec.Equals, error(nil))
		d2, ok := dec2.(encodable2)
		c.Assume(ok, gospec.Equals, true)
		c.Expect(dec2.(ider).id(), gospec.Equals, 2)
		c.Expect(d2, gospec.Equals, e2)

		e3 := encodable3("fudgeball")
		buf = bytes.NewBuffer(nil)
		err = tr.Encode(e3, buf)
		c.Assume(err, gospec.Equals, error(nil))
		dec3, err := tr.Decode(buf)
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
		err = tr.Encode(e4, buf)
		c.Assume(err, gospec.Equals, error(nil))
		dec4, err := tr.Decode(buf)
		c.Assume(err, gospec.Equals, error(nil))
		d4, ok := dec4.(encodable4)
		c.Assume(ok, gospec.Equals, true)
		c.Expect(dec4.(ider).id(), gospec.Equals, 4)
		c.Expect(d4, gospec.Equals, e4)

		e5 := encodable5{
			A: []byte("Monkeyball"),
		}
		buf = bytes.NewBuffer(nil)
		err = tr.Encode(e5, buf)
		c.Assume(err, gospec.Equals, error(nil))
		dec5, err := tr.Decode(buf)
		c.Assume(err, gospec.Equals, error(nil))
		d5, ok := dec5.(encodable5)
		c.Assume(ok, gospec.Equals, true)
		c.Expect(dec5.(ider).id(), gospec.Equals, 5)
		c.Expect(string(d5.A), gospec.Equals, string(e5.A))
	})
}

func UnorderedTypeRegistrySpec(c gospec.Context) {
	c.Specify("Test that the coder can decode packets out of order.", func() {
		var tr core.TypeRegistry
		tr.Register(encodable1{})
		tr.Register(encodable2{})
		tr.Complete()

		v1 := encodable1{1, 2}
		v2 := encodable2{3}
		v3 := encodable2{4}
		v4 := encodable1{5, 6}

		b1 := bytes.NewBuffer(nil)
		err := tr.Encode(v1, b1)
		c.Assume(err, gospec.Equals, error(nil))

		b2 := bytes.NewBuffer(nil)
		err = tr.Encode(v2, b2)
		c.Assume(err, gospec.Equals, error(nil))

		b3 := bytes.NewBuffer(nil)
		err = tr.Encode(v3, b3)
		c.Assume(err, gospec.Equals, error(nil))

		b4 := bytes.NewBuffer(nil)
		err = tr.Encode(v4, b4)
		c.Assume(err, gospec.Equals, error(nil))

		buf := bytes.NewBuffer(nil)
		buf.Write(b1.Bytes())
		buf.Write(b3.Bytes())
		buf.Write(b4.Bytes())
		buf.Write(b2.Bytes())

		d1, err := tr.Decode(buf)
		c.Assume(err, gospec.Equals, error(nil))
		d3, err := tr.Decode(buf)
		c.Assume(err, gospec.Equals, error(nil))
		d4, err := tr.Decode(buf)
		c.Assume(err, gospec.Equals, error(nil))
		d2, err := tr.Decode(buf)
		c.Assume(err, gospec.Equals, error(nil))

		c.Expect(d1, gospec.Equals, v1)
		c.Expect(d2, gospec.Equals, v2)
		c.Expect(d3, gospec.Equals, v3)
		c.Expect(d4, gospec.Equals, v4)
	})
}

type invasiveWriter struct {
	reg *core.TypeRegistry
	buf io.Writer
}

func (iw *invasiveWriter) Write(data []byte) (int, error) {
	_, err := iw.reg.Decode(bytes.NewBuffer(data))
	if err != nil {
		return 0, err
	}
	return iw.buf.Write(data)
}

func BlockWriteTypeRegistrySpec(c gospec.Context) {
	c.Specify("Test that the coder writes complete values all at once.", func() {
		var tr core.TypeRegistry
		tr.Register(encodable1{})
		tr.Register(encodable2{})
		tr.Register(encodable3([]byte{}))
		tr.Register(encodable4{})
		tr.Register(encodable5{})
		tr.Register(encodable6{})
		tr.Register(BarerImpl1{})
		tr.Complete()

		var iw = invasiveWriter{
			reg: &tr,
			buf: bytes.NewBuffer(nil),
		}

		e1 := encodable1{1, 2}
		err := tr.Encode(e1, &iw)
		c.Assume(err, gospec.Equals, error(nil))

		e2 := encodable2{14}
		err = tr.Encode(e2, &iw)
		c.Assume(err, gospec.Equals, error(nil))

		e3 := encodable3("fudgeball")
		err = tr.Encode(e3, &iw)
		c.Assume(err, gospec.Equals, error(nil))

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
		err = tr.Encode(e4, &iw)
		c.Assume(err, gospec.Equals, error(nil))

		e5 := encodable5{
			A: []byte("Monkeyball"),
		}
		err = tr.Encode(e5, &iw)
		c.Assume(err, gospec.Equals, error(nil))

		e6 := encodable6{
			Interface: BarerImpl1{10, "foo"},
		}
		err = tr.Encode(e6, &iw)
		c.Assume(err, gospec.Equals, error(nil))
	})
}

func InterfaceTypeRegistrySpec(c gospec.Context) {
	c.Specify("Test that types containing interfaces are properly encoded.", func() {
		var tr core.TypeRegistry
		tr.Register(encodable6{})
		tr.Register(BarerImpl1{})
		tr.Register(BarerImpl2{})
		tr.Complete()

		v1 := encodable6{
			Interface: BarerImpl1{10, "foo"},
		}
		v2 := encodable6{
			Interface: BarerImpl2{3, 4},
		}

		b := bytes.NewBuffer(nil)
		err := tr.Encode(v1, b)
		c.Assume(err, gospec.Equals, error(nil))
		err = tr.Encode(v2, b)
		c.Assume(err, gospec.Equals, error(nil))

		d1, err := tr.Decode(b)
		c.Assume(err, gospec.Equals, error(nil))
		d2, err := tr.Decode(b)
		c.Assume(err, gospec.Equals, error(nil))

		c1, ok := d1.(encodable6)
		c.Assume(ok, gospec.Equals, true)
		c.Expect(c1.Interface.Bar(), gospec.Equals, "10:foo")
		c2, ok := d2.(encodable6)
		c.Assume(ok, gospec.Equals, true)
		c.Expect(c2.Interface.Bar(), gospec.Equals, "3:4")
	})
}
