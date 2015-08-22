package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

type _escLocalFS struct{}

var _escLocal _escLocalFS

type _escStaticFS struct{}

var _escStatic _escStaticFS

type _escDir struct {
	fs   http.FileSystem
	name string
}

type _escFile struct {
	compressed string
	size       int64
	local      string
	isDir      bool

	data []byte
	once sync.Once
	name string
}

func (_escLocalFS) Open(name string) (http.File, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	return os.Open(f.local)
}

func (_escStaticFS) prepare(name string) (*_escFile, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	var err error
	f.once.Do(func() {
		f.name = path.Base(name)
		if f.size == 0 {
			return
		}
		var gr *gzip.Reader
		b64 := base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(f.compressed))
		gr, err = gzip.NewReader(b64)
		if err != nil {
			return
		}
		f.data, err = ioutil.ReadAll(gr)
	})
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (fs _escStaticFS) Open(name string) (http.File, error) {
	f, err := fs.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.File()
}

func (dir _escDir) Open(name string) (http.File, error) {
	return dir.fs.Open(dir.name + name)
}

func (f *_escFile) File() (http.File, error) {
	type httpFile struct {
		*bytes.Reader
		*_escFile
	}
	return &httpFile{
		Reader:   bytes.NewReader(f.data),
		_escFile: f,
	}, nil
}

func (f *_escFile) Close() error {
	return nil
}

func (f *_escFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

func (f *_escFile) Stat() (os.FileInfo, error) {
	return f, nil
}

func (f *_escFile) Name() string {
	return f.name
}

func (f *_escFile) Size() int64 {
	return f.size
}

func (f *_escFile) Mode() os.FileMode {
	return 0
}

func (f *_escFile) ModTime() time.Time {
	return time.Time{}
}

func (f *_escFile) IsDir() bool {
	return f.isDir
}

func (f *_escFile) Sys() interface{} {
	return f
}

// FS returns a http.Filesystem for the embedded assets. If useLocal is true,
// the filesystem's contents are instead used.
func FS(useLocal bool) http.FileSystem {
	if useLocal {
		return _escLocal
	}
	return _escStatic
}

// Dir returns a http.Filesystem for the embedded assets on a given prefix dir.
// If useLocal is true, the filesystem's contents are instead used.
func Dir(useLocal bool, name string) http.FileSystem {
	if useLocal {
		return _escDir{fs: _escLocal, name: name}
	}
	return _escDir{fs: _escStatic, name: name}
}

// FSByte returns the named file from the embedded assets. If useLocal is
// true, the filesystem's contents are instead used.
func FSByte(useLocal bool, name string) ([]byte, error) {
	if useLocal {
		f, err := _escLocal.Open(name)
		if err != nil {
			return nil, err
		}
		return ioutil.ReadAll(f)
	}
	f, err := _escStatic.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.data, nil
}

// FSMustByte is the same as FSByte, but panics if name is not present.
func FSMustByte(useLocal bool, name string) []byte {
	b, err := FSByte(useLocal, name)
	if err != nil {
		panic(err)
	}
	return b
}

// FSString is the string version of FSByte.
func FSString(useLocal bool, name string) (string, error) {
	b, err := FSByte(useLocal, name)
	return string(b), err
}

// FSMustString is the string version of FSMustByte.
func FSMustString(useLocal bool, name string) string {
	return string(FSMustByte(useLocal, name))
}

var _escData = map[string]*_escFile{

	"/bam.css": {
		local: "public/bam.css",
		size:  2187,
		compressed: `
H4sIAAAJbogA/5RV207rOBR971dYVPMyaqKkpC1NpZGYXjRP8w+uvZNauHZkO0AP4t+PHcdJmhZ0gBey
ve9rrc3f6GOC0FG+R5r9YqLM7d+KgoqsaTP5nBwlvXgXTF5KJWtBIyK5VDmaAimSIt3Yx0IKExX4zPgl
Rw//AX8FwwhG/0MND7Pue/asGOYzjYWONChWdLG2OOQozaqm6CltSjYvb8DKk8nRKklG3vN4AWfnPiXW
iJkA1YS9MWpOOVonf7mAM36PWstqkbj8CFWYUjtrpHzqdHFt5lAMrGesSiaCL66NHFi9qzd+TmreNMCZ
tj2aC7dNCilgkDtHSR+dI1vCGT4nnKEPFPaKMd4gly12mdA/yL0Oc6TJsLUmzaZB0SFnn21WLTmjaEoI
6V8ihSmrdY7aNQ8LxKUCELaJ1tUPtugzpfhI1mTcV3wBzuXbN3FFSrKkuIlTQL8JglVGHm+LlQpfvok6
Uhu0aqMwMUwK3W+PMl1xbPnJBGcjUBbdQkIYO5dDMqVLv+JXUI7IPMKclXbzZ0YpBxeLZwjnr0wz4yaz
rgHNJFllu7ULNvBuIgpEKuyK9OToteVtSEEF2CBNlOQcJfbXKCuaCisQpqkWG2Y49DJpJeEEMRLOwgvH
zRydWlsap0PCH6Ux8tytoWNDV6UbpmdBj/2tV495i/StS49wwPXWp8czrmrOvQTte8EltkM0n+51EoNS
sjlZI50EYdzersN8t9/tB9oID/t/t9tds51geV5nWTa/L7ArWL7S2aC902PTYRBuMmahvwZTDViRUzdQ
IGHiT1pAMVte3S1/FzqqDu/qkzeNOPDofr5q++452W63o/7yQpJat/9Erha5XD4f9utm/hOjcK1Bz3z7
pA02tY5ULYQdYYaCQRtZVa2SbvBsj2+g7Tz5quE/gKeVZatnYj1B3ensStKtCu4za3c4JLunO8wKUX3u
4ZAjYfyQtSHK5vYXLDrWdjniHn06EI5ckpcfEejqsGT+sNxb3w8o1Yrcdv47AAD//73tIoCLCAAA
`,
	},

	"/bam.js": {
		local: "public/bam.js",
		size:  342,
		compressed: `
H4sIAAAJbogA/0yQQW8CIRCF7/wKbrCxEu+0h5r00KS3Hjd7QBh3SRAsDEbT+N87iJqeIG/me+/ByWRe
wGS7bNOZv3GXbD1ARDUDfgRo1+3l00nRl9a7dBaDZg0zx2P5T/xUyJdvCGAx5fcQpAh+dAbNmjanRrF9
jRZ9ivdIOfBfxnkzi8nBC4/mAJqUfcpcNtlTwEbT8XqLUwHijAsJq1Vn+Y2krTYe/aS7Rj6ktZEyiNnv
KkIZxaONmNTJhAr66aBsMKV8+YIK0zwHkGLxDkTvpO59nz/V8YFqbYZmcmVXxh6P0uwvAAD//0JT8xxW
AQAA
`,
	},

	"/images/info.png": {
		local: "public/images/info.png",
		size:  566,
		compressed: `
H4sIAAAJbogA/wA2Asn9iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAYAAABzenr0AAAABHNCSVQICAgI
fAhkiAAAAAlwSFlzAAALEwAACxMBAJqcGAAAAdhJREFUWIXF1z9rFEEYBvDfrQeaSLwPYKWexkJIsLJO
qfZC8IPkC4inIdiKKDaKWIuxSmlh50m0jI1IiChE8kc9QYvZI5v1crezt+c+8MIyM+/7zMzOvO8zDcVx
CtexgDmcxUza9x0f0cUaXqZtlWAWj7GPPwVtD49wfhziaazgdwRx3npYxlQs+QW8H4M4b+/QLkp+GV8q
JO/bFuaLrHwS5NlJHLkTJ1W77cN+x8AzsVIi2C3har6I9LubJ58Vf9qfZ/yP41uEb0/6K5I0wBKODdqW
IdjJfPfwM8K3mXKClrgk07cfWMQZ3CvhvyvNpIslnKuyG00ht8fgPl4PaO/gdGSshaZQWGLwCevCnf6c
aV8qMYE5+Krc9nVywdZLxNhKHJTUOtBKRo+ZLBIVCocS2E4EJVMXNhJBRtWFbiJouLqwlggCcr8G8j2s
9g/hsxom8BQ7/WvYEcrx/0IPd/KNy4plr5sjgj8pEON2f3Aj4ziNN7g0guAVPgzpv4aLQ/q7uCKU83/Q
ForMpMrvpvCiGor5CU1iU0TlbQvqtSryt0VWnseUoF57YxD/Eg7ciVjyLNp4KGi4osS7eIBzo4I3Rg3I
YAZXHX6et9K+bWw4eJ6vOqyaj8RfVbWi7FM6OQ8AAAAASUVORK5CYIIBAAD//1AhoFQ2AgAA
`,
	},

	"/images/share.png": {
		local: "public/images/share.png",
		size:  549,
		compressed: `
H4sIAAAJbogA/wAlAtr9iVBORw0KGgoAAAANSUhEUgAAABwAAAAgCAYAAAABtRhCAAAABHNCSVQICAgI
fAhkiAAAAAlwSFlzAAALEwAACxMBAJqcGAAAAcdJREFUSImt10+ITlEYx/HPYIYYZpJkIwuJ/CkmsiOy
ZmFnQRbKxlZElCRlPRtWUiRFLGyUlUKiZjFjWAhrf6b8K/PyWpx76/V6z+vce8+vnm7d89zne89znvPc
c8mrhTiNabTwGbexNTMHLMME2j2shcO5gfcjsNJmsTkXbMt/YKXdmJMJuCvVLxdwJLNfX23AM2kpnWoC
WoRLQjGkwNo4Vxe2H+8rgNp4gyVVQavFy79VjLV6jE0Uz/6jxdhTzGBjx/35OIsfEdhTjBW+a3AG1zBe
xJrbDRrExR4Bn+MIXkdAn3AUlap9AHciAfvZVSyvAip1sCJoEjvqgEo9SgR9xXEh/Y30JRF4oimIsNjz
cgSqosfSZvgdpzDUFHgoEVjaNHY3AQ4Ix4Cq2+I6VtSFDuICvnUFfSIcDV5GoDM4pkcnSdWwkK59WNtx
fwgne7xQZ0faVviux3ncxBUc0GArrcLdCPQXHhTX7rEpocfW1l68jYBj9g6jdXP/CpeFfbxd2hqO4HdN
3l9aJxRX0paqXV0d+oCV2JngO5zr1DaT6pcL+DCzX5Lu6b9+P7EpJ3ApXkRgs8KHPrsWCB1pUpjRR9wS
/j3AHyoOFAHGnFsmAAAAAElFTkSuQmCCAQAA//+bINtuJQIAAA==
`,
	},

	"/images/shared.png": {
		local: "public/images/shared.png",
		size:  1455,
		compressed: `
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
`,
	},

	"/images/start.png": {
		local: "public/images/start.png",
		size:  334,
		compressed: `
H4sIAAAJbogA/+oM8HPn5ZLiYmBg4PX0cAkC0rJArMDBBiTflRfXACmWYifPEA4gqOFI6QDyOQs8IosZ
GLiFQZiRYdYcCaDgB08XxxCPzr3XFXWOBoqwP8hPSw4W8byc4bLKR+3yyRSjmQ6Kk5+KZ24JD4gLvPdt
2/bt2gXmyXLP/tel/376/t27d3OqK1YXn+G8rrzrxtEIId33mYmKvWJfNvR/VIkqbl9qEnXhuKLa5Iyz
zGGXMg4yHJ/htcg0d82yi7OCfZZfbEpfs/ziLO8zSy428ZldzjhrmnY54yD3NiCr1z/Z2PCY1r+lbqV7
ow/7XTapUv0xwerGD85sob0COXz+Eb+5v8V/cdnwcDur+20G9ybnP3KnfRP2O/62OOAy+YHBF7Fu3+T7
C23fcNZti54hr/pN8G2Se0GjlXCMaDzjxu+FufUZMfJAvzN4uvq5rHNKaAIEAAD//1HixQJOAQAA
`,
	},

	"/images/stop.png": {
		local: "public/images/stop.png",
		size:  140,
		compressed: `
H4sIAAAJbogA/+oM8HPn5ZLiYmBg4PX0cAkC0gogzMEGJIurqr4AKZZiJ88QDiCo4UjpAPI5CzwiixkY
uIVBmJFh1hwJoKCep4tjSETr23OGjAwGPA0GUf/nmUvWNjEaXX29Ye8cZgYUEFKrKMRof7XzQT2I5+nq
57LOKaEJEAAA///mDrgYjAAAAA==
`,
	},

	"/": {
		isDir: true,
		local: "public",
	},

	"/images": {
		isDir: true,
		local: "public/images",
	},
}
