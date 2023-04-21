package main

import (
	"fmt"
	"os"
	"sync/atomic"
	"time"
)

func cmdServer() {
	shm, err := NewShmemServer("my_shmem", 0, ShmemOptions{
		ArraySize:    0x1000,
		MemBlockSize: 0,
	})
	if err != nil {
		panic(err)
	}
	defer shm.Close()

	pid := os.Getpid()

	fmt.Printf("Start /proc/%d/fd/%d\n", pid, shm.fd)

	// copy(shm.RawData(), []byte("hello world"))
	// if err := shm.Commit(); err != nil {
	// 	panic(err)
	// }

	for {
		fmt.Printf("VIEW: %+#v", shm.RawData()[:512])

		head := atomic.LoadUint32(&shm.Header.Head)
		atomic.AddUint32(&shm.Header.Tail, 1)
		tail := atomic.LoadUint32(&shm.Header.Tail)

		atomic.StoreUint32(&shm.Header.Tail, tail)
		fmt.Printf("header: %p head=%+#v tail=%+#v\n", shm.Header,
			head, tail)

		// if err := shm.Commit(); err != nil {
		// 	panic(err)
		// }
		time.Sleep(time.Millisecond * 100)
	}

}
