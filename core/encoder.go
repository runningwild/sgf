package core

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"sort"
)

type typeSorter []reflect.Type

func (ts typeSorter) Len() int { return len(ts) }
func (ts typeSorter) Less(i, j int) bool {
	ni := ts[i].PkgPath() + ":::" + ts[i].Name()
	nj := ts[i].PkgPath() + ":::" + ts[i].Name()
	return ni < nj
}
func (ts typeSorter) Swap(i, j int) {
	ts[i], ts[j] = ts[j], ts[i]
}

type EncoderRegistry struct {
	typeToId  map[reflect.Type]uint32
	types     []reflect.Type
	completed bool
}

func (er *EncoderRegistry) Register(t interface{}) {
	if er.completed {
		panic("Cannot call Register() after Complete().")
	}
	er.types = append(er.types, reflect.TypeOf(t))
}

func (er *EncoderRegistry) Complete() {
	er.completed = true
	sort.Sort(typeSorter(er.types))
	er.typeToId = make(map[reflect.Type]uint32)
	for id, t := range er.types {
		er.typeToId[t] = uint32(id)
	}
}
func (er *EncoderRegistry) writeVal(writer io.Writer, v interface{}) error {
	var err error
	val := reflect.ValueOf(v)
	typ := val.Type()
	kind := typ.Kind()
	switch kind {
	case reflect.Bool:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		fallthrough
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		fallthrough
	case reflect.Complex64:
		fallthrough
	case reflect.Complex128:
		fallthrough
	case reflect.Array:
		err = binary.Write(writer, binary.LittleEndian, v)
	case reflect.Slice:
		err = binary.Write(writer, binary.LittleEndian, uint32(val.Len()))
		if err != nil {
			break
		}
		for i := 0; i < val.Len(); i++ {
			err = er.writeVal(writer, val.Index(i).Interface())
			if err != nil {
				break
			}
		}
	case reflect.String:
		err = binary.Write(writer, binary.LittleEndian, uint32(val.Len()))
		if err != nil {
			break
		}
		err = binary.Write(writer, binary.LittleEndian, []byte(val.String()))
		if err != nil {
			break
		}
	case reflect.Struct:
		n := typ.NumField()
		for i := 0; i < n; i++ {
			err = er.writeVal(writer, val.Field(i).Interface())
			if err != nil {
				break
			}
		}

	default:
		return fmt.Errorf("Can't write %v, not implemented.", reflect.TypeOf(val))
	}
	if err != nil {
		return err
	}
	return nil
}

func (er *EncoderRegistry) readVal(reader io.Reader, v interface{}) error {
	var err error
	val := reflect.ValueOf(v)
	if val.Type().Kind() != reflect.Ptr {
		panic("Can only read into pointers.")
	}
	typ := val.Elem().Type()
	kind := typ.Kind()
	switch kind {
	case reflect.Bool:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		fallthrough
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		fallthrough
	case reflect.Complex64:
		fallthrough
	case reflect.Complex128:
		fallthrough
	case reflect.Array:
		err = binary.Read(reader, binary.LittleEndian, v)
	case reflect.Slice:
		var length uint32
		err = binary.Read(reader, binary.LittleEndian, &length)
		if err != nil {
			break
		}
		dst := reflect.MakeSlice(typ, int(length), int(length))
		for i := 0; i < dst.Len(); i++ {
			err = binary.Read(reader, binary.LittleEndian, dst.Index(i).Addr().Interface())
			if err != nil {
				break
			}
		}
		val.Elem().Set(dst)
	case reflect.String:
		var length uint32
		err = binary.Read(reader, binary.LittleEndian, &length)
		if err != nil {
			break
		}
		buffer := make([]byte, int(length))
		for i := range buffer {
			er.readVal(reader, &buffer[i])
			if err != nil {
				break
			}
		}
		src := reflect.ValueOf(string(buffer)).Convert(typ)
		val.Elem().Set(src)
	case reflect.Struct:
		n := typ.NumField()
		for i := 0; i < n; i++ {
			err = er.readVal(reader, val.Elem().Field(i).Addr().Interface())
			if err != nil {
				break
			}
		}
	default:
		return fmt.Errorf("Can't read %v, not implemented.", kind)
	}
	if err != nil {
		return err
	}
	return nil
}

func (er *EncoderRegistry) Encode(v interface{}) ([]byte, error) {
	if !er.completed {
		return nil, fmt.Errorf("Cannot call Encode() before calling Complete()")
	}
	id, ok := er.typeToId[reflect.TypeOf(v)]
	if !ok {
		panic(fmt.Sprintf("Type %v was not registered.", reflect.TypeOf(v)))
	}
	buf := bytes.NewBuffer(nil)
	err := binary.Write(buf, binary.LittleEndian, id)
	if err != nil {
		return nil, err
	}
	err = er.writeVal(buf, v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (er *EncoderRegistry) Decode(data []byte) (interface{}, error) {
	if !er.completed {
		return nil, fmt.Errorf("Cannot call Decode() before calling Complete()")
	}
	buf := bytes.NewBuffer(data)
	var id uint32
	err := binary.Read(buf, binary.LittleEndian, &id)
	if err != nil {
		return nil, err
	}
	if int(id) >= len(er.types) {
		return nil, fmt.Errorf("Invalid id: %d", id)
	}
	t := er.types[int(id)]
	v := reflect.New(t)
	err = er.readVal(buf, v.Interface())
	if err != nil {
		return nil, err
	}
	return v.Elem().Interface(), nil
}
