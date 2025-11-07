package discordsender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Message represents one parsed megaphone message
type Message struct {
	Nickname    string
	ProfileCode string
	Text        string
	Channel     string
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

func (s *Sender) sendToDiscord(msg Message) {
	// Discord Markdown 格式：
	// 1️⃣ 第一行：粗體顯示使用者與代碼
	// 2️⃣ 第二行：灰色小字顯示頻道
	// 3️⃣ 第三行：訊息正文

	content := fmt.Sprintf("**%s#%s**\n> 頻道: %s\n> %s\n",
		msg.Nickname,
		msg.ProfileCode,
		msg.Channel,
		msg.Text,
	)

	if strings.Contains(content, "組隊") ||
		strings.Contains(content, "组队") ||
		strings.Contains(content, "訓練") ||
		strings.Contains(content, "弓箭手村") ||
		strings.Contains(content, "都懂") {
		return
	}

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
