package gofncstd3000

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"crypto/md5"

	"github.com/satori/go.uuid"

	"os"

	"io"
	"io/ioutil"
	"path/filepath"

	"encoding/json"
	"runtime/debug"
)

//------------------------------------------------------------------------------
//функции для работы с ошибками
func ErrStr(err error) string {
	s := fmt.Sprintf("%+v", err)
	a := string(debug.Stack())
	//убераем указатель на текущую строку
	i := strings.Index(a, "\n")
	a = a[i+1:]
	i = strings.Index(a, "\n")
	a = a[i+1:]

	s += "\n" + a
	return s
}

//------------------------------------------------------------------------------
//функции для вывода лога загрузки страниц
type slogreq struct {
	HandleName string
	Fnc        func(http.ResponseWriter, *http.Request)
}

func (p *slogreq) ServeHTTP(p1 http.ResponseWriter, p2 *http.Request) {
	log.Println(p.HandleName + " <-")
	defer log.Println(p.HandleName + " ->")
	p.Fnc(p1, p2)
}

func LogreqH(name string, x http.Handler) http.Handler {
	a := new(slogreq)
	a.HandleName = name
	a.Fnc = x.ServeHTTP
	return a
}

func LogreqF(name string, f func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(p1 http.ResponseWriter, p2 *http.Request) {
		log.Println(name + " <-")
		defer log.Println(name + " ->")
		f(p1, p2)
	}
}

//------------------------------------------------------------------------------
//функции вывода даты и времени
func CurTimeStr() string {
	t := time.Now()
	p := fmt.Sprintf("%s", strings.Replace(t.Format(time.RFC3339)[0:19], "T", " ", 1))
	return p
}

func CurTimeStrShort() string {
	//2016-04-02T18:21:09+03:00
	t := time.Now()
	p := fmt.Sprintf("%s", t.Format(time.RFC3339)[0:19])
	p = p[0:19]
	p = strings.Replace(p, "-", "", -1)
	p = strings.Replace(p, ":", "", -1)
	p = strings.Replace(p, "T", "-", -1)
	return p
}

//------------------------------------------------------------------------------
//функции для работы со строками
func RegexpCompile(re string) (*regexp.Regexp, error) {
	return regexp.Compile(re)
}

func StrReplaceRegexp(text string, regx_from string, to string) (string, error) {
	reg, err := regexp.Compile(regx_from)
	if err != nil {
		return text, err
	}
	text = reg.ReplaceAllString(text, to)
	return text, nil
}

func StrReplaceRegexp2(text string, regx_from string, to string) string {
	reg, err := regexp.Compile(regx_from)
	if err != nil {
		return text
	}
	text = reg.ReplaceAllString(text, to)
	return text
}

func IntToStr(i int) string {
	return strconv.Itoa(i)
}
func StrToInt(s string) (int, error) {
	return strconv.Atoi(s)
}

func StrPos(s, substr string) int {
	return strings.Index(s, substr)
}

//JSON
//преобразует структуру в json строку
func ToJson(v interface{}) ([]byte, error) {
	j, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return j, nil
}

//преобразует из json строки в map[string]interface{}
func FromJson(data []byte) (map[string]interface{}, error) {
	var d map[string]interface{}
	err := json.Unmarshal(data, &d)
	if err != nil {
		return nil, err
	}
	return d, nil
}

//------------------------------------------------------------------------------
//функции для работы с файлами
func ReadFile(file string) ([]byte, error) {
	d, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func ReadFileStr(file string) (string, error) {
	d, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	return string(d), nil
}

func WriteFile(file string, data []byte) error {
	err := ioutil.WriteFile(file, data, 0644)
	return err
}

func WriteFileStr(file string, data string) error {
	err := ioutil.WriteFile(file, []byte(data), 0644)
	return err
}

func RemoveFile(file string) error {
	return os.Remove(file)
}

func Rename(src, dst string) error {
	return os.Rename(src, dst)
}

func FileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

func MkdirAll(path string) error {
	return os.MkdirAll(path, 0777)
}

func CopyFile2(src, dst string) error {
	d, err := ReadFile(src)
	if err != nil {
		return err
	}
	err = WriteFile(dst, d)
	return err
}

//эта штуковина не копирует файл а создает на него ссылку или что то типа того
//при изменении src меняется и dst!
// CopyFile copies a file from src to dst. If src and dst files exist, and are
// the same, then return success. Otherise, attempt to create a hard link
// between the two files. If that fail, copy the file contents from src to dst.
func CopyFile(src, dst string) (err error) {
	sfi, err := os.Stat(src)
	if err != nil {
		return
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return
		}
	}
	if err = os.Link(src, dst); err == nil {
		return
	}
	err = copyFileContents(src, dst)
	return
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

//------------------------------------------------------------------------------
//функции для работы с каталогами
//текущий путь к приложению
func AppPath() (string, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", err
	}
	return dir, err
}

//список файлов в каталоге path
func ReadDir(path string) ([]os.FileInfo, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	return files, err
}

//------------------------------------------------------------------------------
//crypto
func Uuid() string {
	return fmt.Sprintf("%s", uuid.NewV4())
}

func Md5(text []byte) string {
	d := md5.Sum(text)
	s := fmt.Sprintf("%x", d)
	return s
}
