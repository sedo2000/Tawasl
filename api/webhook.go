package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// --- Telegram Bot API 9.4 Structs ---

type InlineKeyboardButton struct {
	Text              string `json:"text"`
	CallbackData      string `json:"callback_data,omitempty"`
	URL               string `json:"url,omitempty"`
	Style             string `json:"style,omitempty"` // "primary", "success", "danger"
	IconCustomEmojiID string `json:"icon_custom_emoji_id,omitempty"`
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type User struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	Username  string `json:"username,omitempty"`
}

type Chat struct {
	ID int64 `json:"id"`
}

type Message struct {
	MessageID int                   `json:"message_id"`
	From      *User                 `json:"from"`
	Chat      Chat                  `json:"chat"`
	Text      string                `json:"text,omitempty"`
	Caption   string                `json:"caption,omitempty"`
	ReplyTo   *Message              `json:"reply_to_message,omitempty"`
	Markup    *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

type CallbackQuery struct {
	ID      string   `json:"id"`
	From    *User    `json:"from"`
	Message *Message `json:"message"`
	Data    string   `json:"data"`
}

type Update struct {
	UpdateID      int            `json:"update_id"`
	Message       *Message       `json:"message,omitempty"`
	CallbackQuery *CallbackQuery `json:"callback_query,omitempty"`
}

// --- Telegram API Helpers ---

func sendTelegramRequest(method string, payload interface{}) ([]byte, error) {
	token := os.Getenv("BOT_TOKEN")
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/%s", token, method)

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func sendMessage(chatID int64, text string, keyboard *InlineKeyboardMarkup) {
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "HTML",
	}
	if keyboard != nil {
		payload["reply_markup"] = keyboard
	}
	_, _ = sendTelegramRequest("sendMessage", payload)
}

func copyMessage(toChatID, fromChatID int64, messageID int, keyboard *InlineKeyboardMarkup) {
	payload := map[string]interface{}{
		"chat_id":      toChatID,
		"from_chat_id": fromChatID,
		"message_id":   messageID,
	}
	if keyboard != nil {
		payload["reply_markup"] = keyboard
	}
	_, _ = sendTelegramRequest("copyMessage", payload)
}

func answerCallbackQuery(callbackQueryID, text string) {
	payload := map[string]interface{}{
		"callback_query_id": callbackQueryID,
		"text":              text,
		"show_alert":        true,
	}
	_, _ = sendTelegramRequest("answerCallbackQuery", payload)
}

// استخراج آيدي العميل بشكل صحيح
func extractUserIDFromReply(msg *Message) int64 {
	if msg == nil {
		return 0
	}
	text := msg.Text
	if text == "" {
		text = msg.Caption
	}

	re := regexp.MustCompile(`(?:🆔\s*ID:\s*|ID:\s*)(\d+)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		id, err := strconv.ParseInt(matches[1], 10, 64)
		if err == nil {
			return id
		}
	}
	return 0
}

// --- تصميم الأزرار الملونة ---

func buildAdminPanelKeyboard() *InlineKeyboardMarkup {
	return &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{
				{Text: "🟢 الحالة: مفعّل", CallbackData: "status_active", Style: "success"},
				{Text: "ℹ️ تعليمات الاستخدام", CallbackData: "admin_help", Style: "primary"},
			},
		},
	}
}

func buildUserActionKeyboard(userID int64) *InlineKeyboardMarkup {
	return &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{
				{Text: "💬 كيفية الرد", CallbackData: fmt.Sprintf("how_reply_%d", userID), Style: "success"},
				{Text: "👤 آيدي العميل", CallbackData: fmt.Sprintf("info_%d", userID), Style: "primary"},
			},
		},
	}
}

// --- Main Handler ---

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Tawasl Bot is Running!"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var update Update
	if err := json.Unmarshal(body, &update); err != nil {
		http.Error(w, "JSON Decode Error", http.StatusBadRequest)
		return
	}

	adminID, _ := strconv.ParseInt(os.Getenv("ADMIN_ID"), 10, 64)

	if update.CallbackQuery != nil {
		cb := update.CallbackQuery
		if cb.From.ID == adminID {
			handleAdminCallbacks(cb, adminID)
		} else {
			answerCallbackQuery(cb.ID, "هذه الأزرار مخصصة لإدارة البوت فقط.")
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	if update.Message != nil {
		msg := update.Message
		userID := msg.From.ID

		if msg.Text == "/start" {
			if userID == adminID {
				sendMessage(adminID, "<b>لوحة تحكم المطور 🛠️</b>", buildAdminPanelKeyboard())
			} else {
				sendMessage(userID, "مرحباً بك! أرسل رسالتك وسيقوم الدعم الفني بالرد عليك في أقرب وقت.", nil)
			}
			w.WriteHeader(http.StatusOK)
			return
		}

		if userID == adminID {
			if msg.Text == "/admin" {
				sendMessage(adminID, "<b>لوحة تحكم المطور:</b>", buildAdminPanelKeyboard())
				w.WriteHeader(http.StatusOK)
				return
			}

			if msg.ReplyTo != nil {
				targetUserID := extractUserIDFromReply(msg.ReplyTo)
				if targetUserID != 0 {
					copyMessage(targetUserID, adminID, msg.MessageID, nil)
					sendMessage(adminID, "✅ تم تحويل الرد بنجاح إلى العميل.", nil)
				} else {
					sendMessage(adminID, "⚠️ تعذر التعرف على آيدي العميل. يرجى عمل Reply على الهيدر الذي يحتوي على 🆔 ID:.", nil)
				}
				w.WriteHeader(http.StatusOK)
				return
			}
		}

		if userID != adminID {
			header := fmt.Sprintf("📩 <b>رسالة جديدة من العميل:</b>\n👤 <b>الاسم:</b> %s\n🆔 ID: <code>%d</code>", msg.From.FirstName, userID)
			if msg.From.Username != "" {
				header += fmt.Sprintf("\n🔗 <b>المعرف:</b> @%s", msg.From.Username)
			}

			sendMessage(adminID, header, buildUserActionKeyboard(userID))
			copyMessage(adminID, msg.Chat.ID, msg.MessageID, nil)

			sendMessage(userID, "✅ تم استلام رسالتك بنجاح، سيتم الرد عليك قريباً.", nil)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func handleAdminCallbacks(cb *CallbackQuery, adminID int64) {
	data := cb.Data

	switch {
	case data == "status_active":
		answerCallbackQuery(cb.ID, "🟢 البوت يعمل بكفاءة على Vercel.")

	case data == "admin_help":
		answerCallbackQuery(cb.ID, "قم بعمل Reply على رسالة التنبيه التي تحتوي على آيدي العميل.")

	case strings.HasPrefix(data, "how_reply_"):
		answerCallbackQuery(cb.ID, "اضغط رد (Reply) على رسالة الهيدر واكتب ردك مباشرة.")

	case strings.HasPrefix(data, "info_"):
		targetID := strings.TrimPrefix(data, "info_")
		answerCallbackQuery(cb.ID, fmt.Sprintf("آيدي العميل: %s", targetID))
	}
}
