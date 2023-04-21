package main

import (
	"fmt"
	"sync/atomic"
	"time"

	"golang.org/x/sys/unix"
)

func cmdClient(path string) {
	fd, err := unix.Open(path, unix.O_RDWR, 0)
	if err != nil {
		panic(err)
	}

	shm, err := NewShmem(fd, ShmemOptions{
		ArraySize:    0x1000,
		MemBlockSize: 0,
	})
	if err != nil {
		panic(err)
	}
	defer shm.Close()

	for {
		head := atomic.LoadUint32(&shm.Header.Head)
		tail := atomic.LoadUint32(&shm.Header.Tail)

		atomic.AddUint32(&shm.Header.Head, 1)
		atomic.AddUint32(&shm.Header.Tail, 1)

		fmt.Printf("header: %p head=%+#v tail=%+#v\n", shm.Header,
			head, tail)

		// if err := shm.Commit(); err != nil {
		// 	panic(err)
		// }
		time.Sleep(time.Millisecond * 100)
	}

}
