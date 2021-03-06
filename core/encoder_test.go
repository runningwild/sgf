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

type encodable7 struct {
	A int
	B uint
}

type encodable8 struct {
	A []int
	B []uint
}

type encodable9 struct {
	A ***e9sub
}
type e9sub struct {
	B int32
}

type encodable10 struct {
	A int32
	b int32
	C string
	d string
}

type encodable11 struct {
	A *e11sub1
	B *e11sub2
	C Barer
	D []int
}
type e11sub1 struct {
	A int
}
type e11sub2 struct{}

type encodable12 struct {
	A LooksLikeAnInt
	B LooksLikeAUInt
}
type LooksLikeAnInt int
type LooksLikeAUInt uint

type encodable13 struct {
	A []*int
	B []*int
}

type encodable14 struct {
	A bool
	B bool
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
func (encodable7) id() int {
	return 7
}
func (encodable8) id() int {
	return 8
}
func (encodable9) id() int {
	return 9
}
func (encodable10) id() int {
	return 10
}
func (encodable11) id() int {
	return 11
}
func (encodable12) id() int {
	return 12
}
func (encodable13) id() int {
	return 13
}
func (encodable14) id() int {
	return 14
}

type ider interface {
	id() int
}

func BasicTypeRegistrySpec(c gospec.Context) {
	c.Specify("Test the coder.", func() {
		var tr core.TypeRegistry
		tr.Register(encodable1{})
		tr.Register(encodable2{})
		tr.Register(encodable3([]byte{}))
		tr.Register(encodable4{})
		tr.Register(encodable5{})
		tr.Register(encodable7{})
		tr.Register(encodable8{})
		tr.Register(encodable9{})
		tr.Register(encodable10{})
		tr.Register(encodable11{})
		tr.Register(encodable12{})
		tr.Register(encodable13{})
		tr.Register(encodable14{})
		tr.Complete()

		c.Specify("Test that encoder handles int32.", func() {
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
		})

		c.Specify("Test that encoder handles byte.", func() {
			e2 := encodable2{14}
			buf := bytes.NewBuffer(nil)
			err := tr.Encode(e2, buf)
			c.Assume(err, gospec.Equals, error(nil))
			dec2, err := tr.Decode(buf)
			c.Assume(err, gospec.Equals, error(nil))
			d2, ok := dec2.(encodable2)
			c.Assume(ok, gospec.Equals, true)
			c.Expect(dec2.(ider).id(), gospec.Equals, 2)
			c.Expect(d2, gospec.Equals, e2)
		})

		c.Specify("Test that encoder handles string.", func() {
			e3 := encodable3("fudgeball")
			buf := bytes.NewBuffer(nil)
			err := tr.Encode(e3, buf)
			c.Assume(err, gospec.Equals, error(nil))
			dec3, err := tr.Decode(buf)
			c.Assume(err, gospec.Equals, error(nil))
			d3, ok := dec3.(encodable3)
			c.Assume(ok, gospec.Equals, true)
			c.Expect(dec3.(ider).id(), gospec.Equals, 3)
			c.Expect(string(d3), gospec.Equals, string(e3))
		})

		c.Specify("Test that encoder handles composite structs.", func() {
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
			buf := bytes.NewBuffer(nil)
			err := tr.Encode(e4, buf)
			c.Assume(err, gospec.Equals, error(nil))
			dec4, err := tr.Decode(buf)
			c.Assume(err, gospec.Equals, error(nil))
			d4, ok := dec4.(encodable4)
			c.Assume(ok, gospec.Equals, true)
			c.Expect(dec4.(ider).id(), gospec.Equals, 4)
			c.Expect(d4, gospec.Equals, e4)
		})

		c.Specify("Test that encoder handles []byte.", func() {
			e5 := encodable5{
				A: []byte("Monkeyball"),
			}
			buf := bytes.NewBuffer(nil)
			err := tr.Encode(e5, buf)
			c.Assume(err, gospec.Equals, error(nil))
			dec5, err := tr.Decode(buf)
			c.Assume(err, gospec.Equals, error(nil))
			d5, ok := dec5.(encodable5)
			c.Assume(ok, gospec.Equals, true)
			c.Expect(dec5.(ider).id(), gospec.Equals, 5)
			c.Expect(string(d5.A), gospec.Equals, string(e5.A))
		})

		c.Specify("Test that encoder handles int and uint.", func() {
			e7 := encodable7{
				A: 1,
				B: 2,
			}
			buf := bytes.NewBuffer(nil)
			err := tr.Encode(e7, buf)
			c.Assume(err, gospec.Equals, error(nil))
			dec7, err := tr.Decode(buf)
			c.Assume(err, gospec.Equals, error(nil))
			d7, ok := dec7.(encodable7)
			c.Assume(ok, gospec.Equals, true)
			c.Expect(dec7.(ider).id(), gospec.Equals, 7)
			c.Expect(string(d7.A), gospec.Equals, string(e7.A))
			c.Expect(string(d7.B), gospec.Equals, string(e7.B))
		})

		c.Specify("Test that encoder handles []int and []uint.", func() {
			e8 := encodable8{
				A: []int{1, 2, 3},
				B: []uint{6, 7, 8},
			}
			buf := bytes.NewBuffer(nil)
			err := tr.Encode(e8, buf)
			c.Assume(err, gospec.Equals, error(nil))
			dec8, err := tr.Decode(buf)
			c.Assume(err, gospec.Equals, error(nil))
			d8, ok := dec8.(encodable8)
			c.Assume(ok, gospec.Equals, true)
			c.Expect(dec8.(ider).id(), gospec.Equals, 8)
			c.Expect(fmt.Sprintf("%v", d8.A), gospec.Equals, fmt.Sprintf("%v", e8.A))
			c.Expect(fmt.Sprintf("%v", d8.B), gospec.Equals, fmt.Sprintf("%v", e8.B))
		})

		c.Specify("Test that encoder handles pointers.", func() {
			sub := e9sub{
				B: 123,
			}
			psub := &sub
			ppsub := &psub
			e9 := encodable9{
				A: &ppsub,
			}
			buf := bytes.NewBuffer(nil)
			err := tr.Encode(e9, buf)
			c.Assume(err, gospec.Equals, error(nil))
			dec9, err := tr.Decode(buf)
			c.Assume(err, gospec.Equals, error(nil))
			d9, ok := dec9.(encodable9)
			c.Assume(ok, gospec.Equals, true)
			c.Expect(dec9.(ider).id(), gospec.Equals, 9)
			c.Assume(d9.A, gospec.Not(gospec.Equals), (*e9sub)(nil))
			c.Expect((**d9.A).B, gospec.Equals, (**e9.A).B)
		})

		c.Specify("Test that encoder can ignore unexported fields.", func() {
			e10 := encodable10{
				A: 10,
				b: 12,
				C: "foo",
				d: "bar",
			}
			buf := bytes.NewBuffer(nil)
			err := tr.Encode(e10, buf)
			c.Assume(err, gospec.Equals, error(nil))
			dec10, err := tr.Decode(buf)
			c.Assume(err, gospec.Equals, error(nil))
			d10, ok := dec10.(encodable10)
			c.Assume(ok, gospec.Equals, true)
			c.Expect(dec10.(ider).id(), gospec.Equals, 10)
			c.Expect(d10.A, gospec.Equals, e10.A)
			c.Expect(d10.b, gospec.Equals, int32(0))
			c.Expect(d10.C, gospec.Equals, e10.C)
			c.Expect(d10.d, gospec.Equals, "")
		})

		c.Specify("Test that encoder can handle nil pointers.", func() {
			e11 := encodable11{}
			buf := bytes.NewBuffer(nil)
			err := tr.Encode(e11, buf)
			c.Assume(err, gospec.Equals, error(nil))
			dec11, err := tr.Decode(buf)
			c.Assume(err, gospec.Equals, error(nil))
			d11, ok := dec11.(encodable11)
			c.Assume(ok, gospec.Equals, true)
			c.Expect(dec11.(ider).id(), gospec.Equals, 11)
			c.Expect(d11.A, gospec.Equals, (*e11sub1)(nil))
			c.Expect(d11.B, gospec.Equals, (*e11sub2)(nil))
			c.Expect(d11.C, gospec.Equals, (Barer)(nil))
		})

		c.Specify("Test that encoder can handle aliased primitives.", func() {
			e12 := encodable12{A: 123, B: 456}
			buf := bytes.NewBuffer(nil)
			err := tr.Encode(e12, buf)
			c.Assume(err, gospec.Equals, error(nil))
			dec12, err := tr.Decode(buf)
			c.Assume(err, gospec.Equals, error(nil))
			d12, ok := dec12.(encodable12)
			c.Assume(ok, gospec.Equals, true)
			c.Expect(dec12.(ider).id(), gospec.Equals, 12)
			c.Expect(d12.A, gospec.Equals, (LooksLikeAnInt)(123))
			c.Expect(d12.B, gospec.Equals, (LooksLikeAUInt)(456))
		})

		c.Specify("Test that encoder can handle slices of nil values.", func() {
			ints := make([]*int, 5)
			for i := range ints {
				n := i
				ints[i] = &n
			}
			e13 := encodable13{A: make([]*int, 5), B: ints}
			buf := bytes.NewBuffer(nil)
			err := tr.Encode(e13, buf)
			c.Assume(err, gospec.Equals, error(nil))
			dec13, err := tr.Decode(buf)
			c.Assume(err, gospec.Equals, error(nil))
			d13, ok := dec13.(encodable13)
			c.Assume(ok, gospec.Equals, true)
			c.Expect(dec13.(ider).id(), gospec.Equals, 13)
			c.Expect(len(d13.A), gospec.Equals, len(e13.A))
			for i := range d13.A {
				c.Expect(d13.A[i], gospec.Equals, e13.A[i])
			}
			c.Expect(len(d13.B), gospec.Equals, len(e13.B))
			for i := range d13.B {
				c.Expect(*d13.B[i], gospec.Equals, *e13.B[i])
			}
		})

		c.Specify("Test that encoder can handle bools.", func() {
			e14 := encodable14{A: true, B: false}
			buf := bytes.NewBuffer(nil)
			err := tr.Encode(e14, buf)
			c.Assume(err, gospec.Equals, error(nil))
			dec14, err := tr.Decode(buf)
			c.Assume(err, gospec.Equals, error(nil))
			d14, ok := dec14.(encodable14)
			c.Assume(ok, gospec.Equals, true)
			c.Expect(dec14.(ider).id(), gospec.Equals, 14)
			c.Expect(d14.A, gospec.Equals, e14.A)
			c.Expect(d14.B, gospec.Equals, e14.B)
		})
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
