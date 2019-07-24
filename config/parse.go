package config

import (
	"bufio"
	"io/ioutil"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/pkg/errors"
	"github.com/zclconf/go-cty/cty"
)

func ParseFile(filename string, vars map[string]string) (cfg *Config, err error) {
	var (
		file    *os.File
		content []byte
	)

	file, err = os.Open(filename)
	if err != nil {
		return
	}

	defer file.Close()

	content, err = ioutil.ReadAll(file)
	if err != nil {
		return
	}

	return Parse(content, filename, vars)
}

// Parse parses the contents of a given file `filename`, interpolating
// variables (`vars`), performing not only syntax, but also semantic checks.
//
func Parse(content []byte, filename string, vars map[string]string) (cfg *Config, err error) {
	f, diags := hclsyntax.ParseConfig(content, filename, hcl.Pos{})
	if diags.HasErrors() {
		err = errors.Wrapf(diags, "failed to parse")
		return
	}

	cfg = new(Config)

	diags = gohcl.DecodeBody(f.Body, createEvalContext(vars), cfg)
	if diags.HasErrors() {
		err = errors.Wrapf(diags, "failed to decode")
		return
	}

	return
}

type Lines struct {
	lines []string
}

func NewLines(content string) (l *Lines) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	l = new(Lines)

	for scanner.Scan() {
		l.lines = append(l.lines, scanner.Text())
	}

	return
}

func (l *Lines) At(i int) string {
	return l.lines[i]
}

func (l *Lines) AddLineAt(i int, line string) {
	l.lines = append(l.lines[:i], append([]string{line}, l.lines[i:]...)...)
}

func (l *Lines) String() string {
	return strings.Join(l.lines, "\n")
}

func PrettyDiagnosticFile(filename string, diag *hcl.Diagnostic) (res string) {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	content, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	return PrettyDiagnostic(string(content), diag)

}

// PrettyDiagnostic generates a human-readable pretty diagnostic.
//
func PrettyDiagnostic(content string, diag *hcl.Diagnostic) (res string) {
	var (
		lines     = NewLines(content)
		red       = color.New(color.FgRed, color.Bold).SprintFunc()
		lineBytes = []byte{}
	)

	for i := 1; i < diag.Subject.Start.Column; i++ {
		lineBytes = append(lineBytes, ' ')
	}

	for i := diag.Subject.Start.Column; i < diag.Subject.End.Column; i++ {
		lineBytes = append(lineBytes, '^')
	}

	lines.AddLineAt(diag.Subject.End.Line+1, red(string(lineBytes)))

	res = lines.String()
	return
}

func createEvalContext(vars map[string]string) *hcl.EvalContext {
	var variables = map[string]cty.Value{}

	for key, value := range vars {
		variables[key] = cty.StringVal(value)
	}

	return &hcl.EvalContext{Variables: variables}
}
