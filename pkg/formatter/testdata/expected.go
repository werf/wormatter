package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
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
	AlphaVal = "alpha" // alpha inline
	// AppVersion is a documented constant.
	AppVersion  = "2.0"
	BetaVal     = "beta" // beta inline
	ConstA      = "a"
	ConstB      = "b"
	ConstMiddle = "m"
	// Test: consts should be merged and sorted
	ConstZ = "z"
	// Test: inline comments on consts are preserved during reorder
	ZetaVal = "zeta" // zeta inline

	SeverityCritical Severity = "critical"
	SeverityWarning  Severity = "warning"
	SeverityInfo     Severity = "info"

	StatusOK      StatusCode = "ok"
	StatusError   StatusCode = "error"
	StatusPending StatusCode = "pending"

	constPrivate = "private"
	// internalBuild is private.
	internalBuild = "abc123"
)

// Test: iota const block should stay separate
const (
	PriorityLow Priority = iota
	PriorityMedium
	PriorityHigh
)

var (
	// Test: blank var interface check
	_ fmt.Stringer = (*Server)(nil)
	_ Reader       = (*Server)(nil)
	_ Writer       = (*Client)(nil)

	GlobalPublic = "public"

	// Test: custom type grouping in var block
	DefaultStatus StatusCode = "default"
	ErrorStatus   StatusCode = "error"

	// Test: doc comments on individual vars/consts are preserved during merge
	// EntryPoints defines supported entry points.
	EntryPoints = []string{
		"index.ts",
		"index.js",
	}
	// MaxAttempts is the max retry count.
	MaxAttempts = 5
	// Test: variable declaration order preserved for init dependencies
	Registry     = []*Feature{}
	FeatureAlpha = NewFeature("alpha")
	FeatureBeta  = NewFeature("beta")

	// Test: vars should be merged and sorted
	globalZ      = 10
	globalA      = 5
	globalMiddle = 7
	globalB      = 3
	singleConst  = 1
	// Test: slice of anonymous structs with positional literals
	sliceOfStructs = []struct {
		path    string
		content string
	}{
		{path: filepath.Join("a", "b"), content: "content1"},
		{path: filepath.Join("c", "d"), content: "content2"},
	}
	// defaultDelay is the default delay.
	defaultDelay = 100
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

// Test: typed consts with multiple custom types preserve original order
type Severity string

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
	return &Server{Host: host, port: port}
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
	return &Config{Timeout: 30, Verbose: true, debug: false, name: "default"}
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
	return &myPrivateType{value: 1}
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

// Test: structs with encoding tags (json/yaml) should NOT be reordered
type APIResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// Test: structs with non-encoding tags should still be reordered
type ValidatedInput struct {
	Alpha string `validate:"required"`
	Zulu  string `validate:"required"`

	bravo int
}

// Test: spurious blank lines between struct fields should be removed
type SpacedStruct struct {
	client   string
	host     string
	logStore string
	maxWidth int
}

type Feature struct {
	Name string
}

func NewFeature(name string) *Feature {
	f := &Feature{Name: name}
	Registry = append(Registry, f)

	return f
}

func HelperUpper() {}

func ProcessDataPublic(data string) string {
	return strings.ToLower(data)
}

// Test: anonymous struct with positional literal
func createAnonymous() interface{} {
	return struct {
		B int
		A string
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
	return &PositionalTest{Age: 35, City: "Boston", Name: "Alice"}
}

// Test: struct literal field reordering
func createMixed() *Mixed {
	return &Mixed{Address: "addr", Name: "test", age: 25, count: 1}
}

func createPositional() *PositionalTest {
	return &PositionalTest{Age: 30, City: "NYC", Name: "John"}
}

func createPositionalPartial() *PositionalTest {
	return &PositionalTest{Age: 25, Name: "Jane"}
}

func createWithEmbedded() *WithEmbedded {
	return &WithEmbedded{PositionalTest: PositionalTest{Age: 40, City: "LA", Name: "Bob"}, Extra: "extra"}
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
