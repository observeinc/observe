package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"
)

type Output interface {
	io.Writer // for data
	Error(fmt string, args ...any)
	Info(fmt string, args ...any)
	Debug(fmt string, args ...any)
	Exit(status int)
}

type DefaultOutput struct {
	EnableDebug bool
	DisableInfo bool
	DataOutput  io.Writer
	DoTimestamp bool
}

var _ Output = DefaultOutput{}

func datestr() string {
	return time.Now().Format(time.RFC3339)
}

func (d DefaultOutput) Error(ff string, args ...any) {
	fmt.Fprintf(os.Stderr, d.Timestamp("E ")+ff, args...)
}

func (d DefaultOutput) Info(ff string, args ...any) {
	if !d.DisableInfo {
		fmt.Fprintf(os.Stderr, d.Timestamp("I ")+ff, args...)
	}
}

func (d DefaultOutput) Debug(ff string, args ...any) {
	if d.EnableDebug {
		fmt.Fprintf(os.Stderr, d.Timestamp("D ")+ff, args...)
	}
}

func (t DefaultOutput) Write(data []byte) (int, error) {
	return t.DataOutput.Write(data)
}

func (d DefaultOutput) Timestamp(kind string) string {
	if !d.DoTimestamp {
		return ""
	}
	return kind + datestr() + " "
}

func (t DefaultOutput) Exit(i int) {
	os.Exit(i)
}

type TaggedOutput struct {
	Chain  Output
	Prefix string
}

var _ Output = TaggedOutput{}

func (t TaggedOutput) Error(ff string, args ...any) {
	t.Chain.Error("%s: "+ff, append([]any{t.Prefix}, args...)...)
}

func (t TaggedOutput) Info(ff string, args ...any) {
	t.Chain.Info("%s: "+ff, append([]any{t.Prefix}, args...)...)
}

func (t TaggedOutput) Debug(ff string, args ...any) {
	t.Chain.Debug("%s: "+ff, append([]any{t.Prefix}, args...)...)
}

func (t TaggedOutput) Write(data []byte) (int, error) {
	return t.Chain.Write(data)
}

func (t TaggedOutput) Exit(i int) {
	t.Chain.Exit(i)
}

type CaptureOutput struct {
	DebugBuf  bytes.Buffer
	InfoBuf   bytes.Buffer
	ErrorBuf  bytes.Buffer
	OutputBuf bytes.Buffer
}

func NewCaptureOutput() *CaptureOutput {
	return &CaptureOutput{}
}

func (c *CaptureOutput) Debug(ff string, args ...any) {
	fmt.Fprintf(&c.DebugBuf, ff, args...)
}

func (c *CaptureOutput) Info(ff string, args ...any) {
	fmt.Fprintf(&c.InfoBuf, ff, args...)
}

func (c *CaptureOutput) Error(ff string, args ...any) {
	fmt.Fprintf(&c.ErrorBuf, ff, args...)
}

func (c *CaptureOutput) Write(data []byte) (int, error) {
	return c.OutputBuf.Write(data)
}

func (c *CaptureOutput) Exit(status int) {
	// really, this is the least bad option for testing ...
	// Typically, this will turn into two panics; one for the
	// first exit, and one for the exit called in the panic
	// handler of RunRecoverWithTag().
	panic(status)
}
