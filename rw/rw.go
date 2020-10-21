package rw

import (
	"bufio"
	"io"
	"os"
)

func ReadWrite(src string, inPlace, keepBackup bool, selector func(string) (string, error)) error {
	tmp := src + ".tmp"
	iStream, err := os.Open(src)
	if err != nil {
		return err
	}

	oStream, tmpFp, err := func() (io.Writer, *os.File, error) {
		if inPlace {
			fp, err := os.Stat(src)
			if err != nil {
				return nil, nil, err
			}
			t, err := os.OpenFile(tmp, os.O_CREATE|os.O_RDWR, fp.Mode())
			if err != nil {
				return nil, nil, err
			}

			return io.MultiWriter(os.Stdout, t), t, nil
		}

		return os.Stdout, nil, nil
	}()

	if err != nil {
		return err
	}

	err = func() error {
		defer func() { _ = tmpFp.Close() }()
		defer func() { _ = iStream.Close() }()

		scan := bufio.NewScanner(iStream)
		w := bufio.NewWriter(oStream)
		for scan.Scan() {
			line, err := selector(scan.Text())
			if err != nil {
				return err
			}
			if _, err := w.WriteString(line); err != nil {
				return err
			}
		}

		if err := w.Flush(); err != nil {
			return err
		}

		return nil
	}()

	if err != nil {
		return err
	}

	if inPlace {
		if keepBackup {
			// backupを作る
			if err := os.Rename(src, src+".bak"); err != nil {
				return err
			}
		}
		// tmpをsrcに変える
		_ = os.Rename(tmp, src)
	}

	return nil
}
