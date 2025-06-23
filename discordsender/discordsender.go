package discordsender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Message represents one parsed megaphone message
type Message struct {
	Nickname    string
	ProfileCode string
	Text        string
}

// Sender 用來接收訊息並轉發到 Discord
type Sender struct {
	WebhookURL string
	InputChan  chan Message
	quit       chan struct{}
}

// NewSender 建立 Sender 並啟動 goroutine
func NewSender(webhookURL string) *Sender {
	s := &Sender{
		WebhookURL: webhookURL,
		InputChan:  make(chan Message, 100),
		quit:       make(chan struct{}),
	}
	go s.loop()
	return s
}

// Close 關閉 channel 並結束 goroutine
func (s *Sender) Close() {
	close(s.quit)
	close(s.InputChan)
}

// loop 持續監聽 channel 並傳送到 Discord
func (s *Sender) loop() {
	for {
		select {
		case msg, ok := <-s.InputChan:
			if !ok {
				return
			}
			s.sendToDiscord(msg)
		case <-s.quit:
			return
		}
	}
}

// sendToDiscord 格式化並送出訊息
func (s *Sender) sendToDiscord(msg Message) {
	content := fmt.Sprintf("%s#%s: %s", msg.Nickname, msg.ProfileCode, msg.Text)
	body, _ := json.Marshal(map[string]string{
		"content": content,
	})

	resp, err := http.Post(s.WebhookURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("send error:", err)
		return
	}
	defer resp.Body.Close()
}
