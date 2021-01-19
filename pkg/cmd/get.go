package cmd

import (
	"boltview/pkg/boltdb"
	"io/ioutil"
	"sort"
	"strings"
)

const (
	filterPriority = iota
	exportPriority

	exportOpt = "-e"
	regexOpt  = "-r"
	filterOpt = "-f"

	cmdGet         = "get"
	descriptionGet = "get the specific value via key"
)

func init() {
	RegisterCommand(cmdGet, &get{})
}

type get struct {
	base
	bucket  string
	key     string
	options []option
	data    []byte
	result  interface{}
}

func (g *get) Open(opts ...ParseOption) (Command, error) {
	o := &get{base: base{
		name:        cmdGet,
		cmd:         cmdGet,
		description: descriptionGet,
	}}
	for _, opt := range opts {
		opt(o)
	}
	return o, nil
}

func (g *get) parse(args []string) error {
	if len(args) < 2 || !strings.Contains(args[1], ".") || len(args)%2 != 0 {
		return ErrInvalidParams
	}
	g.params = args
	g.options = nil
	field := strings.Split(args[1], ".")
	g.bucket = field[0]
	g.key = field[1]

	for i := 2; i < len(args); i += 2 {
		v, ok := optionMap[args[i]]
		if !ok {
			return ErrInvalidParams
		}
		v.set(args[i+1])
		g.options = append(g.options, v)
	}
	return nil
}

func (g *get) ok() {
	writeToConsole(g.result)
}

func (g *get) exec() error {
	var err error
	g.data, err = boltdb.Get(g.bucket, g.key)
	if err != nil {
		return err
	}
	g.result = string(g.data)
	sort.Slice(g.options, func(i, j int) bool {
		return g.options[i].Priority() < g.options[j].Priority()
	})
	for _, option := range g.options {
		if err := option.Do(g); err != nil {
			return err
		}
	}
	return nil
}

var optionMap = map[string]option{
	exportOpt: &export{priority: exportPriority},
}

type option interface {
	Priority() int
	set(string)
	Do(*get) error
}

type export struct {
	priority int
	file     string
}

func (e *export) set(s string) {
	e.file = s
}

func (e *export) Do(g *get) error {
	g.result = "ok"
	return ioutil.WriteFile(e.file, g.data, 0666)
}

func (e *export) Priority() int {
	return e.priority
}
