// Copyright (c) 2019 MindStand Technologies, Inc
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package gen

//expect .StructName .OtherStructName .StructField .OtherStructField .StructFieldIsMany .OtherStructFieldIsMany
var linkSpec = `
{{ define "linkSpec" }}// LinkTo{{ .OtherStructName }}OnField{{.StructField}} links {{ .StructName }} to {{ .OtherStructName }} on the fields {{ .StructName }}.{{ .StructField }} and {{ .OtherStructName }}.{{ .OtherStructField }}.
// note this uses the special edge {{ .SpecialEdgeType }}
func(l *{{ .StructName }}) LinkTo{{ .OtherStructName }}OnField{{.StructField}}(target *{{ .OtherStructName }}, edge *{{.SpecialEdgeType}}) error {
	if target == nil {
		return errors.New("start and end can not be nil")
	}

	if edge == nil {
		return errors.New("edge can not be nil")
	}
	{{ if .SpecialEdgeDirection }}
	err := edge.SetStartNode(l)
	if err != nil {
		return err
	}
	
	err = edge.SetEndNode(target)
	if err != nil {
		return err
	}{{ else }}
	err := edge.SetStartNode(target)
	if err != nil {
		return err
	}
	
	err = edge.SetEndNode(l)
	if err != nil {
		return err
	}{{ end }}
	{{if .StructFieldIsMany  }}
	if l.{{ .StructField }} == nil {
		l.{{ .StructField }} = make([]*{{ .SpecialEdgeType }}, 1, 1)
		l.{{ .StructField }}[0] = edge
	} else {
		l.{{ .StructField }} = append(l.{{ .StructField }}, edge)
	}{{ else }}
	l.{{ .StructField }} = edge{{ end }}
	{{if .OtherStructFieldIsMany  }}
	if target.{{ .OtherStructField }} == nil {
		target.{{ .OtherStructField }} = make([]*{{ .SpecialEdgeType }}, 1, 1)
		target.{{ .OtherStructField }}[0] = edge
	} else {
		target.{{ .OtherStructField }} = append(target.{{ .OtherStructField }}, edge)
	}{{ else }}
	target.{{ .OtherStructField }} = edge{{ end }}

	return nil
}{{ end }}
`

var singleLink = `
{{ define "linkSingle" }}//LinkTo{{ .OtherStructName }}OnField{{.StructField}} links {{ .StructName }} to {{ .OtherStructName }} on the fields {{ .StructName }}.{{ .StructField }} and {{ .OtherStructName }}.{{ .OtherStructField }}
func(l *{{ .StructName }}) LinkTo{{ .OtherStructName }}OnField{{.StructField}}(target *{{ .OtherStructName }}) error {
	if target == nil {
		return errors.New("start and end can not be nil")
	}
	{{if .StructFieldIsMany  }}
	if l.{{ .StructField }} == nil {
		l.{{ .StructField }} = make([]*{{ .OtherStructName }}, 1, 1)
		l.{{ .StructField }}[0] = target
	} else {
		l.{{ .StructField }} = append(l.{{ .StructField }}, target)
	}{{ else }}
	l.{{ .StructField }} = target{{ end }}
	{{if .OtherStructFieldIsMany  }}
	if target.{{ .OtherStructField }} == nil {
		target.{{ .OtherStructField }} = make([]*{{ .StructName }}, 1, 1)
		target.{{ .OtherStructField }}[0] = l
	} else {
		target.{{ .OtherStructField }} = append(target.{{ .OtherStructField }}, l)
	}{{ else }}
	target.{{ .OtherStructField }} = l{{ end }}

	return nil
}{{ end }}
`

var linkMany = `
{{ define "linkMany" }}// LinkTo{{ .OtherStructName }}OnField{{.StructField}} links {{ .StructName }} to {{ .OtherStructName }} on the fields {{ .StructName }}.{{ .StructField }} and {{ .OtherStructName }}.{{ .OtherStructField }}
func(l *{{ .StructName }}) LinkTo{{ .OtherStructName }}OnField{{.StructField}}(targets ...*{{ .OtherStructName }}) error {
	if targets == nil {
		return errors.New("start and end can not be nil")
	}

	for _, target := range targets {
		{{if .StructFieldIsMany  }}
		if l.{{ .StructField }} == nil {
			l.{{ .StructField }} = make([]*{{ .OtherStructName }}, 1, 1)
			l.{{ .StructField }}[0] = target
		} else {
			l.{{ .StructField }} = append(l.{{ .StructField }}, target)
		}{{ else }}
		l.{{ .StructField }} = target{{ end }}
		{{if .OtherStructFieldIsMany  }}
		if target.{{ .OtherStructField }} == nil {
			target.{{ .OtherStructField }} = make([]*{{ .StructName }}, 1, 1)
			target.{{ .OtherStructField }}[0] = l
		} else {
			target.{{ .OtherStructField }} = append(target.{{ .OtherStructField }}, l)
		}{{ else }}
		target.{{ .OtherStructField }} = l{{ end }}
	}

	return nil
}{{ end }}
`

var unlinkSingle = `
{{ define "unlinkSingle" }}//UnlinkFrom{{ .OtherStructName }}OnField{{.StructField}} unlinks {{ .StructName }} from {{ .OtherStructName }} on the fields {{ .StructName }}.{{ .StructField }} and {{ .OtherStructName }}.{{ .OtherStructField }}
func(l *{{ .StructName }}) UnlinkFrom{{ .OtherStructName }}OnField{{.StructField}}(target *{{ .OtherStructName }}) error {
	if target == nil {
		return errors.New("start and end can not be nil")
	}
	{{if .StructFieldIsMany  }}
	if l.{{ .StructField }} != nil {
		for i, unlinkTarget := range l.{{ .StructField }} {
			if unlinkTarget.UUID == target.UUID {
				a := &l.{{ .StructField }}
				(*a)[i] = (*a)[len(*a)-1]
				(*a)[len(*a)-1] = nil
				*a = (*a)[:len(*a)-1]
				break
			}
		}
	}{{ else }}
	l.{{ .StructField }} = nil{{ end }}
	{{if .OtherStructFieldIsMany  }}
	if target.{{ .OtherStructField }} != nil {
		for i, unlinkTarget := range target.{{ .OtherStructField }} {
			if unlinkTarget.UUID == l.UUID {
				a := &target.{{ .OtherStructField }}
				(*a)[i] = (*a)[len(*a)-1]
				(*a)[len(*a)-1] = nil
				*a = (*a)[:len(*a)-1]
				break
			}
		}
	}{{ else }}
	target.{{ .OtherStructField }} = nil{{ end }}

	return nil
}{{ end }}
`

var unlinkMulti = `
{{ define "unlinkMulti" }}//UnlinkFrom{{ .OtherStructName }}OnField{{.StructField}} unlinks {{ .StructName }} from {{ .OtherStructName }} on the fields {{ .StructName }}.{{ .StructField }} and {{ .OtherStructName }}.{{ .OtherStructField }}
func(l *{{ .StructName }}) UnlinkFrom{{ .OtherStructName }}OnField{{.StructField}}(targets ...*{{ .OtherStructName }}) error {
	if targets == nil {
		return errors.New("start and end can not be nil")
	}

	for _, target := range targets {
		{{if .StructFieldIsMany  }}
		if l.{{ .StructField }} != nil {
			for i, unlinkTarget := range l.{{ .StructField }} {
				if unlinkTarget.UUID == target.UUID {
					a := &l.{{ .StructField }}
					(*a)[i] = (*a)[len(*a)-1]
					(*a)[len(*a)-1] = nil
					*a = (*a)[:len(*a)-1]
					break
				}
			}
		}{{ else }}
		l.{{ .StructField }} = nil{{ end }}
		{{if .OtherStructFieldIsMany  }}
		if target.{{ .OtherStructField }} != nil {
			for i, unlinkTarget := range target.{{ .OtherStructField }} {
				if unlinkTarget.UUID == l.UUID {
					a := &target.{{ .OtherStructField }}
					(*a)[i] = (*a)[len(*a)-1]
					(*a)[len(*a)-1] = nil
					*a = (*a)[:len(*a)-1]
					break
				}
			}
		}{{ else }}
		target.{{ .OtherStructField }} = nil{{ end }}
	}

	return nil
}{{ end }}
`

var unlinkSpec = `
{{ define "unlinkSpec" }}// UnlinkFrom{{ .OtherStructName }}OnField{{.StructField}} unlinks {{ .StructName }} from {{ .OtherStructName }} on the fields {{ .StructName }}.{{ .StructField }} and {{ .OtherStructName }}.{{ .OtherStructField }}.
// also note this uses the special edge {{ .SpecialEdgeType }}
func(l *{{ .StructName }}) UnlinkFrom{{ .OtherStructName }}OnField{{.StructField}}(target *{{ .OtherStructName }}) error {
	if target == nil {
		return errors.New("start and end can not be nil")
	}
	{{if .StructFieldIsMany  }}
	if l.{{ .StructField }} != nil {
		for i, unlinkTarget := range l.{{ .StructField }} {
			{{ if .SpecialEdgeDirection }}
			obj := unlinkTarget.GetStartNode(){{ else }}
			obj := unlinkTarget.GetEndNode(){{end}}

			checkObj, ok := obj.(*{{ .OtherStructName }})
			if !ok {
				return errors.New("unable to cast unlinkTarget to [{{ .OtherStructName }}]")
			}
			if checkObj.UUID == target.UUID {
				a := &l.{{ .StructField }}
				(*a)[i] = (*a)[len(*a)-1]
				(*a)[len(*a)-1] = nil
				*a = (*a)[:len(*a)-1]
				break
			}
		}
	}{{ else }}
	l.{{ .StructField }} = nil{{ end }}
	{{if .OtherStructFieldIsMany  }}
	if target.{{ .OtherStructField }} != nil {
		for i, unlinkTarget := range target.{{ .OtherStructField }} {
			{{ if .SpecialEdgeDirection }}
			obj := unlinkTarget.GetStartNode(){{ else }}
			obj := unlinkTarget.GetEndNode(){{end}}

			checkObj, ok := obj.(*{{ .StructName }})
			if !ok {
				return errors.New("unable to cast unlinkTarget to [{{ .StructName }}]")
			}
			if checkObj.UUID == l.UUID {
				a := &target.{{ .OtherStructField }}
				(*a)[i] = (*a)[len(*a)-1]
				(*a)[len(*a)-1] = nil
				*a = (*a)[:len(*a)-1]
				break
			}
		}
	}{{ else }}
	target.{{ .OtherStructField }} = nil{{ end }}

	return nil
}{{ end }}
`

var masterTpl = `
{{ define "linkFile" }}// Code generated by GoGM v1.0.1. DO NOT EDIT
package {{ .PackageName }}

import (
	"errors"
)
{{range $key, $val := .Funcs}}{{range $val}} {{ if .UsesSpecialEdge }}
{{ template "linkSpec" . }}

{{ template "unlinkSpec" . }}{{ else if .StructFieldIsMany}}
{{template "linkMany" .}}

{{ template "unlinkMulti" .}}{{ else }}
{{ template "linkSingle" .}}

{{ template "unlinkSingle" . }}{{end}} {{end}} {{end}} {{ end }}
`

type templateConfig struct {
	Imports     []string
	PackageName string
	// type: funcs
	Funcs map[string][]*tplRelConf
}

type tplRelConf struct {
	StructName             string
	StructField            string
	OtherStructField       string
	OtherStructName        string
	StructFieldIsMany      bool
	OtherStructFieldIsMany bool

	//stuff for special edges
	UsesSpecialEdge bool
	SpecialEdgeType string
	// StructName = Start if true
	SpecialEdgeDirection bool
}
