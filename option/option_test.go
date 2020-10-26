package option_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xztaityozx/sel/option"
)

func TestInputFiles_Enumerate(t *testing.T) {
	as := assert.New(t)

	t.Run("配列がnilならエラーが返されるべき", func(t *testing.T) {
		_, err := option.InputFiles{Files: nil}.Enumerate()
		as.Error(err)
	})

	t.Run("配列の長さが0ならエラーが返されるべき", func(t *testing.T) {
		_, err := option.InputFiles{Files: []string{}}.Enumerate()
		as.Error(err)
	})

	base := filepath.Join(os.TempDir(), "sel_test")
	_ = os.MkdirAll(base, 0755)
	_ = os.Chdir(base)
	t.Run("Open出来ないファイルがあると例外が投げられる", func(t *testing.T) {
		actual, err := option.InputFiles{Files: []string{"ないわよ"}}.Enumerate()
		as.Error(err)
		as.Nil(actual)

		actual, err = option.InputFiles{Files: []string{"ないわよ"}}.Enumerate()
		as.Error(err)
		as.Nil(actual)
	})

	t.Run("OpenできるファイルのみならOK", func(t *testing.T) {
		var files []string
		for i := 0; i < 10; i++ {
			f := filepath.Join(base, fmt.Sprint(i))
			files = append(files, f)
			_ = ioutil.WriteFile(f, []byte("はい"), 0644)
		}

		a, err := option.InputFiles{Files: files}.Enumerate()

		as.Nil(err)
		as.Equal(10, len(a))

		for i, v := range a {
			as.Equal(filepath.Join(base, fmt.Sprint(i)), v)
		}
	})

	t.Run("ディレクトリはOpenできない", func(t *testing.T) {
		_, err := option.InputFiles{Files: []string{base}}.Enumerate()

		as.Error(err)
	})

	t.Run("スペシャルファイルはOpenできない", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip()
		}
		_, err := option.InputFiles{Files: []string{"/dev/null"}}.Enumerate()
		as.Error(err)
	})

	t.Run("Globもいける", func(t *testing.T) {
		a, err := option.InputFiles{Files: []string{filepath.Join(base, "*")}}.Enumerate()

		as.Nil(err)
		as.Equal(10, len(a))

		for i, v := range a {
			as.Equal(filepath.Join(base, fmt.Sprint(i)), v)
		}
	})

	_ = os.RemoveAll(base)
}
