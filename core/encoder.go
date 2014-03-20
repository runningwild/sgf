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

type TypeRegistry struct {
	typeToId  map[reflect.Type]uint32
	types     []reflect.Type
	completed bool
}

func (tr *TypeRegistry) Register(t interface{}) {
	if tr.completed {
		panic("Cannot call Register() after Complete().")
	}
	tr.types = append(tr.types, reflect.TypeOf(t))
}

func (tr *TypeRegistry) Complete() {
	tr.completed = true
	sort.Sort(typeSorter(tr.types))
	tr.typeToId = make(map[reflect.Type]uint32)
	for id, t := range tr.types {
		tr.typeToId[t] = uint32(id)
	}
}
func (tr *TypeRegistry) writeVal(writer io.Writer, v interface{}) error {
	fmt.Printf("writeVal: %T %v\n", v, v)
	var err error
	val := reflect.ValueOf(v)
	typ := val.Type()
	kind := typ.Kind()
	for kind == reflect.Ptr {
		val = val.Elem()
		typ = val.Type()
		kind = typ.Kind()
	}
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
	case reflect.Int:
		err = binary.Write(writer, binary.LittleEndian, val.Int())
	case reflect.Uint:
		err = binary.Write(writer, binary.LittleEndian, val.Uint())
	case reflect.Slice:
		err = tr.writeVal(writer, uint32(val.Len()))
		if err != nil {
			break
		}
		for i := 0; i < val.Len(); i++ {
			err = tr.writeVal(writer, val.Index(i).Interface())
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
	case reflect.Struct:
		n := typ.NumField()
		for i := 0; i < n; i++ {
			if typ.Field(i).PkgPath != "" {
				// This indicates an unexported field.
				continue
			}
			kind := typ.Field(i).Type.Kind()
			if kind == reflect.Interface || kind == reflect.Map || kind == reflect.Ptr || kind == reflect.Slice {
				if val.Field(i).IsNil() {
					binary.Write(writer, binary.LittleEndian, byte(1))
					continue
				} else {
					binary.Write(writer, binary.LittleEndian, byte(0))
				}
			}
			if typ.Field(i).Type.Kind() == reflect.Interface {
				err = tr.Encode(val.Field(i).Interface(), writer)
			} else {
				err = tr.writeVal(writer, val.Field(i).Interface())
			}
			if err != nil {
				break
			}
		}
	case reflect.Ptr:
		return fmt.Errorf("FUUUKKK")
	default:
		return fmt.Errorf("Can't write %v, not implemented.", typ)
	}
	if err != nil {
		return err
	}
	return nil
}

func (tr *TypeRegistry) readVal(reader io.Reader, v interface{}) error {
	var err error
	val := reflect.ValueOf(v)
	if val.Type().Kind() != reflect.Ptr {
		panic("Can only read into pointers.")
	}
	typ := val.Elem().Type()
	kind := typ.Kind()
	deepness := 0
	for kind == reflect.Ptr {
		val.Elem().Set(reflect.New(typ.Elem()))
		val = val.Elem()
		typ = val.Elem().Type()
		kind = typ.Kind()
		deepness++
	}
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
	case reflect.Int:
		var n int64
		err = binary.Read(reader, binary.LittleEndian, &n)
		if err != nil {
			break
		}
		val.Elem().Set(reflect.ValueOf(int(n)))
	case reflect.Uint:
		var n uint64
		err = binary.Read(reader, binary.LittleEndian, &n)
		if err != nil {
			break
		}
		val.Elem().Set(reflect.ValueOf(uint(n)))
	case reflect.Slice:
		var length uint32
		err = binary.Read(reader, binary.LittleEndian, &length)
		if err != nil {
			break
		}
		src := reflect.MakeSlice(typ, int(length), int(length))
		for i := 0; i < src.Len(); i++ {
			err = tr.readVal(reader, src.Index(i).Addr().Interface())
			if err != nil {
				break
			}
		}
		val.Elem().Set(src)
	case reflect.String:
		var length uint32
		err = binary.Read(reader, binary.LittleEndian, &length)
		if err != nil {
			break
		}
		buffer := make([]byte, int(length))
		for i := range buffer {
			tr.readVal(reader, &buffer[i])
			if err != nil {
				break
			}
		}
		src := reflect.ValueOf(string(buffer)).Convert(typ)
		val.Elem().Set(src)
	case reflect.Struct:
		n := typ.NumField()
		for i := 0; i < n; i++ {
			if typ.Field(i).PkgPath != "" {
				// This indicates an unexported field.
				continue
			}
			kind := typ.Field(i).Type.Kind()
			if kind == reflect.Interface || kind == reflect.Map || kind == reflect.Ptr || kind == reflect.Slice {
				var isNil byte
				binary.Read(reader, binary.LittleEndian, &isNil)
				if isNil == 1 {
					continue
				}
			}
			err = tr.readVal(reader, val.Elem().Field(i).Addr().Interface())
			if err != nil {
				break
			}
		}
	case reflect.Interface:
		var src interface{}
		src, err = tr.Decode(reader)
		val.Elem().Set(reflect.ValueOf(src))
	default:
		return fmt.Errorf("Can't read %v, not implemented.", kind)
	}
	if err != nil {
		return err
	}
	return nil
}

func (tr *TypeRegistry) Encode(v interface{}, writer io.Writer) error {
	if !tr.completed {
		return fmt.Errorf("Cannot call Encode() before calling Complete()")
	}
	id, ok := tr.typeToId[reflect.TypeOf(v)]
	if !ok {
		panic(fmt.Sprintf("Type %v was not registered.", reflect.TypeOf(v)))
	}
	tmp := bytes.NewBuffer(nil)
	err := binary.Write(tmp, binary.LittleEndian, id)
	if err != nil {
		return err
	}
	err = tr.writeVal(tmp, v)
	if err != nil {
		return err
	}
	_, err = writer.Write(tmp.Bytes())
	if err != nil {
		return err
	}
	return nil
}

func (tr *TypeRegistry) Decode(reader io.Reader) (interface{}, error) {
	if !tr.completed {
		return nil, fmt.Errorf("Cannot call Decode() before calling Complete()")
	}
	var id uint32
	err := binary.Read(reader, binary.LittleEndian, &id)
	if err != nil {
		return nil, err
	}
	if int(id) >= len(tr.types) {
		return nil, fmt.Errorf("TypeRegistry: Invalid id: %d", id)
	}
	t := tr.types[int(id)]
	v := reflect.New(t)
	err = tr.readVal(reader, v.Interface())
	if err != nil {
		return nil, err
	}
	return v.Elem().Interface(), nil
}
