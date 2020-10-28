package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompletion_Gen(t *testing.T) {
	as := assert.New(t)

	dict := map[string]func() []byte{
		"zsh": func() []byte {
			var data []byte
			buf := bytes.NewBuffer(data)
			rootCmd.GenZshCompletion(buf)

			return data
		},
		"bash": func() []byte {
			var data []byte
			buf := bytes.NewBuffer(data)
			rootCmd.GenBashCompletion(buf)

			return data
		},
		"fish": func() []byte {
			var data []byte
			buf := bytes.NewBuffer(data)
			rootCmd.GenFishCompletion(buf, true)

			return data
		},
		"pwsh": func() []byte {
			var data []byte
			buf := bytes.NewBuffer(data)
			rootCmd.GenPowerShellCompletion(buf)

			return data
		},
		"PowerShell": func() []byte {
			var data []byte
			buf := bytes.NewBuffer(data)
			rootCmd.GenPowerShellCompletion(buf)

			return data
		},
	}

	for shell, expect := range dict {
		t.Run(shell, func(t *testing.T) {
			var data []byte
			buf := bytes.NewBuffer(data)
			err := Completion{W: buf}.Gen(rootCmd, shell)

			as.Nil(err)
			as.Equalf(expect(), data, "shell: %s", shell)
		})
	}

	t.Run("invalid shell name", func(t *testing.T) {
		var data []byte
		buf := bytes.NewBuffer(data)
		err := Completion{W: buf}.Gen(rootCmd, "")

		as.Error(err)
	})
}
