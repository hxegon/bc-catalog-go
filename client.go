package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func debugJSON(bs []byte) {
	var debugBuf bytes.Buffer
	json.Indent(&debugBuf, bs, "", "\t")
	log.Println(string(debugBuf.Bytes()))
}

// Want to spike the bigcommerce client library and extract it once I have some working prototype code

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
	C        *Category
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

func (ct *CatTree) LookupByPath(query string) (last *CatNode, found bool) {
	if query == "" {
		return
	}

	path := strings.Split(query, "/")

	last = ct.Root

	for _, p := range path {
		found = false

		for _, child := range last.Children {
			if child.C.Name == p {
				found = true
				last = child
			}
		}

		if !found {
			last = &CatNode{}
			break
		}
	}

	return
}

func MakeCatTree(cats []*Category) (*CatTree, error) {
	rootNode := &CatNode{
		C: &Category{
			Name:        "root",
			ID:          0,
			ParentID:    0,
			Description: "",
			IsVisible:   false,
		},
		Children: []*CatNode{},
	}

	var table map[int]*CatNode
	table[0] = rootNode

	// initalize nodes, populate table
	for _, c := range cats {
		node := &CatNode{
			C:        c,
			Children: []*CatNode{},
		}

		table[c.ID] = node
	}

	// use table to build tree
	for _, c := range cats {
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

// Add puts a Category into a CatMap, returning an "Overwritten" bool and an Error

// func (cm *CatMap) Lookup(id int) (Category, bool)

// func (cm *CatMap) Query(query string) (Category, bool)

type categoryListings struct {
	Status int        `json:"status"`
	Data   []Category `json:"data"`
}

// Bigcommerce catalogue client
type Client struct {
	HTTP        *http.Client
	Store       string
	ID          string
	Secret      string
	AccessToken string
	catalogURL  string
}

// newRequest forwards arguments to http.NewRequest, then sets Content-Type, Accept, and auth id, token headers
func (c *Client) newRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return &http.Request{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Auth-Client", c.ID)
	req.Header.Set("X-Auth-Token", c.AccessToken)

	return req, nil
}

// func NewClient(http http.Client, env map[string]string) (Client, error) {
// }

// GetCategoriesByPage returns the json response with the categories and category data, and an error
func (c *Client) GetCategoriesByPage(page int) ([]Category, error) {
	endpoint := fmt.Sprintf("%s/%s", c.catalogURL, "categories")
	req, err := c.newRequest("GET", endpoint, nil)
	if err != nil {
		return nil, nil
	}

	// Send request and get response from bigcommerce
	res, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	if res.Body != nil {
		defer res.Body.Close()
	}

	// get bytes from response body reader
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var listings categoryListings

	// parse json bytes into categories
	err = json.Unmarshal(body, &listings)
	if err != nil {
		return nil, err
	}

	return listings.Data, nil
}

// GetAllCategories abstracts over the pagination of GetCategoriesByPage
func (c *Client) GetAllCategories() ([]Category, error) {
	var all []Category
	pageIndex := 1
	for {
		page, err := c.GetCategoriesByPage(pageIndex)
		if err != nil {
			return nil, err
		}

		// CHECK: If you ask for a page which wouldn't contain any categories:
		// i.e. with 50 categories and 10 categories per page you ask for page 6
		// Does bigcommerce return an empty page or return with error?
		// if it returns with error is that something we should abstract in GetCategoriesByPage
		if len(page) == 0 {
			break
		}

		all = append(all, page...)
		pageIndex++
	}

	return all, nil
}

// Reference: https://developer.bigcommerce.com/api-reference/catalog/catalog-api/products/createproduct
// func (c *Client) CreateProduct() (ProductID, error)
/* Required info for  request
Name: string
Price: floatstring
Categories: []CategoryId(int)
Weight: Int
Type: "physical"

- Adding an Image is a separate request (See: (c *Client) AddProductImage)
Does this update an existing product?
*/

// func (c *Client) AddProductImage(ProductID, image_url, desc)
