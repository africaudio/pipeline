package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/mailgun/mailgun-go/v4"
)

type PushEvent struct {
	Ref string `json:"ref"`
}

func main() {
	http.HandleFunc("/webhook", handleWebhook)
	log.Println("Webhook server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var event PushEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("Error parsing JSON: %v", err)
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	branch := strings.TrimPrefix(event.Ref, "refs/heads/")
	if branch == "main" || branch == "master" {
		go func() {
			if err := buildProject(); err != nil {
				log.Printf("Build failed: %v", err)
				sendEmail("Build Failed", fmt.Sprintf("Build failed for branch %s: %v", branch, err))
			} else {
				sendEmail("Build Successful", fmt.Sprintf("Build completed successfully for branch %s", branch))
			}
		}()
		fmt.Fprintf(w, "Build triggered for branch: %s", branch)
	} else {
		fmt.Fprintf(w, "Ignoring push to branch: %s", branch)
	}
}

func buildProject() error {
	log.Println("Pulling latest changes...")
	if err := runCommand("git", "pull"); err != nil {
		return fmt.Errorf("git pull failed: %w", err)
	}

	log.Println("Building project...")
	if err := runCommand("flutter", "build", "apk"); err != nil {
		return fmt.Errorf("flutter build failed: %w", err)
	}

	log.Println("Build complete")
	return nil
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %s\nOutput: %s", err, output)
	}
	log.Printf("Command output: %s", output)
	return nil
}

func sendEmail(subject, body string) {
	domain := os.Getenv("MAILGUN_DOMAIN")
	apiKey := os.Getenv("MAILGUN_API_KEY")
	sender := os.Getenv("EMAIL_SENDER")
	recipients := strings.Split(os.Getenv("EMAIL_RECIPIENTS"), ",")

	mg := mailgun.NewMailgun(domain, apiKey)

	message := mg.NewMessage(sender, subject, body, recipients...)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_, id, err := mg.Send(ctx, message)
	if err != nil {
		log.Printf("Failed to send email: %v", err)
	} else {
		log.Printf("Email sent successfully. ID: %s", id)
	}
}
