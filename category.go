package main

import (
	"fmt"
	"strings"
)

// A Category is taxon information and metadata. See CatNode and CatTree.
type Category struct {
	Name        string `json:"name"`
	ID          int    `json:"id"`
	ParentID    int    `json:"parent_id"`
	Description string `json:"description"`
	IsVisible   bool   `json:"is_visible"`
}

func (c *Category) IsChild() bool {
	return c.ParentID != 0
}

type CatNode struct {
	C        Category
	Children []*CatNode
}

type CatTree struct {
	Root  *CatNode
	Table map[int]*CatNode
}

func (ct *CatTree) LookupByID(id int) (*CatNode, bool) {
	node, ok := ct.Table[id]
	return node, ok
}

// LookupByPath takes a string with category names (case sensitive) separated by /. Returns *CatNode, bool
func (ct *CatTree) LookupByPath(query string) (node *CatNode, found bool) {
	if query == "" {
		return
	}

	path := strings.Split(query, "/")

	node = ct.Root

	for _, p := range path {
		found = false

		for _, child := range node.Children {
			if child.C.Name == p {
				found = true
				node = child
			}
		}

		if !found {
			node = &CatNode{}
			break
		}
	}

	return
}

func MakeCatTree(cats *[]Category) (*CatTree, error) {
	rootNode := &CatNode{
		C: Category{
			Name:        "root",
			ID:          0,
			ParentID:    0,
			Description: "",
			IsVisible:   false,
		},
		Children: []*CatNode{},
	}

	table := map[int]*CatNode{}
	table[0] = rootNode

	// initalize nodes, populate table
	for _, c := range *cats {
		table[c.ID] = &CatNode{
			C:        c,
			Children: []*CatNode{},
		}

	}

	// use table to build tree
	for _, c := range *cats {
		parent, ok := table[c.ParentID]
		if !ok { // bad parent id
			return &CatTree{}, fmt.Errorf("Category (id: %d) with non-existant parent (id: %d)", c.ID, c.ParentID)
		}

		parent.Children = append(parent.Children, table[c.ID])
	}

	return &CatTree{
		Table: table,
		Root:  rootNode,
	}, nil
}

type categoryListings struct {
	Status int        `json:"status"`
	Data   []Category `json:"data"`
	Meta   struct {
		Pagination pagination `json:"pagination"`
	} `json:"meta"`
}

func (l *categoryListings) CurrentPage() int {
	return l.Meta.Pagination.CurrentPage
}

func (l *categoryListings) TotalPages() int {
	return l.Meta.Pagination.TotalPages
}

func (l *categoryListings) IsLastPage() bool {
	if l.TotalPages() > 0 {
		return l.CurrentPage()/l.TotalPages() == 1
	}

	return true
}
