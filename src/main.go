package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"norsetinge/approval"
	"norsetinge/config"
	"norsetinge/watcher"
)

func main() {
	fmt.Println("Norsetinge - Automated Multilingual News Service")
	fmt.Println("================================================")

	// Load config
	cfg, err := config.Load("../config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Loaded config: monitoring %s", cfg.Dropbox.BasePath)

	// Create approval server
	approvalServer := approval.NewServer(cfg)

	// Start approval server in background
	go func() {
		log.Printf("Starting approval server on port %d", cfg.Approval.Port)
		if err := approvalServer.Start(); err != nil {
			log.Fatalf("Approval server failed: %v", err)
		}
	}()

	// Create watcher
	w, err := watcher.NewWatcher(cfg, "../folder-aliases.yaml")
	if err != nil {
		log.Fatalf("Failed to create watcher: %v", err)
	}

	// Connect approval server to watcher
	w.SetApprovalServer(approvalServer)

	// Connect mover to approval server so it can move files
	approvalServer.SetMover(w.GetMover())

	// Start watching
	if err := w.Start(); err != nil {
		log.Fatalf("Failed to start watcher: %v", err)
	}

	log.Println("Watcher started. Monitoring for article changes...")

	// Start IMAP email monitoring in background
	imapReader := approval.NewIMAPReader(cfg)
	go func() {
		log.Println("Starting IMAP email monitoring...")
		if err := imapReader.StartMonitoring(30*time.Second, func(reply *approval.EmailReply) {
			log.Printf("📧 Email reply received: Action=%v, Subject=%s", reply.Action, reply.Subject)
			if err := approvalServer.HandleEmailReply(reply); err != nil {
				log.Printf("Error processing email reply: %v", err)
			}
		}); err != nil {
			log.Printf("IMAP monitoring failed: %v", err)
		}
	}()

	log.Println("Press Ctrl+C to stop")

	// Listen for events
	go func() {
		for event := range w.Events() {
			log.Printf("📄 Event: %v - %s", event.Type, event.FilePath)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	w.Stop()
}
