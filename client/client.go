package main

import (
	"io"
	"log"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", ":2023")
	if err != nil {
		log.Println(err)
	}

	done := make(chan struct{})
	go func() {
		io.Copy(os.Stdout, conn) // Note: ignoring errors
		log.Println("done")
		done <- struct{}{}
	}()

	mustCopy(conn, os.Stdin)
	conn.Close()
	<-done
}

func mustCopy(dst io.Writer, src io.Reader) {
	if _, err := io.Copy(dst, src); err != nil {
		log.Fatal()
	}
}
