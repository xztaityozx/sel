package rw_test

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"xztaityozx/sel/rw"
)

func TestReadWrite(t *testing.T) {
	tmp := os.TempDir()
	base := filepath.Join(tmp, "sel_test")
	_ = os.MkdirAll(base, 0755)
	inFilePath := filepath.Join(base, "input")
	text := "This is input file"
	_ = ioutil.WriteFile(inFilePath, []byte(text), 0644)

	nullSelector := func(s string) (string, error) {
		return "", nil
	}

	as := assert.New(t)

	setup := func() {
		_ = os.MkdirAll(base, 0755)
		_ = ioutil.WriteFile(inFilePath, []byte(text), 0644)
	}

	teardown := func() { _ = os.RemoveAll(base) }
	defer teardown()

	t.Run("srcがファイルオープン出来ないとき", func(t *testing.T) {
		for _, ip := range []bool{true, false} {
			for _, kb := range []bool{true, false} {
				setup()
				err := rw.ReadWrite(filepath.Join(base, "ないよそんなの"), ip, kb, nullSelector)
				as.Error(err, "エラーが返されるべき")
				teardown()
			}
		}
	})

	t.Run("srcがファイルオープンできるとき", func(t *testing.T) {
		sel := func(s string) (string, error) {
			return strings.Join(strings.Split(s, " "), ","), nil
		}
		l, _ := sel(text)
		expect := []byte(l)

		for _, ip := range []bool{true, false} {
			for _, kb := range []bool{true, false} {
				setup()
				err := rw.ReadWrite(inFilePath, ip, kb, sel)
				as.Nil(err, "エラーを返さない")
				as.FileExists(inFilePath, "元のファイルがある")
				if ip && kb {
					as.FileExists(inFilePath+".bak", "バックアップがある")
					actual, err := ioutil.ReadFile(inFilePath + ".bak")
					as.Nil(err)
					as.Equal([]byte(text), actual, "バックアップは元のファイルの中身のままなべき")
				}

				if ip {
					actual, err := ioutil.ReadFile(inFilePath)
					as.Nil(err)
					as.Equal(expect, actual, "In-Placeなので、元のファイルが変更されているべき")
				}
				teardown()
			}
		}
	})
}
