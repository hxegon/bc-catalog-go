package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

type pagination struct {
	Total       int `json:"total"`
	Count       int `json:"count"`
	PerPage     int `json:"per_page"`
	CurrentPage int `json:"current_page"`
	TotalPages  int `json:"total_pages"`
	Links       struct {
		Next    string `json:"next"`
		Current string `json:"current"`
	} `json:"links"`
}

func debugJSON(bs []byte) {
	var debugBuf bytes.Buffer
	json.Indent(&debugBuf, bs, "", "\t")
	fmt.Println(string(debugBuf.Bytes()))
}

func addURLParams(raw string, params map[string]string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	q, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return "", err
	}

	for k, v := range params {
		q.Add(k, v)
	}
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// Want to spike the bigcommerce client library and extract it once I have some working prototype code

// Client for BigCommerce catalog API
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

// GetCategoriesByPage returns the json response with the categories and category data, and an error
func (c *Client) GetCategoriesByPage(page int) ([]Category, error) {
	endpoint := fmt.Sprintf("%s/%s", c.catalogURL, "categories")
	url, err := addURLParams(endpoint, map[string]string{"page": fmt.Sprint(page)})
	if err != nil {
		return nil, err
	}

	req, err := c.newRequest("GET", url, nil)
	if err != nil {
		return nil, err
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

	// CLEANUP
	if len(body) == 0 {
		return []Category{}, nil
	}

	var listings categoryListings

	// parse json bytes into categories
	err = json.Unmarshal(body, &listings)
	if err != nil {
		return nil, err
	}

	if status := listings.Status; status != 200 && status != 0 {
		return listings.Data, fmt.Errorf("Request returned non-successful status code %d", status)
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

func (c *Client) GetCatTree() (*CatTree, error) {
	nilct := &CatTree{}

	cats, err := c.GetAllCategories()
	if err != nil {
		return nilct, err
	}

	ct, err := MakeCatTree(&cats)
	if err != nil {
		return nilct, err
	}

	return ct, nil
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
