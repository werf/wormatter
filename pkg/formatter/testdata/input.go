package main

import (
	"fmt"
	"strings"
)

func helperLower() { fmt.Println("helper") }

func HelperUpper() {}

func main() {
	fmt.Println("main")
}

var globalZ = 10
var globalA = 5
var GlobalPublic = "public"

const ConstZ = "z"
const constPrivate = "private"

type Processor func(input string) (output string, err error)

func (s *Server) privateMethod() { return }

const ConstA = "a"

type Reader interface {
	Read(p []byte) (n int, err error)
}

var _ fmt.Stringer = (*Server)(nil)

func NewServer() *Server { return &Server{} }

type Writer interface {
	Write(p []byte) (n int, err error)
}

type Server struct {
	port    int
	Host    string
	Address string
	*Client
	timeout int
	Embedded
	MaxConns int
}

func (s *Server) PublicMethod() {}

func init() { fmt.Println("init 1") }

var _ Reader = (*Server)(nil)

type Client struct {
	name string
	URL  string
}

func NewClientWithTimeout(timeout int) (*Client, error) { return nil, nil }

type Embedded struct{}

func (s *Server) AnotherPublic() { fmt.Println("another") }

type Handler func(s string) error

func init() {
	fmt.Println("init 2")
}

func NewClient() *Client { return &Client{} }

const (
	ConstMiddle = "m"
	ConstB      = "b"
)

type ReadWriter interface {
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
}

func NewServerWithOptions(host string, port int) *Server {
	return &Server{Host: host, port: port}
}

var (
	globalMiddle = 7
	globalB      = 3
)

func (c *Client) Connect() error { return nil }

type MyString string

func (c *Client) disconnect() { return }

type Closer interface {
	Close() error
}

func processData(data string) string { return strings.ToUpper(data) }

var _ Writer = (*Client)(nil)

func ProcessDataPublic(data string) string {
	return strings.ToLower(data)
}

func (s *Server) handleRequest() {}

type Config struct {
	debug   bool
	Verbose bool
	name    string
	Timeout int
}

func NewConfig() Config { return Config{} }

func NewConfigWithDefaults() *Config {
	return &Config{Verbose: true, Timeout: 30, debug: false, name: "default"}
}

type Empty struct{}

type OnlyEmbedded struct {
	fmt.Stringer
	Reader
}

type OnlyPublic struct {
	Name string
	Age  int
}

type OnlyPrivate struct {
	name string
	age  int
}

type Mixed struct {
	Embedded
	*Client
	Name    string
	Address string
	age     int
	count   int
}

func createMixed() *Mixed {
	return &Mixed{count: 1, Name: "test", age: 25, Address: "addr"}
}

func functionWithReturn() int {
	x := 1
	y := 2
	return x + y
}

func functionWithEarlyReturn(x int) int {
	if x < 0 {
		return 0
	}
	y := x * 2
	return y
}

func functionWithOnlyReturn() int {
	return 42
}

type SingleField struct {
	Value int
}

var singleConst = 1

type EmptyInterface interface{}

type IntAlias int

func standaloneHelper() {}
