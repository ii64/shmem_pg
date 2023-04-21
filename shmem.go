package main

import (
	"os"
	"reflect"
	"syscall"
	"unsafe"

	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

var (
	PageSize = os.Getpagesize()

	SizeofShmemHeader = unsafe.Sizeof(ShmemHeader{})
)

func _SizeCheck() {
	// var x [1]struct{}
	// _ = x[SizeofShmemHeader-16]
}

type ShmemOffsets struct {
	HeaderOffset   uintptr
	ArrayOffset    uintptr
	MemBlockOffset uintptr
}

type Shmem struct {
	Header   *ShmemHeader
	Array    []unsafe.Pointer
	MemBlock []byte

	Offsets ShmemOffsets

	mmptr unsafe.Pointer
	size  uintptr
	fd    int
}

type ShmemOptions struct {
	ArraySize    uint32
	MemBlockSize uint32
}

func (o ShmemOptions) Default() ShmemOptions {
	return o
}

type ShmemHeader struct {
	Head uint32
	Tail uint32
}

func NewShmemServer(name string, flags int, opts ShmemOptions) (*Shmem, error) {
	fd, err := unix.MemfdCreate(name, flags)
	if err != nil {
		err = errors.Wrap(err, "memfd_create")
		return nil, err
	}
	return NewShmem(fd, opts)
}

func NewShmem(fd int, opts ShmemOptions) (*Shmem, error) {
	var err error
	offsets := ShmemOffsets{HeaderOffset: 0}
	offsets.ArrayOffset = offsets.HeaderOffset + SizeofShmemHeader
	offsets.MemBlockOffset = offsets.ArrayOffset + uintptr(opts.ArraySize)
	size := offsets.MemBlockOffset + uintptr(opts.MemBlockSize)
	if err = unix.Ftruncate(fd, int64(size)); err != nil {
		err = errors.Wrap(err, "ftrunc")
		return nil, err
	}

	addr, err := mmap(nil, size,
		unix.PROT_READ|unix.PROT_WRITE,
		unix.MAP_SHARED, //|unix.MAP_ANONYMOUS,
		fd, 0)
	if err != nil {
		err = errors.Wrap(err, "mmap")
		return nil, err
	}

	shAddr := (*ShmemHeader)(unsafe.Add(addr, offsets.HeaderOffset))
	arrAddr := unsafe.Add(addr, offsets.ArrayOffset)
	memAddr := unsafe.Add(addr, offsets.MemBlockOffset)

	sh := &Shmem{
		fd:      fd,
		mmptr:   addr,
		size:    size,
		Offsets: offsets,
		Header:  shAddr,
	}

	_ = sh.Header.Head
	_ = sh.Header.Tail

	arr := (*reflect.SliceHeader)(unsafe.Pointer(&sh.Array))
	arr.Data = uintptr(arrAddr)
	arr.Len = 0
	arr.Cap = int(opts.ArraySize)

	mmb := (*reflect.SliceHeader)(unsafe.Pointer(&sh.MemBlock))
	mmb.Data = uintptr(memAddr)
	mmb.Len = 0
	mmb.Cap = int(opts.MemBlockSize)

	return sh, nil
}

func (c *Shmem) RawData() (b []byte) {
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh.Data = uintptr(c.mmptr)
	sh.Cap = int(c.size)
	sh.Len = sh.Cap
	return
}

func (c *Shmem) Commit() error {
	return msync(c.mmptr, c.size, unix.MS_SYNC)
}

func (c *Shmem) Close() error {
	munmap(c.mmptr, c.size)
	return unix.Close(c.fd)
}

//go:linkname mmap syscall.mmap
func mmap(addr unsafe.Pointer, length uintptr, prot int, flags int, fd int, offset int64) (xaddr unsafe.Pointer, err error)

//go:linkname munmap syscall.munmap
func munmap(addr unsafe.Pointer, length uintptr) (err error)

func msync(addr unsafe.Pointer, length uintptr, flags uintptr) error {
	r1, _, e1 := syscall.Syscall(syscall.SYS_MSYNC, uintptr(addr), length, flags)
	if e1 != 0 {
		return syscall.Errno(e1)
	}
	_ = r1
	return nil
}
