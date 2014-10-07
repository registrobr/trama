package trama

import (
	"fmt"
	"html/template"
	"io"
)

type TemplateGroupSet map[string]*TemplateGroup

func NewTemplateGroupSet() TemplateGroupSet {
	return make(TemplateGroupSet)
}

func (t TemplateGroupSet) Insert(g TemplateGroup) error {
	if _, found := t[g.Name]; found {
		return fmt.Errorf("Another template group with the name “%s” is already registered", g.Name)
	}

	t[g.Name] = &g
	return nil
}

func (t TemplateGroupSet) union(other TemplateGroupSet) error {
	for name, otherGroup := range other {
		if group, found := t[name]; found {
			err := group.merge(otherGroup)

			if err != nil {
				return err
			}
		} else {
			t[name] = otherGroup
		}
	}

	return nil
}

func (t TemplateGroupSet) parse(leftDelim, rightDelim string) error {
	for _, group := range t {
		err := group.parse(leftDelim, rightDelim)

		if err != nil {
			return err
		}
	}

	return nil
}

type TemplateGroup struct {
	Name    string
	Files   []string
	FuncMap template.FuncMap

	templ *template.Template
}

func (t *TemplateGroup) merge(other *TemplateGroup) error {
	t.Files = append(t.Files, other.Files...)

	for k, v := range other.FuncMap {
		if _, found := t.FuncMap[k]; found {
			return fmt.Errorf("Function “%s” already registered in group “%s”", k, t.Name)
		}

		t.FuncMap[k] = v
	}

	return nil
}

func (t *TemplateGroup) parse(leftDelim, rightDelim string) (err error) {
	t.templ = template.New(t.Name)

	if leftDelim != "" || rightDelim != "" {
		t.templ = t.templ.Delims(leftDelim, rightDelim)
	}

	if t.FuncMap != nil {
		t.templ = t.templ.Funcs(t.FuncMap)
	}

	t.templ, err = t.templ.ParseFiles(t.Files...)
	return err
}

func (t *TemplateGroup) executeTemplate(w io.Writer, name string, data interface{}) error {
	return t.templ.ExecuteTemplate(w, name, data)
}
