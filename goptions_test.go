package goptions

import (
	"os"
	"reflect"
	"testing"
)

func TestParse_StringValue(t *testing.T) {
	var args []string
	var err error
	var fs *FlagSet
	var options struct {
		Name string `goptions:"--name, -n"`
	}
	expected := "SomeName"

	args = []string{"--name", "SomeName"}
	fs = NewFlagSet("goptions", &options)
	err = fs.Parse(args)
	if err != nil {
		t.Fatalf("Flag parsing failed: %s", err)
	}
	if options.Name != expected {
		t.Fatalf("Expected %s for options.Name, got %s", expected, options.Name)
	}

	options.Name = ""

	args = []string{"-n", "SomeName"}
	fs = NewFlagSet("goptions", &options)
	err = fs.Parse(args)
	if err != nil {
		t.Fatalf("Flag parsing failed: %s", err)
	}
	if options.Name != expected {
		t.Fatalf("Expected %s for options.Name, got %s", expected, options.Name)
	}
}

func TestParse_ObligatoryStringValue(t *testing.T) {
	var args []string
	var err error
	var fs *FlagSet
	var options struct {
		Name string `goptions:"-n, obligatory"`
	}
	args = []string{}
	fs = NewFlagSet("goptions", &options)
	err = fs.Parse(args)
	if err == nil {
		t.Fatalf("Parsing should have failed.")
	}

	args = []string{"-n", "SomeName"}
	fs = NewFlagSet("goptions", &options)
	err = fs.Parse(args)
	if err != nil {
		t.Fatalf("Parsing failed: %s", err)
	}

	expected := "SomeName"
	if options.Name != expected {
		t.Fatalf("Expected %s for options.Name, got %s", expected, options.Name)
	}
}

func TestParse_UnknownFlag(t *testing.T) {
	var args []string
	var err error
	var fs *FlagSet
	var options struct {
		Name string `goptions:"--name, -n"`
	}
	args = []string{"-k", "4"}
	fs = NewFlagSet("goptions", &options)
	err = fs.Parse(args)
	if err == nil {
		t.Fatalf("Parsing should have failed.")
	}
}

func TestParse_FlagCluster(t *testing.T) {
	var args []string
	var err error
	var fs *FlagSet
	var options struct {
		Fast    bool `goptions:"-f"`
		Silent  bool `goptions:"-q"`
		Serious bool `goptions:"-s"`
		Crazy   bool `goptions:"-c"`
		Verbose int  `goptions:"-v, accumulate"`
	}
	args = []string{"-fqcvvv"}
	fs = NewFlagSet("goptions", &options)
	err = fs.Parse(args)
	if err != nil {
		t.Fatalf("Parsing failed: %s", err)
	}

	if !(options.Fast &&
		options.Silent &&
		!options.Serious &&
		options.Crazy &&
		options.Verbose == 3) {
		t.Fatalf("Unexpected value: %v", options)
	}
}

func TestParse_MutexGroup(t *testing.T) {
	var args []string
	var err error
	var fs *FlagSet
	var options struct {
		Create bool `goptions:"--create, mutexgroup='action'"`
		Delete bool `goptions:"--delete, mutexgroup='action'"`
	}
	args = []string{"--create", "--delete"}
	fs = NewFlagSet("goptions", &options)
	err = fs.Parse(args)
	if err == nil {
		t.Fatalf("Parsing should have failed.")
	}
}

func TestParse_HelpFlag(t *testing.T) {
	var args []string
	var err error
	var fs *FlagSet
	var options struct {
		Name string `goptions:"--name, -n"`
		Help `goptions:"--help, -h"`
	}
	args = []string{"-n", "SomeNone", "-h"}
	fs = NewFlagSet("goptions", &options)
	err = fs.Parse(args)
	if err != ErrHelpRequest {
		t.Fatalf("Expected ErrHelpRequest, got: %s", err)
	}

	args = []string{"-n", "SomeNone"}
	fs = NewFlagSet("goptions", &options)
	err = fs.Parse(args)
	if err != nil {
		t.Fatalf("Unexpected error returned: %s", err)
	}
}

func TestParse_Verbs(t *testing.T) {
	var args []string
	var err error
	var fs *FlagSet
	var options struct {
		Server string `goptions:"--server, -s"`

		Verbs
		Create struct {
			Name string `goptions:"--name, -n"`
		} `goptions:"create"`
	}

	args = []string{"-s", "127.0.0.1", "create", "-n", "SomeDocument"}
	fs = NewFlagSet("goptions", &options)
	err = fs.Parse(args)
	if err != nil {
		t.Fatalf("Parsing failed: %s", err)
	}

	if !(options.Server == "127.0.0.1" &&
		options.Create.Name == "SomeDocument") {
		t.Fatalf("Unexpected value: %v", options)
	}
}

func TestParse_IntValue(t *testing.T) {
	var args []string
	var err error
	var fs *FlagSet
	var options struct {
		Limit int `goptions:"-l"`
	}

	args = []string{"-l", "123"}
	fs = NewFlagSet("goptions", &options)
	err = fs.Parse(args)
	if err != nil {
		t.Fatalf("Parsing failed: %s", err)
	}

	if !(options.Limit == 123) {
		t.Fatalf("Unexpected value: %v", options)
	}
}

func TestParse_Remainder(t *testing.T) {
	var args []string
	var err error
	var fs *FlagSet
	var options struct {
		Limit int `goptions:"-l"`
		Remainder
	}

	args = []string{"-l", "123", "Something", "SomethingElse"}
	fs = NewFlagSet("goptions", &options)
	err = fs.Parse(args)
	if err != nil {
		t.Fatalf("Parsing failed: %s", err)
	}

	if !(len(options.Remainder) == 2 &&
		options.Remainder[0] == "Something" &&
		options.Remainder[1] == "SomethingElse") {
		t.Fatalf("Unexpected value: %v", options)
	}
}

func TestParse_VerbRemainder(t *testing.T) {
	var args []string
	var err error
	var fs *FlagSet
	var options struct {
		Limit int `goptions:"-l"`
		Remainder

		Verbs
		Create struct {
			Fast bool `goptions:"-f"`
			Remainder
		} `goptions:"create"`
	}

	args = []string{"create", "-f", "Something", "SomethingElse"}
	fs = NewFlagSet("goptions", &options)
	err = fs.Parse(args)
	if err != nil {
		t.Fatalf("Parsing failed: %s", err)
	}

	if !(len(options.Remainder) == 2 &&
		options.Remainder[0] == "Something" &&
		options.Remainder[1] == "SomethingElse") {
		t.Fatalf("Unexpected value: %v", options)
	}
}

func TestParse_NoRemainder(t *testing.T) {
	var args []string
	var err error
	var fs *FlagSet
	var options struct {
		Fast bool `goptions:"-f"`
	}

	args = []string{"-f", "Something", "SomethingElse"}
	fs = NewFlagSet("goptions", &options)
	err = fs.Parse(args)
	if err == nil {
		t.Fatalf("Parsing should have failed")
	}
}

func TestParse_ObligatoryMutexGroup(t *testing.T) {
	var args []string
	var err error
	var fs *FlagSet
	var options struct {
		Create bool `goptions:"-c, mutexgroup='action', obligatory"`
		Delete bool `goptions:"-d, mutexgroup='action'"`
	}

	args = []string{}
	fs = NewFlagSet("goptions", &options)
	err = fs.Parse(args)
	if err == nil {
		t.Fatalf("Parsing should have failed")
	}
	args = []string{"-c", "-d"}
	fs = NewFlagSet("goptions", &options)
	err = fs.Parse(args)
	if err == nil {
		t.Fatalf("Parsing should have failed")
	}
	args = []string{"-d"}
	fs = NewFlagSet("goptions", &options)
	err = fs.Parse(args)
	if err != nil {
		t.Fatalf("Parsing failed: %s", err)
	}
}

func TestParseTag_minimal(t *testing.T) {
	var tag string
	tag = `--name, -n, description='Some name'`
	f, e := parseTag(tag)
	if e != nil {
		t.Fatalf("Tag parsing failed: %s", e)
	}
	expected := &Flag{
		Long:        []string{"name"},
		Short:       []string{"n"},
		Description: "Some name",
	}
	if !reflect.DeepEqual(f, expected) {
		t.Fatalf("Expected %#v, got %#v", expected, f)
	}
}

func TestParseTag_more(t *testing.T) {
	var tag string
	tag = `--name, -n, description='Some name', mutexgroup='selector', obligatory`
	f, e := parseTag(tag)
	if e != nil {
		t.Fatalf("Tag parsing failed: %s", e)
	}
	expected := &Flag{
		Long:        []string{"name"},
		Short:       []string{"n"},
		Accumulate:  false,
		Description: "Some name",
		MutexGroup:  "selector",
		Obligatory:  true,
	}
	if !reflect.DeepEqual(f, expected) {
		t.Fatalf("Expected %#v, got %#v", expected, f)
	}
}

func ExampleFlagSet_PrintHelp() {
	var options struct {
		Server    string `goptions:"-s, --server, obligatory, description='Server to connect to'"`
		Password  string `goptions:"-p, --password, description='Don\\'t prompt for password'"`
		Verbosity int    `goptions:"-v, --verbose, accumulate, description='Set output threshold level'"`
		Help      `goptions:"-h, --help, description='Show this help'"`

		Verbs
		Create struct {
			Name      string `goptions:"-n, --name, obligatory, description='Name of the entity to be created'"`
			Directory bool   `goptions:"--directory, mutexgroup='type', description='Create a directory'"`
			File      bool   `goptions:"--file, mutexgroup='type', description='Create a file'"`
		} `goptions:"create"`
		Delete struct {
			Name      string `goptions:"-n, --name, obligatory, description='Name of the entity to be deleted'"`
			Directory bool   `goptions:"--directory, mutexgroup='type', description='Delete a directory'"`
			File      bool   `goptions:"--file, mutexgroup='type', description='Delete a file'"`
		} `goptions:"delete"`
	}
	args := []string{"--help"}
	fs := NewFlagSet("goptions", &options)
	err := fs.Parse(args)
	if err == ErrHelpRequest {
		fs.PrintHelp(os.Stderr)
		return
	} else if err != nil {
		panic(err)
	}

	// Output:
	// Usage: goptions [global options] <verb> [verb options]
	//
	// Global options:
	//     -s, --server   Server to connect to (*)
	//     -p, --password Don't prompt for password
	//     -v, --verbose  Set output threshold level
	//     -h, --help     Show this help
	//
	// Verbs:
	//     create:
	//         -n, --name      Name of the entity to be created (*)
	//             --directory Create a directory
	//             --file      Create a file
	//     delete:
	//         -n, --name      Name of the entity to be deleted (*)
	//             --directory Delete a directory
	//             --file      Delete a file
}
