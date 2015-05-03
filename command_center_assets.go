package main

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

type _esc_localFS struct{}

var _esc_local _esc_localFS

type _esc_staticFS struct{}

var _esc_static _esc_staticFS

type _esc_file struct {
	compressed string
	size       int64
	local      string
	isDir      bool

	data []byte
	once sync.Once
	name string
}

func (_esc_localFS) Open(name string) (http.File, error) {
	f, present := _esc_data[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	return os.Open(f.local)
}

func (_esc_staticFS) Open(name string) (http.File, error) {
	f, present := _esc_data[path.Clean(name)]
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
		gr, err = gzip.NewReader(bytes.NewBufferString(f.compressed))
		if err != nil {
			return
		}
		f.data, err = ioutil.ReadAll(gr)
	})
	if err != nil {
		return nil, err
	}
	return f.File()
}

func (f *_esc_file) File() (http.File, error) {
	type httpFile struct {
		*bytes.Reader
		*_esc_file
	}
	return &httpFile{
		Reader:    bytes.NewReader(f.data),
		_esc_file: f,
	}, nil
}

func (f *_esc_file) Close() error {
	return nil
}

func (f *_esc_file) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

func (f *_esc_file) Stat() (os.FileInfo, error) {
	return f, nil
}

func (f *_esc_file) Name() string {
	return f.name
}

func (f *_esc_file) Size() int64 {
	return f.size
}

func (f *_esc_file) Mode() os.FileMode {
	return 0
}

func (f *_esc_file) ModTime() time.Time {
	return time.Time{}
}

func (f *_esc_file) IsDir() bool {
	return f.isDir
}

func (f *_esc_file) Sys() interface{} {
	return f
}

// FS returns a http.Filesystem for the embedded assets. If useLocal is true,
// the filesystem's contents are instead used.
func FS(useLocal bool) http.FileSystem {
	if useLocal {
		return _esc_local
	}
	return _esc_static
}

var _esc_data = map[string]*_esc_file{

	"/bam.css": {
		local: "public/bam.css",
		size:  1630,
		compressed: "\x1f\x8b\b\x00\x00\tn\x88\x00\xff|U\xdbn\xe2<\x10\xbe\xe7)\xac\xa2\xff\x8eDN\x1b\xe0'\x95V\xea\x02\xd5^\xed;\f\xf6$\xb1։\x91mZ؊w_;Ή\xd0\xd3U\xc6\xf3\xcd\xf1\xfb\x86\x83\xe2\x17\xf21#\xe4\x00\xecO\xa1թ\xe6\x11SR\xe9\x8c̑\xe54O\x9e\xddc\xaej\x1b\xe5P\ty\xc9\xc8\xc3/" +
			"\x94oh\x05\x03\xf2\x1bO\xf8\xb0\xe8\xbf\x17/Z\x80\\\x18\xa8MdP\x8b\xbc\xc7\x1a\xf1\x173\x92\xa4\xc7\xf3\xf3\xec:+\x93&e\xf3\xf2\x8e\xa2(mF֔N\xbc\x1f\xe3%V\xde}Μ\x11D\x8d\xba\x81\xbd\vnˌl\xe8\u007f\x1eP\xc19j-\xeb%\xf5\xf1\t9\x02\xe7\xa2.\"\x1dB'\xcb[\xb3\xc4|d\xad@" +
			"\x17\xa2\xee|\xe1d\xd5\xc8\x1a\\\x83\xf1:;ɦ\x00)\x8c\xab\xd1^\xa4+\xb2V5\x8ebg\x84\x0e茸\x14\xdep\x9dIA>H7W\x00x&>Z\xec#\x91\x1fĿ\x8ec$t\\Z\x13\xc6\u007f\x1e\x94\xe6\xe8\x02$.\xaaQRp2g\x8c\r/\x91\x06.N&#\xed\x98\xc7\t\xe2B#֮\x88\xd654" +
			"\xb6\x1c\"%p`\x1b6\xad+\xbe\xa0\x94\xea\xfd\x1b\\\x9e\xb0\x94\xe6w8\x8d\xfc\x1b\x10\xaeS\xf6t\x9f\xac\xd0p\xf9\x06u\xe0\x0e\xb4nQ\xc0\xacP\xb5q\xee\xb9T\xe0\x1c\x9b\x05N\x1e\xfb\xd1ra\x8e\x12\x1cyE-\xc5dc\xcb~Z\x1dLTŘi\xc9*\xcc\xff\r\xb5g\xb9\x8c@\x8a\u00ad\xa5\x12\x9cK\xf4XX\x10" +
			"\xc8ބ\x11ַ\xed\\\xbbUS\xbaNw\x1b\x0f\xb6x\xb6\x11G\xa64\xf8$\x03s\x06\xe1\x05\x1b\xd1xD\xb0\xc40\xad\xa4$\xd4\xfd[\xed\x14u\x04\x8d\xb5m\xb2\xc5VX\x89\x83\x86Z\xbdx\xb5LT\xb5\f\xaa\xf2=GekK\xe2d\xac\x86\x83\xb2VU\xfd\x18z\xaa\xf4Y\xfaf\x06\x8a\fĸ\xf7\x1a\b\xd1\xd2\xe0\xde" +
			"eX\u007f\xb7\xf4{\x9faٳ\x18\xb5V\xda\x15z\x9eȤ\xd3\xc5\xfd\xe9z}\xdc\xedw\xfb\x914\xba\x87\xfd\xcf\xedv\xd7\xf4\xdfY^6i\x9a>~\xae\xaf\x9b\xc1\u007f%\xb3Qy\xe5SSa\xa7[:\xe5Y8\x06s\x83\xa0Y\xd97\xd4ь\x86\x8b\xd6\xed)]ݜ\xadp\x16z2\x8e\xcf\xea\xff\xc14\xd9\xf2\x93\xff\xfb" +
			"\xaa\xecO\xaf\xc9v\xbb\x9dԗ劝L\xf8\x81\xb8\x1d\xe4j\xf5\xf2\xba\xdf4\xfd\x97\x82\xe3\xad\xca\x02\xb7\xaf\xb3\u007f\x01\x00\x00\xff\xff4\x11\x9e\xb6^\x06\x00\x00",
	},

	"/bam.js": {
		local: "public/bam.js",
		size:  342,
		compressed: "\x1f\x8b\b\x00\x00\tn\x88\x00\xffL\x90Ao\x02!\x10\x85\xef\xfc\nn\xb0\xb1\x12ﴇ\x9a\xf4Ф\xb7\x1e7{@\x18wI\x10,\fF\xd3\xf8\xdf;\x88\x9a\x9e o\xe6{\xef\xc1\xc9d^\xc0d\xbblә\xbfq\x97l=@D5\x03~\x04h\xd7\xed\xe5\xd3Iїֻt\x16\x83f\r3\xc7c\xf9O\xfcT\xc8" +
			"\x97o\b`1\xe5\xf7\x10\xa4\b~t\x06͚6\xa7F\xb1}\x8d\x16}\x8a\xf7H9\xf0_\xc6y3\x8b\xc9\xc1\v\x8f\xe6\x00\x9a\x94}\xca\\6\xd9S\xc0F\xd3\xf1z\x8bS\x01\xe2\x8c\v\t\xabUg\xf9\x8d\xa4\xad6\x1e\xfd\xa4\xbbF>\xa4\xb5\x912\x88\xd9\xef*B\x19ţ\x8d\x98\xd4Ʉ\n\xfa\xe9\xa0l0\xa5|\xf9\x82" +
			"\n\xd3<\a\x90b\xf1\x0eD\xef\xa4\xee}\x9f?\xd5\xf1\x81jm\x86freW\xc6\x1e\x8f\xd2\xec/\x00\x00\xff\xffBS\xf3\x1cV\x01\x00\x00",
	},

	"/error.html": {
		local: "public/error.html",
		size:  318,
		compressed: "\x1f\x8b\b\x00\x00\tn\x88\x00\xffD\x90\xc1N\xc4 \x10\x86\xef>\xc58\x17O\xcan6^\f\x90\xac\x89G\x1f\x82\u0098\x12)4\xccĴo/l\xdb\xdd\xdb\xcf\f\xff\xc7\x17\xf4(S\xb2O\x00z$\x17zhQ\xa2$\xb2\x9f\xd7\xefg\xad\xb6\xbc\xcdS̿P)\x19dY\x13\xf1H$\b\xb2\xcedPh\x11\xe5\x99\x11\xc6" +
			"J?\x06\a7\xbd\xf5㍭\x0e\xb8\x1eJXwZ\x88\u007f\x10\x83A_\xb2\xb8\x98\xa9\xe2\xb6\xe8.g\v\xda\xed$\xdcM\x9c\xed\xa0\xf3\xfdR\xef\xfb\xe4\x98\rR\xad\xa5\xbe\x0ee\xb9#:\xe4b\xbf\xda\x1c\xdeO\xa7ֻ<6\xd7yN\xd1;\x89%C(\xc4\xf9E\x80\x96\xc8\xf2\x01\xceOt\xf0U{`W=\xa2V\x9b\u007f" +
			"\xe3ݾ\xed?\x00\x00\xff\xff\xfdq\x0e\xd9>\x01\x00\x00",
	},

	"/images/share.png": {
		local: "public/images/share.png",
		size:  499,
		compressed: "\x1f\x8b\b\x00\x00\tn\x88\x00\xff\x00\xf3\x01\f\xfe\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00 \x00\x00\x00 \b\x06\x00\x00\x00szz\xf4\x00\x00\x00\x04sBIT\b\b\b\b|\bd\x88\x00\x00\x00\tpHYs\x00\x00\v\x13\x00\x00\v\x13\x01\x00\x9a\x9c\x18\x00\x00\x01\x95IDATX\x85\xedֿKV" +
			"Q\x1c\xc7\xf1W\xb9\x06\x05Yj6\x14\x04\x19\x85\xa2`c\xb5\x14\xd1R\r\x85B\xff\x83\x8bs\x114\xa4C\x83\x9b\xa3K\x10\x92\x83cC\x83\bB\xd1\x12mn.\x92\xfd\xa0\xd2\n\u007fq\x1d\xee\xbdq9\x8f>\xf7y\xe4\xdc\a\x82\xe7\r\xdf\xe1\x1e\xee\xf9\xbe?\xf7p\xee=\x976m\xaa\xe12\xe6\xf0-\xab\u05f8\xd4*\xf906" +
			"\x90\x04\xf5\v\x83\xad\b\xf0a\x1fy^KU\xcb\xcfԑ\xe7u\xaa8\xe1h\xe4\x00\xdd\r\xdcs,\xb2\x13ta\x12\xeb\xea?\xfd\x17t\xc4\x14\x9f\xc5\x14\xfe\x94\x88\xf3\x1ak\xa6y\x87\xf4u\xba\xa26\xf5yLc3\x10,\xe0\x16\x9eb\xab0\xbe\x89\xc7\xcd\xc8\x1fa\xb5\xd0`\x15\xa3\xb8\x88\x19l\a\xe27\xb8\x16\xf4\xe8\xc2}\xdc" +
			"\x13l\xbc2F\x1c\xbc\x84\xbb\xc1\xf5<\xae6Ӽ\x8c#X\xa9\x13 \x0f1\x8b\x81\x98\xe2\x9cs%\xf2\x04\xd7c\n\xc3\xef@\xd2\xc0\x9c\xe31\x03\xecǲ\xf2Ux\x8b\x1bU\x05\xb8\xa3v\xb3%\xd8\xc1\xdf`l\x017\xab\bq\x1b\x9f\n\xa2\x8f\x99\xa8\a/\xf0;\b\xb2\x94\x05/ҋ\x87x\x90\xcd;\x14\x9dY\x85\x9cƄ" +
			"\xdaO\xef{\xdc\xc5s\xe9\x8a\xe5\xe3\xdbxv\xd8\x10\xf58\x995\xfe\xa9|\xdf$\x18\xaf\"\x04\x9c\xc0\x13|/\t\xf0U\xe4\xc3(d\xa0$@\"=G\xfe\x11\xfb\u007f`\xad\x81{\xd6#;kx\xe7\xe0\xa7_\xacZ\x0eC\xd2\x1f\xd0P\xfe\x03\xfd\xad\b\x00}x\x85\xcfң\xfc%.\xb4J\xde\xe6\xffb\x0f=\x8d\xbc\xd2q\x10" +
			"\v>\x00\x00\x00\x00IEND\xaeB`\x82\x01\x00\x00\xff\xff\xed\"\xfc`\xf3\x01\x00\x00",
	},

	"/images/start.png": {
		local: "public/images/start.png",
		size:  334,
		compressed: "\x1f\x8b\b\x00\x00\tn\x88\x00\xff\xea\f\xf0s\xe7\xe5\x92\xe2b``\xe0\xf5\xf4p\t\x02Ҳ@\xac\xc0\xc1\x06$ߕ\x17\xd7\x00)\x96b'\xcf\x10\x0e \xa8\xe1H\xe9\x00\xf29\v<\"\x8b\x19\x18\xb8\x85A\x98\x91a\xd6\x1c\t\xa0\xe0\aO\x17\xc7\x10\x8fν\xd7\x15u\x8e\x06\x8a\xb0?\xc8OK\x0e\x16\xf1\xbc\x9c\xe1\xb2\xcaG" +
			"\xed\xf2\xc9\x14\xa3\x99\x0e\x8a\x93\x9f\x8agn\t\x0f\x88\v\xbc\xf7m\xdb\xf6\xed\xda\x05\xe6\xc9r\xcf\xfeץ\xff~\xfa\xfeݻws\xaa+V\x17\x9fἮ\xbc\xeb\xc6\xd1\b!\xdd\xf7\x99\x89\x8a\xbdb_6\xf4\u007fT\x89*n_j\x12uḢ\xda䌳\xcca\x972\x0e2\x1c\x9f\xe1\xb5\xc84wͲ\x8b\xb3\x82}\x96" +
			"_lJ_\xb3\xfc\xe2,\xef3K.6\xf1\x99]\xce8k\x9av9\xe3 \xf76 \xab\xd7?\xd9\xd8\xf0\x98ֿ\xa5n\xa5{\xa3\x0f\xfb]6\xa9R\xfd1\xc1\xea\xc6\x0f\xcel\xa1\xbd\x029|\xfe\x11\xbf\xb9\xbf\xc5\u007fq\xd9\xf0p;\xab\xfbm\x06\xf7&\xe7?r\xa7}\x13\xf6;\xfe\xb68\xe02\xf9\x81\xc1\x17\xb1n\xdf\xe4\xfb" +
			"\vm\xdfp\xd6m\x8b\x9e!\xaf\xfaM\xf0m\x92{A\xa3\x95p\x8ch<\xe3\xc6\uf179\xf5\x191\xf2@\xbf3x\xba\xfa\xb9\xacsJh\x02\x04\x00\x00\xff\xffQ\xe2\xc5\x02N\x01\x00\x00",
	},

	"/images/stop.png": {
		local: "public/images/stop.png",
		size:  140,
		compressed: "\x1f\x8b\b\x00\x00\tn\x88\x00\xff\xea\f\xf0s\xe7\xe5\x92\xe2b``\xe0\xf5\xf4p\t\x02\xd2\n \xcc\xc1\x06$\x8b\xab\xaa\xbe\x00)\x96b'\xcf\x10\x0e \xa8\xe1H\xe9\x00\xf29\v<\"\x8b\x19\x18\xb8\x85A\x98\x91a\xd6\x1c\t\xa0\xa0\x9e\xa7\x8bcHD\xeb\xdbs\x86\x8c\f\x06<\r\x06Q\xff\xe7\x99K\xd661\x1a]}\xbda" +
			"\xef\x1cf\x06\x14\x10R\xab(\xc4h\u007f\xb5\xf3A=\x88\xe7\xe9\xea\xe7\xb2\xce)\xa1\t\x10\x00\x00\xff\xff\xe6\x0e\xb8\x18\x8c\x00\x00\x00",
	},

	"/layout.html": {
		local: "public/layout.html",
		size:  1209,
		compressed: "\x1f\x8b\b\x00\x00\tn\x88\x00\xff\xb4\x94\xcfr\xdc \f\xc6\xef}\n\xaaS{\xd80\xb9u\xa6\xd83\xed\xbd\xa7>\x81\x16+6\t\xc6\f\x923\xf1\xdb\x17\xb0w\xc3N{H\xff\xdd0\x12ߧ\x9f\f2\x93̾\u007f\xa7\x94\x99\b\x87\xb2\xc8˙\x04\x95\x9d01I\a\xab<\x9c>\xc1\x11\x12'\x9e\xfa\xaf_\xbe\xbd7z_\xef" +
			"\xfbޅ'\x95\xc8w\xc0\xb2y\xe2\x89H@\xc9\x16\xa9\x03\xa1\x17і\x19Ԕ衃3\xcew\xe5\xb3\xdaꋯ9/\xc3v\xa8\r\xeeY\xb9\xa1\x03\xbb\x04A\x17(\x1d\xf6\xa5\xcc\xfb^\x19<\x94\xe0\xa8\x04\xfb\"t\u007fMr!\xaeҸCUc\xc2d\xa7\xd3yy\x01\x15=Z\x9a\x16?P\xea\xe0{\r\x80Z\xc2\x13" +
			"mk\xbcd~\xf8\xf8\x19z\xa3\xab\xd8Uz\xf5\xcazd\xee\xc0;\x96k]\xb5\x05j@\xc1\x13Ƭ\x90\xd0\xf9\f|\xa4\x8e\x89(4\xb9\xaa \x1c\xb1\xdaE\xb8\x02\xa9zR\x15\xa66\x9d#\x86~&f\x1c\xc9\xe8\xfaՆ_\xabB+n\t|cV\x8b뛮\xe5\x0eͣ\xe2d;psVdͲĻ\x18\xc6\n" +
			"<\x8f}-\xc0\xe8|\xac\xb5ѫo\x80o\xa2\xb7\xf8\xd1e\xa9KI\x89\x867\xb2\x97c?\xa1\xff\x1e\x9b\xaa\x9a\xe5\x1ab\x92_\x93\xe6\xc0?D\xcdz\xe2l\xf3\xabq{#\xed~\xf2/y\xff\x98\xf0\xff_\x8a\u05c8\xd1\xf9E\x1fK\xb6\xc9\xc5\xf6q\xeaG|\xc6}\x17v\xfb2 \x1e\xb9\xb8\xee\xdb\xfb\xa0\xd8\xe7C~\xe7u" +
			"b\xfd\b\x00\x00\xff\xffm\xa4~\xb2\xb9\x04\x00\x00",
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
