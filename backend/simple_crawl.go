package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/net/html"
)

func main() {
	testURL := "https://httpbin.org/html"
	
	fmt.Printf("Testing crawling of: %s\n", testURL)
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
	if err != nil {
		log.Fatal("Failed to create request:", err)
	}

	req.Header.Set("User-Agent", "WebsiteAnalyzer/1.0")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Failed to fetch URL:", err)
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d %s\n", resp.StatusCode, resp.Status)

	doc, err := html.Parse(resp.Body)
	if err != nil {
		log.Fatal("Failed to parse HTML:", err)
	}

	h1Count := 0
	h2Count := 0
	
	var count func(*html.Node)
	count = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "h1":
				h1Count++
			case "h2":
				h2Count++
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			count(c)
		}
	}
	
	count(doc)
	
	fmt.Printf("H1 count: %d\n", h1Count)
	fmt.Printf("H2 count: %d\n", h2Count)
	
	fmt.Println("Simple crawling test completed successfully!")
}
