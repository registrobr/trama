package trama

import (
	"html/template"
	"testing"
)

func TestInsertInGroupSet(t *testing.T) {
	data := []struct {
		description string
		group       TemplateGroup
		shouldFail  bool
	}{
		{
			description: "It should insert the group properly",
			group: TemplateGroup{
				Name:  "Grupo",
				Files: []string{"um arquivo", "outro arquivo"},
			},
		},
		{
			description: "It should insert this other group properly",
			group: TemplateGroup{
				Name:  "Outro grupo",
				Files: []string{"um arquivo", "outro arquivo"},
			},
		},
		{
			description: "It shouldn't insert another group with the same name",
			group: TemplateGroup{
				Name:  "Grupo",
				Files: []string{"só um arquivo"},
			},
			shouldFail: true,
		},
	}

	set := NewTemplateGroupSet(nil)

	for i, item := range data {
		err := set.Insert(item.group)

		if item.shouldFail {
			if err == nil {
				t.Errorf("Item %d, “%s”: the group was inserted without any error", i, item.description)
			}
		} else {
			if err != nil {
				t.Errorf("Item %d, “%s”: %s", i, item.description, err)
			} else {
				group, found := set.find(item.group.Name)

				if !found {
					t.Errorf("Item %d, “%s”: could not find inserted group", i, item.description)
				} else {
					if !equalGroups(&item.group, group) {
						t.Errorf("Item %d, “%s”: group mismatch. Expecting %+v; found %+v", i, item.description, item.group, *group)
					}
				}
			}
		}
	}
}

func TestSetUnion(t *testing.T) {
	data := []struct {
		description string
		set         TemplateGroupSet
		shouldFail  bool
	}{
		{
			description: "The union should succeed",
			set: TemplateGroupSet{
				elements: map[string]*TemplateGroup{
					"um nome": &TemplateGroup{
						Name:  "um nome",
						Files: []string{"um arquivo", "outro arquivo"},
					},
				},
			},
		},
		{
			description: "The union should succeed, with a merge of the files",
			set: TemplateGroupSet{
				elements: map[string]*TemplateGroup{
					"nome repetido": &TemplateGroup{
						Name:  "nome repetido",
						Files: []string{"um arquivo", "outro arquivo"},
					},
				},
			},
		},
		{
			description: "The union should succeed, with a merge of the functions",
			set: TemplateGroupSet{
				elements: map[string]*TemplateGroup{
					"outro nome": &TemplateGroup{
						Name:  "outro nome",
						Files: []string{"um só arquivo"},
					},
				},
				FuncMap: template.FuncMap{
					"função": func() bool { return false },
				},
			},
		},
		{
			description: "The union should fail due to a function with the same name of another",
			set: TemplateGroupSet{
				elements: map[string]*TemplateGroup{
					"muitos nomes": &TemplateGroup{
						Name:  "muitos nomes",
						Files: []string{"mais um arquivo"},
					},
				},
				FuncMap: template.FuncMap{
					"nome repetido": func() bool { return false },
				},
			},
			shouldFail: true,
		},
	}

	set := TemplateGroupSet{
		elements: map[string]*TemplateGroup{
			"um grupo": &TemplateGroup{
				Name: "um grupo",
			},
			"nome repetido": &TemplateGroup{
				Name:  "nome repetido",
				Files: []string{"este arquivo", "aquele arquivo"},
			},
		},
		FuncMap: template.FuncMap{
			"nome repetido": func() int { return 17 },
		},
	}

	for i, item := range data {
		newSet := copySet(set)
		err := newSet.union(item.set)

		if item.shouldFail {
			if err == nil {
				t.Errorf("Item %d, “%s”: no errors found", i, item.description)
			}
		} else {
			if err != nil {
				t.Errorf("Item %d, “%s”: unexpected error: “%s”", i, item.description, err)
			} else {
				if !subSet(set, newSet) {
					t.Errorf("Item %d, “%s”: the original set is not a subset of the new set. Original: %+v; new: %+v", i, item.description, set, newSet)
				}

				if !subSet(item.set, newSet) {
					t.Errorf("Item %d, “%s”: the item's set is not a subset of the new set. Item's set: %+v; new: %+v", i, item.description, set, newSet)
				}
			}
		}
	}
}

func equalGroups(a, b *TemplateGroup) bool {
	if a.Name != b.Name {
		return false
	}

	if len(a.Files) != len(b.Files) {
		return false
	}

	for i, v := range a.Files {
		if b.Files[i] != v {
			return false
		}
	}

	return true
}

func copySet(set TemplateGroupSet) TemplateGroupSet {
	newSet := NewTemplateGroupSet(make(template.FuncMap))

	for k, v := range set.FuncMap {
		newSet.FuncMap[k] = v
	}

	for _, group := range set.elements {
		files := make([]string, len(group.Files))
		copy(files, group.Files)
		newGroup := TemplateGroup{Name: group.Name, Files: files}
		newSet.Insert(newGroup)
	}

	return newSet
}

func subSet(a, b TemplateGroupSet) bool {
	for k := range a.FuncMap {
		if _, found := b.FuncMap[k]; !found {
			return false
		}
	}

	for k, v := range a.elements {
		if element, found := b.elements[k]; !found {
			return false
		} else {
			if v.Name != element.Name {
				return false
			}

			for _, file := range v.Files {
				found := false

				for _, fileInB := range element.Files {
					if file == fileInB {
						found = true
						break
					}
				}

				if !found {
					return false
				}
			}
		}
	}

	return true
}
