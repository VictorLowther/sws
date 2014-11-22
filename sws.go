package main
import (
	"flag"
	"net/http"
	"log"
	"os"
	"path"
)

var listen, site string
func init() {
	flag.StringVar(&listen,"listen","","Address:port to listen on.  Address can be left blank.")
	flag.StringVar(&site,"site","","Path to serve static files from")
}

func main() {
	flag.Parse()
	if listen == "" || site == "" {
		flag.Usage()
		os.Exit(1)
	}
	sitepath := path.Clean(site)
	stat,err := os.Stat(sitepath)
	if err != nil {
		log.Fatal(err)
	}
	if !stat.IsDir() {
		log.Fatalf("%v is not a directory!\n",sitepath)
	}
	fs := http.FileServer(http.Dir(sitepath))
	http.Handle("/",fs)
	err = http.ListenAndServe(listen,nil)
	log.Fatalf("Listen error: %v",err)
}
