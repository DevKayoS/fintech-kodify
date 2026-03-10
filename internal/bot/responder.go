package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/DevKayoS/fintech-kodify/internal/models"
)

// sendMessage envia uma mensagem de texto para um chat do Telegram.
func sendMessage(chatID int64, text string) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		slog.Error("telegram: TELEGRAM_BOT_TOKEN não definido")
		return
	}

	payload := models.TelegramSendMessage{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "Markdown",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		slog.Error("telegram: falha ao serializar mensagem", "error", err)
		return
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		slog.Error("telegram: falha ao enviar mensagem", "chat_id", chatID, "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Warn("telegram: resposta inesperada ao enviar mensagem", "chat_id", chatID, "status", resp.StatusCode)
	}
}
