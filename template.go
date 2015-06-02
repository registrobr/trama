package trama

import (
	"fmt"
	"html/template"
	"io"
)

// A TemplateGroup groups templates by a name. It can be used for organise
// templates by language, for instance, if one needs support for localisation.
type TemplateGroup struct {
	// Name is the given name of a group. If one is grouping its templates by
	// language, Name can be “pt” or “en” for example.
	Name string

	// Files is the list of files contained in the group.
	Files []string

	templ *template.Template
}

func (t *TemplateGroup) merge(other *TemplateGroup) {
	t.Files = append(t.Files, other.Files...)
}

func (t *TemplateGroup) parse(leftDelim, rightDelim string, funcMap template.FuncMap) (err error) {
	t.templ = template.New(t.Name)

	if leftDelim != "" || rightDelim != "" {
		t.templ = t.templ.Delims(leftDelim, rightDelim)
	}

	if funcMap != nil {
		t.templ = t.templ.Funcs(funcMap)
	}

	t.templ, err = t.templ.ParseFiles(t.Files...)
	return err
}

func (t *TemplateGroup) executeTemplate(w io.Writer, name string, data interface{}) error {
	return t.templ.ExecuteTemplate(w, name, data)
}

// A TemplateGroupSet is a set of TemplateGroups. The set is indexed by the
// TemplateGroup name. Each handler registers a TemplateGroupSet to be used
// when a request arrives, using its Templates method.
type TemplateGroupSet struct {
	// FuncMap is a template.FuncMap that can be optionally registered to be
	// available for the templates in the group set.
	FuncMap  template.FuncMap
	elements map[string]*TemplateGroup
}

// NewTemplateGroupSet creates a new TemplateGroupSet using a possibly nil
// FuncMap.
func NewTemplateGroupSet(f template.FuncMap) TemplateGroupSet {
	return TemplateGroupSet{elements: make(map[string]*TemplateGroup), FuncMap: f}
}

// Insert inserts a new template group in the set.
func (t *TemplateGroupSet) Insert(g TemplateGroup) error {
	if _, found := t.elements[g.Name]; found {
		return fmt.Errorf("Another template group with the name “%s” is already registered", g.Name)
	}

	t.elements[g.Name] = &g
	return nil
}

func (t *TemplateGroupSet) find(groupName string) (group *TemplateGroup, found bool) {
	group, found = t.elements[groupName]
	return
}

func (t *TemplateGroupSet) union(other TemplateGroupSet) error {
	if len(other.FuncMap) > 0 && t.FuncMap == nil {
		t.FuncMap = make(template.FuncMap)
	}

	for k, v := range other.FuncMap {
		if _, found := t.FuncMap[k]; found {
			return fmt.Errorf("Function “%s” already registered", k)
		}

		t.FuncMap[k] = v
	}

	for name, otherGroup := range other.elements {
		if group, found := t.elements[name]; found {
			group.merge(otherGroup)
		} else {
			t.elements[name] = otherGroup
		}
	}

	return nil
}

func (t *TemplateGroupSet) parse(leftDelim, rightDelim string) error {
	for _, group := range t.elements {
		err := group.parse(leftDelim, rightDelim, t.FuncMap)

		if err != nil {
			return err
		}
	}

	return nil
}
