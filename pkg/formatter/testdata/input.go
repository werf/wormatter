package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// Test: functions should be reordered (main last, init first after imports)
func helperLower() { fmt.Println("helper") }

func HelperUpper() {}

func main() {
	fmt.Println("main")
}

// Test: vars should be merged and sorted
var globalZ = 10
var globalA = 5
var GlobalPublic = "public"

// Test: consts should be merged and sorted
const ConstZ = "z"
const constPrivate = "private"

// Test: type declared in wrong place
type Processor func(input string) (output string, err error)

// Test: method declared before its type
func (s *Server) privateMethod() { return }

const ConstA = "a"

type Reader interface {
	Read(p []byte) (n int, err error)
}

// Test: blank var interface check
var _ fmt.Stringer = (*Server)(nil)

// Test: constructor declared before type
func NewServer() *Server { return &Server{} }

type Writer interface {
	Write(p []byte) (n int, err error)
}

// Test: struct fields should be reordered (embedded, public, private)
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

// Test: init should be moved up
func init() { fmt.Println("init 1") }

var _ Reader = (*Server)(nil)

// Test: struct fields in wrong order
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

// Test: struct literal fields should be reordered
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

// Test: struct fields reordering
type Config struct {
	debug   bool
	Verbose bool
	name    string
	Timeout int
}

func NewConfig() Config { return Config{} }

// Test: struct literal reordering
func NewConfigWithDefaults() *Config {
	return &Config{Verbose: true, Timeout: 30, debug: false, name: "default"}
}

type Empty struct{}

// Test: embedded fields should be sorted
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

// Test: mixed struct fields
type Mixed struct {
	Embedded
	*Client
	Name    string
	Address string
	age     int
	count   int
}

// Test: struct literal field reordering
func createMixed() *Mixed {
	return &Mixed{count: 1, Name: "test", age: 25, Address: "addr"}
}

// Test: blank line before return
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

// Test: custom type grouping in const block
type StatusCode MyString

const (
	StatusOK      StatusCode = "ok"
	StatusError   StatusCode = "error"
	StatusPending StatusCode = "pending"
)

type Priority int

// Test: iota const block should stay separate
const (
	PriorityLow Priority = iota
	PriorityMedium
	PriorityHigh
)

// Test: custom type grouping in var block
var (
	DefaultStatus StatusCode = "default"
	ErrorStatus   StatusCode = "error"
)

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

// Test: no blank lines between select cases
func functionWithSelect(ch chan int) {
	select {

	case v := <-ch:
		fmt.Println(v)

	default:
		fmt.Println("no value")
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

// Test: unexported constructor matching
type myPrivateType struct {
	value int
}

func newMyPrivateType() *myPrivateType {
	return &myPrivateType{value: 1}
}

// Test: positional literals should be converted to keyed
type PositionalTest struct {
	Name string
	Age  int
	City string
}

func createPositional() *PositionalTest {
	return &PositionalTest{"John", 30, "NYC"}
}

func createPositionalPartial() *PositionalTest {
	return &PositionalTest{"Jane", 25}
}

// Test: anonymous struct with positional literal
func createAnonymous() interface{} {
	return struct {
		B int
		A string
	}{42, "hello"}
}

// Test: embedded fields in positional literal
type WithEmbedded struct {
	PositionalTest
	Extra string
}

func createWithEmbedded() *WithEmbedded {
	return &WithEmbedded{PositionalTest{"Bob", 40, "LA"}, "extra"}
}

// Test: external struct literal should NOT be touched
func createExternal() *os.File {
	// This uses positional but type is external - leave untouched
	// (os.File doesn't actually support this, so use a keyed example)
	return nil
}

// Test: already keyed literal - no change
func createKeyed() *PositionalTest {
	return &PositionalTest{Name: "Alice", Age: 35, City: "Boston"}
}

// Test: empty literal - no change
func createEmpty() *PositionalTest {
	return &PositionalTest{}
}

// Test: multi-line func signature should collapse to single line
func multiLineFunc(
	a int,
	b string,
	c bool,
) error {
	return nil
}

// Test: multi-line method signature should collapse
func (s *Server) multiLineMethod(
	ctx context.Context,
	input string,
) (string, error) {
	return input, nil
}

// Test: multi-line return values should collapse
func multiLineReturns() (
	result string,
	err error,
) {
	return "", nil
}

// Test: function type should collapse
type MultiLineHandler func(
	w http.ResponseWriter,
	r *http.Request,
)

// Test: interface method should collapse
type MultiLineInterface interface {
	Process(
		ctx context.Context,
		input Input,
	) (Output, error)
}

// Types for interface test
type Input struct{}
type Output struct{}
