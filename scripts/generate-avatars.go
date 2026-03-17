package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// OpenAI API structures
type ImageRequest struct {
	Prompt string `json:"prompt"`
	N      int    `json:"n"`
	Size   string `json:"size"`
	Format string `json:"response_format"`
}

type ImageResponse struct {
	Data []struct {
		B64Json string `json:"b64_json"`
	} `json:"data"`
}

// Agent represents a sample agent for avatar generation
type Agent struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Class       string `json:"class"`
	Description string `json:"description"`
}

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable not set")
	}

	// Sample agents to generate avatars for
	agents := []Agent{
		{
			ID:          "code-wizard",
			Name:        "CodeWizard",
			Class:       "Developer",
			Description: "mystical robed figure with glowing coding symbols, wizard hat with binary patterns",
		},
		{
			ID:          "bug-hunter",
			Name:        "BugHunter",
			Class:       "Debugger",
			Description: "armored warrior with magnifying glass shield, hunting gear, determined expression",
		},
		{
			ID:          "test-guardian",
			Name:        "TestGuardian",
			Class:       "Tester",
			Description: "knight in shining armor with shield and sword, protective stance, quality assurance emblem",
		},
		{
			ID:          "doc-scribe",
			Name:        "DocScribe",
			Class:       "Writer",
			Description: "scholarly monk with quill and scrolls, wise expression, ancient tome in hand",
		},
		{
			ID:          "deploy-master",
			Name:        "DeployMaster",
			Class:       "DevOps",
			Description: "steampunk engineer with gears and tools, confident pose, mechanical contraptions",
		},
	}

	outputDir := "../frontend/public/avatars"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	for _, agent := range agents {
		log.Printf("Generating avatar for %s (%s)...", agent.Name, agent.Class)
		
		prompt := fmt.Sprintf("64x64 pixel art portrait of %s, a %s character. %s. Dark fantasy RPG style, transparent background, retro 16-bit game aesthetic, detailed but simple, centered composition",
			agent.Name, agent.Class, agent.Description)

		imageData, err := generateImage(apiKey, prompt)
		if err != nil {
			log.Printf("Failed to generate image for %s: %v", agent.Name, err)
			continue
		}

		filename := filepath.Join(outputDir, agent.ID+".png")
		if err := saveImage(imageData, filename); err != nil {
			log.Printf("Failed to save image for %s: %v", agent.Name, err)
			continue
		}

		log.Printf("✓ Saved avatar for %s as %s", agent.Name, filename)
	}

	log.Println("Avatar generation complete!")
}

func generateImage(apiKey, prompt string) ([]byte, error) {
	reqBody := ImageRequest{
		Prompt: prompt,
		N:      1,
		Size:   "512x512", // OpenAI doesn't support 64x64, we'll resize later
		Format: "b64_json",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/images/generations", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var imageResp ImageResponse
	if err := json.NewDecoder(resp.Body).Decode(&imageResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if len(imageResp.Data) == 0 {
		return nil, fmt.Errorf("no images returned")
	}

	imageData, err := base64.StdEncoding.DecodeString(imageResp.Data[0].B64Json)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 image: %v", err)
	}

	return imageData, nil
}

func saveImage(data []byte, filename string) error {
	return os.WriteFile(filename, data, 0644)
}