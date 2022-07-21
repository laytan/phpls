package connection

import (
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

type Stdio struct {
	in  io.ReadCloser
	out io.WriteCloser
}

func NewDefaultStdio() *Stdio {
	return NewStdio(os.Stdin, os.Stdout)
}

func NewStdio(in io.ReadCloser, out io.WriteCloser) *Stdio {
	return &Stdio{
		in:  in,
		out: out,
	}
}

// Read implements io.Reader interface.
func (s *Stdio) Read(b []byte) (int, error) {
	n, err := s.in.Read(b)
	if err != nil {
		return 0, fmt.Errorf("Error reading from stdio: %v", err)
	}

	return n, nil
}

// Write implements io.Writer interface.
func (s *Stdio) Write(b []byte) (int, error) {
	n, err := s.out.Write(b)
	if err != nil {
		return 0, fmt.Errorf("Error writing to stdio: %v", err)
	}

	return n, nil
}

// Close implements io.Closer interface.
func (s *Stdio) Close() error {
	errIn := s.in.Close()
	errOut := s.out.Close()

	if errIn != nil || errOut != nil {
		return fmt.Errorf("Errors closing stdio, stdin: %v, stdout: %v", errIn, errOut)
	}

	return nil
}

// LocalAddr implements net.Conn interface.
func (s Stdio) LocalAddr() net.Addr { return s }

// RemoteAddr implements net.Conn interface.
func (s Stdio) RemoteAddr() net.Addr { return s }

// SetDeadline implements net.Conn interface.
func (Stdio) SetDeadline(t time.Time) error { return nil }

// SetReadDeadline implements net.Conn interface.
func (Stdio) SetReadDeadline(t time.Time) error { return nil }

// SetWriteDeadline implements net.Conn interface.
func (Stdio) SetWriteDeadline(t time.Time) error { return nil }

// Network implements net.Addr interface.
func (Stdio) Network() string { return "Stdio" }

// String implements net.Addr interface.
func (Stdio) String() string { return "Stdio" }
