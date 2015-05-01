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
		size:  1396,
		compressed: "\x1f\x8b\b\x00\x00\tn\x88\x00\xff|TMo\xdb0\f\xbd\xe7W\x10-v\x8b\r\xbbu\x16\xc4\x05\x06tM\x87\x9d\xf6\x1f\x18\x89v\x88\xc9R +m\xb2\xa2\xff}\x92\xbfkoEN!\xdf#)\xf2=\x1f\x8c\xbc\xc2\xdb\n\xe0\x80\xe2wi\xcdY\xcbH\x18el\x0e\xb7$\x8a\xa4H\x1f|\xb20\xdaE\x05V\xac\xae9\xdc\xfc$" +
			"\xf5B\x8e\x05\xc2/:\xd3\xcdz\xf8\xbf~\xb4\x8cj]\xa3\xae\xa3\x9a,\x17\x03\xb7\xe6?\x94C\x9a\x9d.\x0f\xab\xf7\xd51mZ6\x99W\xe2\xf2\xe8r\xd8&\xc9\f}\x17o\xa8\n\xf0[\xe1\x83ȚlC{e\xe9\x8e9\xec\x92/\x81P\xe1%\xea\"\xdbM\x12\xea\x03\x9cPJ\xd6ed\xdb\xd2\xe9\xe6cXQ1\x89VhK" +
			"\xd6=\x16\xcf\xceL\xa2-\xb4\r\xbe\xafΪ\x19@q\xedgtW\xe5\x87\xd4FӤv\x0e\xc9\xc8\xce\xc1\xb7\b\x81\xf7\x95bx\x83~\xaf\x88\xf8\x00\xa1Z\x1c*\xc17\b\xd9i\x8d4\x99\x8e֔\t\u007f\x0f\xc6J\xf2\x05R_\xb56\x8a%\xdc\n!\xc6LdQ\xf2\xb9Ρ[\xf3\xb4A\\Z\"\xed\x87\xe8\xa0\xed\xc36c" +
			"\xa5\x14\x0fb'\xe6s\xc5WRʼ~\xc2+R\x91%łgI~B\xa2m&\xee\x97\xcdJ\x8b\xd7OX\a\xe9Iێ\x85±ѵ\x87\x17ʠ\a6\a\x9c%\x87\xd5J\xaeO\n\xbdxY+\x9e]l3l\xab\xa7qUN\x95\x96~m\xf7\xffB6\xa8\\E\xa8\xb8\xf4g\xa9XJE\x81\x8bk\xc0\xfc\x85k" +
			"v\xe1\xd9\x1eڟ:I\xb6\xd9~\x17Ȏ..\x92$\x8c\xc5\xd0dT\xceh\xbc6\x06\x96N\x84\x0eja\x8dR\x90\xf8\x9f\xb3\xdeQ'\xb4\xa4]\xd3-v\xec\x14\x8d\x1e\xea\xfc\x12\xdc2sզuUxst\xecbi\x9cN\xddp0ΙjX\xc3 \x95\xa1\xcb\xf0\x98Q\"\xa30\x96\xa8Q\x10\x9d\f\x96\x90\xf1\xfc\xfd" +
			"ї\x98\xf1ث\x98\xac5\xd6\x0fz\x99٤\xf7\xc5\xf2\xd3\xf5\xe3n\xff\xbc\u007f\x9eX\xa3O<\u007f\u007fz\xda7\xef\xef#\x8f\xbb,\xcb\xee\xfe\xed\xaf\x0f\x8b\xff\x9f\xcd&\xe3\x1d\xef\x9b\t{\xdf&s\x9d\xb5\x1f\x83\xbf\x01\x00\x00\xff\xff:}\xef]t\x05\x00\x00",
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
		size:  982,
		compressed: "\x1f\x8b\b\x00\x00\tn\x88\x00\xff\xb4\x93On\xf3 \x10\xc5\xf7\xdf)\xe6\xe3\x00A\xd9c\xa4v\xdfC\x10{j\x8f\x8a\xb1\xc5L\xaa\xfa\xf6\x05\xec8D\xed\"\xfd\xb7\x1b\xf2\xe0\xcd\xfb)~f\x90\xd1\xdb\u007f\x00f@\xd7\xe5!\x8dB\xe2\xd1>><\xfd7z\x9d\xd7\xdf=\x85\x17\x88\xe8\x1bŲx\xe4\x01Q\x14\xc82c\xa3\x04\xdf" +
			"D\xb7\xcc\n\x86\x88ύ:\xb9\xf1\x90\x8f\xc5[_\xcc\xcdi\xea\x96ͭ\xa3W\xa0\xaeQ\xed\x14\xc4Q\xc0\xa8V!g9Z0nsR[\x12g\xb3\xd1q\xbft\xf6\xd0z\xc7\xdc(O,\xfb\xe3\x92\xf3\xa2\xf4\x111T\x12d\xdbM+dj_\x02ёg\xc8{\xea\xeb<\xbb`Gdv=\x1a]N\xb5|\r\xe1Z\xa1" +
			")\xf0Ͳ\x92\xc5V$`h\xec\x81c\xdb(\x1a\x93#k\x96i>̡W\xd6\xe8\xa4\xd9\x12\xc0\xe8\xf4\xac^\xa3Ͼ\xe2\xbbQ+ڈݝ\xac3\x85\xfe\x03\xea\xd7X\xa0x\xe6O\xc1E\xf9\x9c,\t\xbf\x85\xd6G\xb7\xdcɖ\xf6\n\xb5?\xa4\xfb6\xcf\xdf\xff\xe5W\xc5\xe8ԡ\xb5`k\xafR?J\x9d\xdf\x03\x00\x00" +
			"\xff\xff\x94-\xc6v\xd6\x03\x00\x00",
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
