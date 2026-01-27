package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// Test: init should be moved up
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

	StatusError   StatusCode = "error"
	StatusOK      StatusCode = "ok"
	StatusPending StatusCode = "pending"

	constPrivate = "private"
)

// Test: iota const block should stay separate
const (
	PriorityLow Priority = iota
	PriorityMedium
	PriorityHigh
)

var (
	_ fmt.Stringer = (*Server)(nil)
	_ Reader       = (*Server)(nil)
	_ Writer       = (*Client)(nil)

	GlobalPublic = "public"

	DefaultStatus StatusCode = "default"
	ErrorStatus   StatusCode = "error"

	globalA      = 5
	globalB      = 3
	globalMiddle = 7
	globalZ      = 10
	singleConst  = 1
)

// Test: type declared in wrong place
type Processor func(input string) (output string, err error)

type Handler func(s string) error

type MyString string

type IntAlias int

// Test: custom type grouping in const block
type StatusCode MyString

type Priority int

// Test: function type should collapse
type MultiLineHandler func(w http.ResponseWriter, r *http.Request)

type Reader interface {
	Read(p []byte) (n int, err error)
}

type Writer interface {
	Write(p []byte) (n int, err error)
}

type Closer interface {
	Close() error
}

// Test: interface method should collapse
type MultiLineInterface interface {
	Process(ctx context.Context, input Input) (Output, error)
}

type ReadWriter interface {
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
}

type EmptyInterface interface{}

// Test: struct fields should be reordered (embedded, public, private)
type Server struct {
	*Client
	Embedded

	Address  string
	Host     string
	MaxConns int

	port    int
	timeout int
}

// Test: constructor declared before type
func NewServer() *Server {
	return &Server{}
}

// Test: struct literal fields should be reordered
func NewServerWithOptions(host string, port int) *Server {
	return &Server{
		Host: host, port: port,
	}
}

func (s *Server) AnotherPublic() {
	fmt.Println("another")
}

func (s *Server) PublicMethod() {}

func (s *Server) handleRequest() {}

// Test: multi-line method signature should collapse
func (s *Server) multiLineMethod(ctx context.Context, input string) (string, error) {
	return input, nil
}

// Test: method declared before its type
func (s *Server) privateMethod() {
	return
}

// Test: struct fields in wrong order
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

// Test: struct fields reordering
type Config struct {
	Timeout int
	Verbose bool

	debug bool
	name  string
}

func NewConfig() Config {
	return Config{}
}

// Test: struct literal reordering
func NewConfigWithDefaults() *Config {
	return &Config{
		Timeout: 30, Verbose: true, debug: false, name: "default",
	}
}

type Empty struct{}

// Test: embedded fields should be sorted
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

// Test: mixed struct fields
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

// Test: unexported constructor matching
type myPrivateType struct {
	value int
}

func newMyPrivateType() *myPrivateType {
	return &myPrivateType{
		value: 1,
	}
}

// Test: positional literals should be converted to keyed
type PositionalTest struct {
	Age  int
	City string
	Name string
}

// Test: embedded fields in positional literal
type WithEmbedded struct {
	PositionalTest

	Extra string
}

// Types for interface test
type Input struct{}

type Output struct{}

func HelperUpper() {}

func ProcessDataPublic(data string) string {
	return strings.ToLower(data)
}

// Test: anonymous struct with positional literal
func createAnonymous() interface{} {
	return struct {
		A string
		B int
	}{B: 42, A: "hello"}
}

// Test: empty literal - no change
func createEmpty() *PositionalTest {
	return &PositionalTest{}
}

// Test: external struct literal should NOT be touched
func createExternal() *os.File {
	// This uses positional but type is external - leave untouched
	// (os.File doesn't actually support this, so use a keyed example)
	return nil
}

// Test: already keyed literal - no change
func createKeyed() *PositionalTest {
	return &PositionalTest{
		Age: 35, City: "Boston", Name: "Alice",
	}
}

// Test: struct literal field reordering
func createMixed() *Mixed {
	return &Mixed{
		Address: "addr", Name: "test", age: 25, count: 1,
	}
}

func createPositional() *PositionalTest {
	return &PositionalTest{
		Age: 30, City: "NYC", Name: "John",
	}
}

func createPositionalPartial() *PositionalTest {
	return &PositionalTest{
		Age: 25, Name: "Jane",
	}
}

func createWithEmbedded() *WithEmbedded {
	return &WithEmbedded{
		PositionalTest: PositionalTest{
			Age: 40, City: "LA", Name: "Bob",
		}, Extra: "extra",
	}
}

// Test: blank line before comments
func functionWithComment() {
	x := 1

	// This is a comment about y
	y := 2
	z := x + y

	// Another comment
	// spanning multiple lines
	fmt.Println(z)
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

// Test: blank line before return
func functionWithReturn() int {
	x := 1
	y := 2

	return x + y
}

// Test: no blank lines between select cases
func functionWithSelect(ch chan int) {
	select {
	case v := <-ch:
		fmt.Println(v)
	default:
		fmt.Println("no value")
	}
}

// Test: no blank lines between switch cases
func functionWithSwitch(x int) string {
	switch x {
	case 1:
		return "one"
	case 2:
		return "two"
	default:
		return "other"
	}
}

// Test: type switch case spacing
func functionWithTypeSwitch(x interface{}) string {
	switch x.(type) {
	case int:
		return "int"
	case string:
		return "string"
	default:
		return "unknown"
	}
}

// Test: functions should be reordered (main last, init first after imports)
func helperLower() {
	fmt.Println("helper")
}

// Test: multi-line func signature should collapse to single line
func multiLineFunc(a int, b string, c bool) error {
	return nil
}

// Test: multi-line return values should collapse
func multiLineReturns() (result string, err error) {
	return "", nil
}

func processData(data string) string {
	return strings.ToUpper(data)
}

func standaloneHelper() {}

func main() {
	fmt.Println("main")
}
