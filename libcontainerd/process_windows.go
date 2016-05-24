package libcontainerd

import (
	"io"

	"github.com/docker/docker/pkg/system"
)

// process keeps the state for both main container process and exec process.
type process struct {
	processCommon

	// Platform specific fields are below here.

	// commandLine is to support returning summary information for docker top
	commandLine string
}

func openReaderFromPipe(p io.ReadCloser) io.Reader {
	r, w := io.Pipe()
	go func() {
		if _, err := io.Copy(w, p); err != nil {
			r.CloseWithError(err)
		}
		w.Close()
		p.Close()
	}()
	return r
}

// fixStdinBackspaceBehavior works around a bug in Windows before build 14350
// where it interpreted DEL as VK_DELETE instead of as VK_BACK. This replaces
// DEL with BS to work around this.
func fixStdinBackspaceBehavior(w io.WriteCloser, tty bool) io.WriteCloser {
	if !tty || system.GetOSVersion().Build >= 14350 {
		return w
	}
	return &delToBsWriter{w}
}

type delToBsWriter struct {
	io.WriteCloser
}

func (w *delToBsWriter) Write(b []byte) (int, error) {
	const (
		backspace = 0x8
		del       = 0x7f
	)
	bc := make([]byte, len(b))
	for i, c := range b {
		if c == del {
			bc[i] = backspace
		} else {
			bc[i] = c
		}
	}
	return w.WriteCloser.Write(bc)
}
