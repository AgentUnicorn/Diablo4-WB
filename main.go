package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	helper "github.com/AgentUnicorn/go-Diablo4-WB/utils"
)

// A Response struct to map the Entire Response
type WorldBoss struct {
	Name    string `json:"name"`
	Minutes int    `json:"time"`
}

var (
	lastSentTime   time.Time
	messageSent    bool
	resetThreshold = 30 * time.Minute // Set the reset threshold to 30 minutes
)

func main() {
	go startAPIServer()
	getWBPeriodically()
}

func startAPIServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Fetch API data
		var worldBoss WorldBoss
		getWolrdBoss(&worldBoss)

		if worldBoss.Minutes <= 30 {
			sendWBTimeToDiscord(worldBoss)
		}

	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getWolrdBoss(worldBoss *WorldBoss) {
	response, err := http.Get("https://api.worldstone.io/world-bosses")
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	json.NewDecoder(response.Body).Decode(&worldBoss)

}

func getWBPeriodically() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		var worldBoss WorldBoss
		getWolrdBoss(&worldBoss)

		fmt.Println("Api respone")
		fmt.Println("Boss name: ", worldBoss.Name)
		fmt.Println("Minutes left: ", worldBoss.Minutes)

		if worldBoss.Minutes >= 25 && worldBoss.Minutes <= 30 && !messageSent {
			now := time.Now()
			if now.Sub(lastSentTime) > resetThreshold {
				sendWBTimeToDiscord(worldBoss)

				// Update the last sent time and set the messageSent flag to true
				lastSentTime = now
				messageSent = true
			}
		} else {
			messageSent = false
		}
		time.Sleep(5 * time.Minute) // Fetch every 5 minutes
		// time.Sleep(10 * time.Second) // Fetch every 10 seconds
	}
}

func sendWBTimeToDiscord(worldBoss WorldBoss) {
	bossNameSnakeCase := helper.ConvertToSnakeCase(worldBoss.Name)
	now := time.Now().Unix()
	spawnTime := now + int64(worldBoss.Minutes)*60
	spawnTimeUTC7, err := helper.ParseTimestampToUTC7(int(spawnTime))
	if err != nil {
		fmt.Println("Error parsing timestamp:", err)
		return
	}

	lukingfishEmote := "<:lukingfish:980749518117142588>"
	hulishitEmote := "<:hulishit:1005313867032834098>"

	ROLE_ID := "1134408189308321822"
	mentionMessage := "<@&" + ROLE_ID + ">"

	webhook := "https://discord.com/api/webhooks/1134303176271609917/iz7hCS9q0FvO8lH1mJm06LvdEq2Zatu9a7nP6D63cOnYWiV7fTKRay5UI3d9TWb38wUq"
	message := mentionMessage + "\n# " + lukingfishEmote + " | Incoming world boss: " + worldBoss.Name + "\n\n\nStart time: " + spawnTimeUTC7.Format("02/01/2006 15:04") + "\n\n\nGood farming " + hulishitEmote
	imageURL := "assets/world_bosses/" + bossNameSnakeCase + ".jpeg"

	sendDiscordMessageWithImage(webhook, message, imageURL)
	if err != nil {
		fmt.Println("Error sending Discord message:", err)
		return
	}

	fmt.Println("Message with image sent successfully to Discord.")
}

func sendDiscordMessageWithImage(url string, message string, imageURL string) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	messagePart, err := writer.CreateFormField("content")
	if err != nil {
		return err
	}
	_, err = io.WriteString(messagePart, message)
	if err != nil {
		return err
	}

	// Add the image part
	imagePart, err := writer.CreateFormFile("file", imageURL)
	if err != nil {
		return err
	}
	imageFile, err := os.Open(imageURL)
	if err != nil {
		return err
	}
	defer imageFile.Close()
	_, err = io.Copy(imagePart, imageFile)
	if err != nil {
		return err
	}

	// Close the multipart writer
	writer.Close()

	// Create the POST request with the multipart body
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("Failed to send Discord message. Status: %d", resp.StatusCode)
	}

	return nil
}
