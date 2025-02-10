package test

import (
	"os"
	"path/filepath"
)

func ProjectRoot() string {
  current, err := os.Getwd()
  if err != nil {
    panic(err)
  }

  for {
    _, err := os.ReadFile(filepath.Join(current, "go.mod"))
    if os.IsNotExist(err) {
      if current == filepath.Dir(current) {
        panic("failed to find project root")
      }

      current = filepath.Dir(current)
      continue
    } else if err != nil {
      panic(err)
    }

    break
  }

  return current
}
