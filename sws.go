package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/pin/tftp"
)

var httpListen, tftpListen, site string

func handleTftpRead(filename string, rf io.ReaderFrom) error {
	p := filepath.Join(site, filename)
	p = filepath.Clean(p)
	if !strings.HasPrefix(p, site) {
		err := fmt.Errorf("Filename %s tries to escape root %s", filename, site)
		log.Println(err)
		return err
	}
	log.Printf("Sending %s from %s", filename, p)
	file, err := os.Open(p)
	if err != nil {
		log.Println(err)
		return err
	}
	if t, ok := rf.(tftp.OutgoingTransfer); ok {
		if fi, err := file.Stat(); err == nil {
			t.SetSize(fi.Size())
		}
	}
	n, err := rf.ReadFrom(file)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Printf("%d bytes sent\n", n)
	return nil
}

func main() {
	flag.StringVar(&httpListen, "listen", "", "Address:port to listen on for HTTP requests.  Address can be left blank.")
	flag.StringVar(&tftpListen, "tftp", "", "Address:port to listen on for TFTP requests.  Address can be left blank.")
	flag.StringVar(&site, "site", "", "Path to serve static files from")
	flag.Parse()
	if site == "" {
		flag.Usage()
		os.Exit(1)
	}
	sitepath := path.Clean(site)
	stat, err := os.Stat(sitepath)
	if err != nil {
		log.Fatal(err)
	}
	if !stat.IsDir() {
		log.Fatalf("%v is not a directory!\n", sitepath)
	}
	if tftpListen != "" {
		a, err := net.ResolveUDPAddr("udp", tftpListen)
		if err != nil {
			log.Fatalf("Error resolving %s for tftp: %v", tftpListen, err)
		}
		conn, err := net.ListenUDP("udp", a)
		if err != nil {
			log.Fatalf("Error opening %s for tftp: %v", tftpListen, err)
		}
		svr := tftp.NewServer(handleTftpRead, nil)
		go svr.Serve(conn)
	}
	if httpListen != "" {
		conn, err := net.Listen("tcp", httpListen)
		if err != nil {
			log.Fatalf("Error opening %s for http: %v", httpListen, err)
		}
		fs := http.FileServer(http.Dir(sitepath))
		http.Handle("/", fs)
		go http.Serve(conn, nil)
	}
	for httpListen != "" || tftpListen != "" {
		time.Sleep(600 * time.Second)
	}
}
