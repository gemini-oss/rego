/*
# Slack

This package initializes the methods for functions which handle a typical Slack bot with the Slack Web API:
https://api.slack.com/web

:Copyright: (c) 2023 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/slack/bot.go
package slack

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func (c *Client) EventHandler(w http.ResponseWriter, r *http.Request) {
	c.Log.Printf("Received a request")

	// Parse the request body
	body, err := io.ReadAll(r.Body)
	c.Log.Println(string(body))
	if err != nil {
		c.Log.Printf("Error reading request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Verify the request
	if !c.VerifyRequest(r, body) {
		c.Log.Printf("Failed to verify request")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	c.Log.Printf("Request verified!")

	// Handle the request based on its type
	// Here, you can add the rest of your Slack-specific handling logic
	// Handle the request based on its type
	cb := &EventCallback{}
	err = json.Unmarshal(body, &cb)
	if err != nil {
		c.Log.Printf("Error unmarshalling request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch cb.Type {
	case "url_verification":
		// Return challenge for URL verification during app installation
		challenge := &SlackChallenge{}
		err := json.Unmarshal(body, &challenge)
		if err != nil {
			c.Log.Printf("Error decoding challenge: %v", err)
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		tmpl := template.Must(template.New("challenge").Parse(`{{.}}`))
		w.Header().Set("Content-Type", "text/plain")
		err = tmpl.Execute(w, challenge.Challenge)
		if err != nil {
			c.Log.Printf("Error executing template: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}

	case "event_callback":
		// Handle event
		c.Log.Printf("Received event: %+v", cb.Event)
		c.Log.Printf(cb.Event.Text)

		c.SendReply(&cb.Event, nil)

	default:
		c.Log.Printf("Unknown request type: %s", cb.Type)
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (c *Client) CommandHandler(w http.ResponseWriter, r *http.Request) {
	c.Log.Printf("Received a request")

	// Parse the request body
	body, err := io.ReadAll(r.Body)
	c.Log.Println(string(body))
	if err != nil {
		c.Log.Printf("Error reading request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Verify the request
	if !c.VerifyRequest(r, body) {
		c.Log.Printf("Failed to verify request")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	c.Log.Printf("Request verified!")

	// Handle the request based on its type
	// Here, you can add the rest of your Slack-specific handling logic
	// Handle the request based on its type
	v, err := url.ParseQuery(string(body))
	if err != nil {
		panic(err)
	}
	sc := &SlashCommand{
		Token:       v.Get("token"),
		TeamID:      v.Get("team_id"),
		ChannelID:   v.Get("channel_id"),
		UserName:    v.Get("user_name"),
		Command:     v.Get("command"),
		Text:        v.Get("text"),
		ResponseURL: v.Get("response_url"),
	}
	c.Log.Println(sc)

	m := &SlackMessage{Channel: sc.ChannelID, Token: c.Token}
	userlist, _ := c.ListUsers()
	for _, user := range userlist.Members {
		m.Text += fmt.Sprintf(":boom: `%s`\n", user.Name)
		ch, _ := c.GetUserChannels(user.ID)
		m.Text += "Channels: \n"
		for _, channel := range ch.Channels {
			m.Text += fmt.Sprintf(":bone: - `%s`\n", channel.Name)
		}
	}
	c.SendMessage(nil, m)
}

func (c *Client) GetBotID() (string, error) {
	url := c.BuildURL("%s/auth.test")

	var p struct {
		Token string
	}
	p.Token = c.Token

	res, body, err := c.HTTP.DoRequest("POST", url, nil, p)
	if err != nil {
		return "", err
	}
	c.Log.Println("Response Status:", res.Status)
	c.Log.Debug("Response Body:", string(body))

	var result struct {
		UserID string `json:"user_id"`
	}

	err = json.Unmarshal(body, &body)
	if err != nil {
		return "", fmt.Errorf("unmarshalling user: %w", err)
	}

	c.Log.Printf("Bot ID is: %s", result.UserID)
	c.BotID = result.UserID
	return result.UserID, nil
}

func (c *Client) SendMessage(e *Event, m *SlackMessage) error {
	url := c.BuildURL("%s/chat.postMessage")

	message := SlackMessage{
		Channel:  m.Channel,
		Text:     m.Text,
		Markdown: true,
	}

	res, body, err := c.HTTP.DoRequest("POST", url, nil, message)
	if err != nil {
		return err
	}
	c.Log.Println("Response Status:", res.Status)
	c.Log.Debug("Response Body:", string(body))

	err = json.Unmarshal(body, &m)
	if err != nil {
		return fmt.Errorf("unmarshalling user: %w", err)
	}

	return nil
}

func (c *Client) SendReply(e *Event, m *SlackMessage) error {
	url := c.BuildURL("%s/chat.postMessage")

	reply := SlackMessage{
		Channel:  e.Channel,
		Text:     e.Text,
		ThreadTS: e.EventTS,
	}

	res, body, err := c.HTTP.DoRequest("POST", url, nil, reply)
	if err != nil {
		return err
	}
	c.Log.Println("Response Status:", res.Status)
	c.Log.Debug("Response Body:", string(body))

	err = json.Unmarshal(body, &m)
	if err != nil {
		return fmt.Errorf("unmarshalling user: %w", err)
	}

	return nil
}

func (c *Client) VerifyRequest(r *http.Request, body []byte) bool {
	c.Log.Println("Verifying Slack Request")

	// Collect headers and signing secret
	timestamp := r.Header.Get("X-Slack-Request-Timestamp")
	slackSignature := r.Header.Get("X-Slack-Signature")

	// Check the timestamp to prevent replay attacks
	t, _ := strconv.ParseInt(timestamp, 10, 64)
	if time.Now().Unix()-t > 60*5 {
		return false
	}

	// Compute the HMAC
	baseString := "v0:" + timestamp + ":" + string(body)
	h := hmac.New(sha256.New, []byte(c.SigningSecret))
	h.Write([]byte(baseString))

	// The computed signature
	mySignature := "v0=" + hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(mySignature), []byte(slackSignature))
}
