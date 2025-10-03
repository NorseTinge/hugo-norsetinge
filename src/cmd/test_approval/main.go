package main

import (
	"fmt"
	"log"
	"norsetinge/src/approval"
	"norsetinge/src/common"
	"norsetinge/src/config"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run test_approval.go <article-file>")
		os.Exit(1)
	}

	articlePath := os.Args[1]

	// Load config
	cfg, err := config.Load("../../../config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Parse article
	article, err := common.ParseArticle(articlePath)
	if err != nil {
		log.Fatalf("Failed to parse article: %v", err)
	}

	fmt.Printf("Article: %s by %s\n", article.Title, article.Author)
	fmt.Printf("Status: publish=%d\n", article.Status.Publish)

	// Create approval server
	server := approval.NewServer(cfg)

	// Trigger approval
	fmt.Println("\nTriggering approval process...")
	if err := server.RequestApproval(article); err != nil {
		log.Fatalf("Approval failed: %v", err)
	}

	fmt.Println("✅ Approval request sent successfully!")
	fmt.Printf("Check your email at: %s\n", cfg.Email.ApprovalRecipient)
	fmt.Printf("Preview location: %s/godkendelse/%s/\n", filepath.Dir(cfg.Dropbox.BasePath), article.GetSlug())
}
