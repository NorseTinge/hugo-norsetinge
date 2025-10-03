package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"norsetinge/approval"
	"norsetinge/builder"
	"norsetinge/config"
	"norsetinge/deployer"
	"norsetinge/watcher"
)

func main() {
	fmt.Println("Norsetinge - Automated Multilingual News Service")
	fmt.Println("================================================")

	// Load config (use absolute path)
	configPath := "/home/ubuntu/hugo-norsetinge/config.yaml"
	cfg, err := config.Load(configPath)
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
	aliasesPath := "/home/ubuntu/hugo-norsetinge/folder-aliases.yaml"
	w, err := watcher.NewWatcher(cfg, aliasesPath)
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

	// Start periodic build+deploy (every 10 minutes)
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()

		hugoBuilder := builder.NewHugoBuilder(cfg)
		dep := deployer.NewDeployer(cfg)

		for range ticker.C {
			log.Printf("⏰ Running periodic build+deploy...")

			// Build full site
			publicDir, mirrorDir, err := hugoBuilder.BuildFullSite()
			if err != nil {
				log.Printf("Error in periodic build: %v", err)
				continue
			}

			// Deploy
			if err := dep.Deploy(publicDir, mirrorDir); err != nil {
				log.Printf("Error in periodic deploy: %v", err)
				continue
			}

			log.Printf("✅ Periodic build+deploy completed")
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
