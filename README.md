# 📘 Discord Webhook 設定教學

本文件說明如何在 Discord 建立 Webhook，並將其用於本專案中的設定檔 `conf.toml`。

---

## 🧩 一、開啟開發者模式

若你需要取得伺服器或頻道的 ID（例如用於調試），請先開啟 **開發者模式**：

1. 打開 Discord 用戶端（桌機或網頁版皆可）。
2. 點擊左下角的 ⚙️ **使用者設定 (User Settings)**。
3. 在左側選單中選擇 **進階 (Advanced)**。
4. 開啟 **開發者模式 (Developer Mode)**。

> ✅ 開啟後，你可以右鍵點擊伺服器、頻道或使用者 →「複製 ID」。

---

## 🪄 二、建立 Discord Webhook

1. 進入你想接收通知的伺服器。  
2. 前往該伺服器的某個 **文字頻道**（例如 `#general`）。
3. 點擊頻道名稱右側的 ⚙️ **編輯頻道 (Edit Channel)**。
4. 在左側選單選擇 **整合 (Integrations)**。
5. 點擊 **建立 Webhook (Create Webhook)**。
6. 為 Webhook 命名（可自訂，如「Go Bot」）。
7. （可選）設定頭像。
8. 複製產生的 **Webhook URL**。

> 範例：
> ```https://discord.com/api/webhooks/1234567890123456789/AbCdEfGhIjKlMnOpQrStUvWxYz```


---

## 🧾 三、設定 conf.toml

將剛才複製的 Webhook URL 貼入 `conf.toml` 檔案中：

```toml
web_hook = "https://discord.com/api/webhooks/1234567890123456789/AbCdEfGhIjKlMnOpQrStUvWxYz"
```


## 四、啟動程序

- 右鍵點擊app.exe並選擇系統管理員身分執行

## 重新打包執行檔案:

1. 確認安裝git
2. 開啟git bash應用程序, 並到專案資料夾
3. 執行以下
```
GOOS=windows GOARCH=amd64 go build -o app.exe main.go
```

