//go:build linux

package bpf

import (
	"errors"
	"io"
	"log"
	"os"
	"syscall"
)

// closeAll detaches resources in reverse attachment order, which is safer when
// later resources depend on earlier setup such as the clsact qdisc.
func closeAll(closers []io.Closer) {
	for i := len(closers) - 1; i >= 0; i-- {
		if err := closers[i].Close(); err != nil {
			log.Printf("close link: %v", err)
		}
	}
}

func isExists(err error) bool {
	return errors.Is(err, os.ErrExist) || errors.Is(err, syscall.EEXIST)
}

func isNotFound(err error) bool {
	return errors.Is(err, os.ErrNotExist) || errors.Is(err, syscall.ENOENT)
}
