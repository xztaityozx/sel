package option_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/spf13/viper"

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

func TestGetOptionNames(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{name: "とれてますか", want: []string{option.NameInputFiles, option.NameInputDelimiter, option.NameOutPutDelimiter, option.NameUseRegexp, option.NameRemoveEmpty, option.NameSplitBefore, option.NameFieldSplit}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := option.GetOptionNames(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetOptionNames() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewOption(t *testing.T) {
	type args struct {
		v *viper.Viper
	}
	tests := []struct {
		name string
		args args
		want option.Option
	}{
		{
			name: "noname", args: args{
				v: func() *viper.Viper {
					v := viper.New()
					v.Set(option.NameInputFiles, []string{"abc", "def"})
					v.Set(option.NameInputDelimiter, "i")
					v.Set(option.NameOutPutDelimiter, "o")
					v.Set(option.NameRemoveEmpty, true)
					v.Set(option.NameUseRegexp, true)
					v.Set(option.NameSplitBefore, true)
					return v
				}(),
			},
			want: option.Option{
				InputFiles: option.InputFiles{Files: []string{"abc", "def"}},
				DelimiterOption: option.DelimiterOption{
					InputDelimiter:  "i",
					OutPutDelimiter: "o",
					RemoveEmpty:     true,
					UseRegexp:       true,
					SplitBefore:     true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := option.NewOption(tt.args.v); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewOption() = %v, want %v", got, tt.want)
			}
		})
	}
}
