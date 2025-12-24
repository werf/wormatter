package main

import (
	"fmt"
	"strings"
)

func init() {
	fmt.Println("init 1")
}

func init() {
	fmt.Println("init 2")
}

const (
	ConstA      = "a"
	ConstB      = "b"
	ConstMiddle = "m"
	ConstZ      = "z"

	constPrivate = "private"
)

var (
	_ fmt.Stringer = (*Server)(nil)
	_ Reader       = (*Server)(nil)
	_ Writer       = (*Client)(nil)

	GlobalPublic = "public"

	globalA      = 5
	globalB      = 3
	globalMiddle = 7
	globalZ      = 10
	singleConst  = 1
)

type Processor func(input string) (output string, err error)

type Handler func(s string) error

type MyString string

type IntAlias int

type Reader interface {
	Read(p []byte) (n int, err error)
}

type Writer interface {
	Write(p []byte) (n int, err error)
}

type Closer interface {
	Close() error
}

type ReadWriter interface {
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
}

type EmptyInterface interface{}

type Server struct {
	*Client
	Embedded

	Address  string
	Host     string
	MaxConns int

	port    int
	timeout int
}

func NewServer() *Server {
	return &Server{}
}

func NewServerWithOptions(host string, port int) *Server {
	return &Server{Host: host, port: port}
}

func (s *Server) PublicMethod() {}

func (s *Server) AnotherPublic() {
	fmt.Println("another")
}

func (s *Server) privateMethod() {
	return
}

func (s *Server) handleRequest() {}

type Client struct {
	URL string

	name string
}

func NewClient() *Client {
	return &Client{}
}

func NewClientWithTimeout(timeout int) (*Client, error) {
	return nil, nil
}

func (c *Client) Connect() error {
	return nil
}

func (c *Client) disconnect() {
	return
}

type Embedded struct{}

type Config struct {
	Timeout int
	Verbose bool

	debug bool
	name  string
}

func NewConfig() Config {
	return Config{}
}

func NewConfigWithDefaults() *Config {
	return &Config{Timeout: 30, Verbose: true, debug: false, name: "default"}
}

type Empty struct{}

type OnlyEmbedded struct {
	Reader
	fmt.Stringer
}

type OnlyPublic struct {
	Age  int
	Name string
}

type OnlyPrivate struct {
	age  int
	name string
}

type Mixed struct {
	*Client
	Embedded

	Address string
	Name    string

	age   int
	count int
}

type SingleField struct {
	Value int
}

func HelperUpper() {}

func ProcessDataPublic(data string) string {
	return strings.ToLower(data)
}

func helperLower() {
	fmt.Println("helper")
}

func processData(data string) string {
	return strings.ToUpper(data)
}

func createMixed() *Mixed {
	return &Mixed{Address: "addr", Name: "test", age: 25, count: 1}
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

func standaloneHelper() {}

func main() {
	fmt.Println("main")
}
