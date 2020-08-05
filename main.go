package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// application logic

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	netClient := &http.Client{
		Timeout: time.Second * 15,
	}

	env := envMap()
	storeHash := getOrFatal("BC_STORE_HASH", env)
	client := Client{
		HTTP:        netClient,
		Store:       storeHash,
		ID:          getOrFatal("BC_CLIENT_ID", env),
		Secret:      getOrFatal("BC_CLIENT_SECRET", env),
		AccessToken: getOrFatal("BC_ACCESS_TOKEN", env),
		catalogURL:  fmt.Sprintf("https://api.bigcommerce.com/stores/%s/v3/catalog", storeHash),
	}

	cats, err := client.GetCategoriesByPage(0)
	if err != nil {
		fmt.Println(err.Error())
	}

}

func getOrFatal(key string, m map[string]string) string {
	val, ok := m[key]
	if !ok {
		log.Fatalf("Key %s is required, but not present. Exiting.", key)
	}
	return val
}

func envMap() map[string]string {
	env := os.Environ()
	m := make(map[string]string, len(env))
	for _, entry := range env {
		pair := strings.Split(entry, "=")
		m[pair[0]] = pair[1]
	}
	return m
}
