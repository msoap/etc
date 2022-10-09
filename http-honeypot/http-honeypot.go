/*

Honeypot for http tracking.

	sudo sh -c 'echo "127.0.0.1 google-analytics.com" >> /etc/hosts'
	sudo sh -c 'echo "127.0.0.1 googleads.g.doubleclick.net" >> /etc/hosts'

	go build http-honeypot.go
	sudo ./http-honeypot

	# for close:
	curl http://localhost/exit_now

*/

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func getHeader(headers map[string][]string, name string) string {
	result, ok := headers[name]
	if !ok {
		result = []string{"-"}
	}
	return result[0]
}

func main() {
	fmt.Println("Begin")

	end := make(chan bool)
	logfile, err := os.OpenFile("http-honeypot.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer logfile.Close()
	log.SetOutput(logfile)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "")

		ua := getHeader(r.Header, "User-Agent")
		referer := getHeader(r.Header, "Referer")

		logString := strings.Join([]string{r.Method, r.Host, r.URL.Path, ua, referer}, "\t")
		log.Println(logString)
		fmt.Println(logString)

		if r.URL.Path == "/exit_now" && r.Host == "localhost" {
			end <- true
		}
	})

	go func() {
		err := http.ListenAndServe("127.0.0.1:80", nil)
		if err != nil {
			fmt.Println(err)
		}
	}()

	go func() {
		// generate SSL self-signed sertificate:
		//	   go run /usr/local/Cellar/go/1.4/libexec/src/crypto/tls/generate_cert.go --host="localhost"
		// but dont work because SSL :(
		certName, keyName := "cert.pem", "key.pem"
		_, errCert := os.Stat(certName)
		_, errKey := os.Stat(keyName)
		if errCert == nil && errKey == nil {
			err := http.ListenAndServeTLS("127.0.0.1:443", certName, keyName, nil)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			err := http.ListenAndServe("127.0.0.1:443", nil)
			if err != nil {
				fmt.Println(err)
			}
		}
	}()

	<-end
	fmt.Println("Finish")
}

/*

&http.Request{
  Method: "GET",
  URL:    &url.URL{
    Scheme:   "",
    Opaque:   "",
    User:     (*url.Userinfo)(nil),
    Host:     "",
    Path:     "/exit_now",
    RawQuery: "",
    Fragment: "",
  },
  Proto:      "HTTP/1.1",
  ProtoMajor: 1,
  ProtoMinor: 1,
  Header:     {
    "Accept-Encoding": []string{
      "gzip, deflate",
    },
    "Accept": []string{
      "text/html,application/xhtml+xml,application/xml;q=0.9",
    },
    "User-Agent": []string{
      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10)",
    },
    "Accept-Language": []string{
      "en-us",
    },
    "Dnt": []string{
      "1",
    },
    "Connection": []string{
      "keep-alive",
    },
  },
  Body: &struct { http.eofReaderWithWriteTo; io.Closer }{
    eofReaderWithWriteTo: http.eofReaderWithWriteTo{
    },
    Closer: ioutil.nopCloser{
      Reader: nil,
    },
  },
  ContentLength:    0,
  TransferEncoding: []string{},
  Close:            false,
  Host:             "localhost",
  Form:             url.Values{},
  PostForm:         url.Values{},
  MultipartForm:    (*multipart.Form)(nil),
  Trailer:          http.Header{},
  RemoteAddr:       "127.0.0.1:54021",
  RequestURI:       "/exit_now",
  TLS:              (*tls.ConnectionState)(nil),
}Finish

*/
