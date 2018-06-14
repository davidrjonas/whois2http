package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	listen                        = flag.String("listen", ":43", "target")
	upstream                      = flag.String("upstream", "http://example.com:80/whois?format=plain&query={{query}}", "Upstream to which we should proxy")
	commandPattern, _             = regexp.Compile("^[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9]\\.[a-zA-Z]{2,}\\.?$")
	translateLineEndingPattern, _ = regexp.Compile("([^\r])\n")
)

func mustListen(laddr string) *net.TCPListener {
	tcpaddr, err := net.ResolveTCPAddr("tcp", laddr)
	if err != nil {
		log.Fatalf("Failed to parse addr; address=%s, err=%v", laddr, err)
	}

	server, err := net.ListenTCP("tcp", tcpaddr)

	if err != nil {
		log.Fatalf("Failed to listen; address=%s, err=%v", laddr, err)
	}

	return server
}

type WhoisServer struct {
	upstream      *url.URL
	acceptTimeout time.Duration
	stop          chan bool
	done          chan bool
}

type Handler func(net.Conn) error

func parseUpstreamOpt(opt_val string) *url.URL {
	upstream, err := url.Parse(opt_val)

	if err != nil {
		log.Fatalf("Failed to parse upstream; %v", err)
	}

	return upstream
}

func (s *WhoisServer) Stop() {
	s.stop <- true
	<-s.done
}

func (s *WhoisServer) Serve(listener *net.TCPListener) {
	var wg sync.WaitGroup

OUTER:
	for {
		listener.SetDeadline(time.Now().Add(s.acceptTimeout))

		conn, err := listener.Accept()

		if err != nil {
			if opErr, ok := err.(*net.OpError); !ok || !opErr.Timeout() {
				log.Printf("error accepting connection: %v", err)
			}

			select {
			case <-s.stop:
				break OUTER
			default:
				continue OUTER
			}
		}

		wg.Add(1)

		go func(handle Handler, conn net.Conn, done func()) {
			if err := handle(conn); err != nil {
				log.Printf("connection error: %v", err)
			}
			conn.Close()
			done()
		}(s.handler, conn, wg.Done)
	}

	log.Println("waiting for clients to finish")

	wg.Wait()
	s.done <- true
}

func (s *WhoisServer) handler(conn net.Conn) error {
	buf := bufio.NewReader(conn)
	domain, err := buf.ReadString('\n')
	if err != nil {
		return err
	}

	domain = strings.TrimRight(domain, "\r\n")

	log.Printf("recevied query; query=%v", domain)

	// Validate
	if !commandPattern.MatchString(domain) {
		log.Println("query did not match pattern")
		conn.Write([]byte("Invalid query\r\n"))
		return nil
	}

	// Query backend
	url := strings.Replace(*upstream, "{{query}}", url.QueryEscape(domain), -1)
	resp, err := http.Get(url)
	if err != nil {
		conn.Write([]byte("Upstream query failed\r\n"))
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		conn.Write([]byte("Invalid query\r\n"))
		return fmt.Errorf("Respnose is not 200; code=%d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		conn.Write([]byte("Error reading upstream body\r\n"))
		return err
	}

	conn.Write(translateLineEndingPattern.ReplaceAll(body, []byte("\r\n")))
	conn.Write([]byte("\r\n"))

	return nil
}

func main() {
	flag.Parse()

	whois := WhoisServer{
		upstream:      parseUpstreamOpt(*upstream),
		acceptTimeout: 10 * time.Millisecond,
		stop:          make(chan bool),
		done:          make(chan bool),
	}

	go whois.Serve(mustListen(*listen))

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	<-sigs

	log.Println("Closing listener and waiting for clients to finish")
	whois.Stop()
}