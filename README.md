# ringbuf

## Introduction

## API
* func MakeRingBuf(size int) *RingBuf
* func (rb *RingBuf) Read(p []byte)(int, error)
* func (rb *RingBuf) Write(p []byte)(int, error)
* func (rb *RingBuf) Close()

