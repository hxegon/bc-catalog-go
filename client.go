package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func debugJSON(bs []byte) {
	var debugBuf bytes.Buffer
	json.Indent(&debugBuf, bs, "", "\t")
	log.Println(string(debugBuf.Bytes()))
}

// Want to spike the bigcommerce client library and extract it once I have some working prototype code

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
