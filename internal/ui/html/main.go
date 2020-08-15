package html

import (
	"html/template"
	"reflect"
	"strconv"
	"strings"
	"time"
)

//Template the page template
var Template = template.New("").Funcs(template.FuncMap{
	"Ns":            func(i uint64) string { return time.Duration(i).String() },
	"FloatFormat32": func(i float32) string { return strconv.FormatFloat(float64(i), 'f', 4, 64) },
	"FloatFormat64": func(i float64) string { return strconv.FormatFloat(float64(i), 'f', 3, 64) },
	"Bytes":         func(i uint64) string { return string(toBytes(float64(i))) },
	"Bytes32":       func(i uint32) string { return string(toBytes(float64(i))) },
	"ToString":      func(v reflect.Value) string { return v.MethodByName("String").Call(nil)[0].String() },
	"TryGetTime": func(t1, t2 time.Time, d time.Duration) string {
		if !t1.IsZero() {
			return t1.String()
		} else if !t2.IsZero() && d != 0 {
			return t2.Add(-d).String()
		}
		return "Unknown"
	},
	"UnRingBuffer": func(v [256]uint64, i uint32) string {
		c := (i + 255) % 256
		w := append(append([]uint64{}, v[:c+1]...), v[c+1:]...)
		ret := ""
		for i, x := range w {
			if x == 0 {
				break
			}
			if i%50 == 0 {
				ret += "\n" + strings.Repeat(" ", 11)
			} else {
				ret += "  "
			}
			ret += time.Duration(x).String()
		}
		return ret
	},
	"Time": func(t uint64) string {
		return time.Unix(0, int64(t)).String()
	},
	"Sub": func(i, i2 uint64) uint64 { return i - i2 },
	"Div": func(i uint64, i2 uint32) uint64 {
		return i / uint64(i2)
	},
},
)

func toBytes(v float64) []byte {
	var b []byte
	const unit = 1000.0
	var e string
	var d float64
	switch {
	case v < unit:
		e = "B"
		d = 1
		b = strconv.AppendFloat(b, v/d, 'f', 0, 64)
		b = append(b, e...)
		return b
	case v < unit*unit:
		e = "kB"
		d = unit
	case v < unit*unit*unit:
		e = "MB"
		d = unit * unit
	case v < unit*unit*unit*unit:
		e = "GB"
		d = unit * unit * unit
	default:
		e = "TB"
		d = unit * unit * unit * unit
	}
	b = strconv.AppendFloat(b, v/d, 'f', 2, 64)
	b = append(b, e...)
	return b
}
