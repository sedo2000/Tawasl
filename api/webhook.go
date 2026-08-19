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
)


// --- Telegram Bot API Structs ---

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

// --- Helper Functions ---

func getAdmins() []int64 {
	var admins []int64
	if id1, err := strconv.ParseInt(os.Getenv("ADMIN_ID_1"), 10, 64); err == nil && id1 != 0 {
		admins = append(admins, id1)
	}
	if id2, err := strconv.ParseInt(os.Getenv("ADMIN_ID_2"), 10, 64); err == nil && id2 != 0 {
		admins = append(admins, id2)
	}
	return admins
}

func isAdmin(userID int64) bool {
	for _, adminID := range getAdmins() {
		if userID == adminID {
			return true
		}
	}
	return false
}

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

func editMessageText(chatID int64, messageID int, text string, keyboard *InlineKeyboardMarkup) {
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"message_id": messageID,
		"text":       text,
		"parse_mode": "HTML",
	}
	if keyboard != nil {
		payload["reply_markup"] = keyboard
	}
	_, _ = sendTelegramRequest("editMessageText", payload)
}

func copyMessage(toChatID, fromChatID int64, messageID int) {
	payload := map[string]interface{}{
		"chat_id":      toChatID,
		"from_chat_id": fromChatID,
		"message_id":   messageID,
	}
	_, _ = sendTelegramRequest("copyMessage", payload)
}

func answerCallbackQuery(callbackQueryID, text string) {
	payload := map[string]interface{}{
		"callback_query_id": callbackQueryID,
		"text":              text,
		"show_alert":        false,
	}
	_, _ = sendTelegramRequest("answerCallbackQuery", payload)
}

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

func getAdminHeader() string {
	return `<b>📊 إحصائيات اليوم:</b>
👥 <b>الاجمالي :</b> 1,250
🆕 <b>مستخدمون :</b> 34`
}

// --- Keyboards Layouts ---

func buildMainPanelKeyboard() *InlineKeyboardMarkup {
	return &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{
				{Text: "⚙️ الإعدادات", CallbackData: "nav_settings", Style: "primary"},
				{Text: "📝 المحتوى", CallbackData: "nav_content", Style: "primary"},
			},
			{
				{Text: "👥 المستخدمون", CallbackData: "nav_users", Style: "primary"},
				{Text: "🔐 الاشتراك", CallbackData: "nav_sub", Style: "primary"},
			},
			{
				{Text: "📢 التواصل", CallbackData: "dummy_contact", Style: "primary"},
				{Text: "💰 المالية", CallbackData: "dummy_finance", Style: "primary"},
			},
			{
				{Text: "🛠️ النظام والدعم", CallbackData: "nav_system", Style: "success"},
			},
			{
				{Text: "❌ إشعار الحظر 🚫", CallbackData: "toggle_ban_notif", Style: "danger"},
				{Text: "❌ إشعار الدخول 🔔", CallbackData: "toggle_login_notif", Style: "danger"},
			},
			{
				{Text: "❓ دليل الاستخدام", CallbackData: "dummy_guide", Style: "primary"},
			},
			{
				{Text: "• لوحه تحكم في بوت السايت •", CallbackData: "dummy_info", Style: "primary"},
			},
		},
	}
}

func buildContentKeyboard() *InlineKeyboardMarkup {
	return &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{{Text: "👋 رسالة الترحيب", CallbackData: "cnt_welcome", Style: "primary"}},
			{{Text: "💬 الردود التلقائية", CallbackData: "cnt_auto_reply", Style: "primary"}},
			{
				{Text: "⚪ الأزرار الشفافة", CallbackData: "cnt_transparent", Style: "primary"},
				{Text: "✏️ تعديل الأزرار", CallbackData: "cnt_edit_btn", Style: "primary"},
			},
			{{Text: "📎 الاختصارات", CallbackData: "cnt_shortcuts", Style: "primary"}},
			{
				{Text: "✏️ تعديل المحتوى", CallbackData: "cnt_edit_content", Style: "primary"},
				{Text: "📋 قائمة التعديلات", CallbackData: "cnt_edit_list", Style: "primary"},
			},
			{{Text: "🔗 ديب لينك مخصص (0)", CallbackData: "cnt_deeplink", Style: "primary"}},
			{{Text: "🌐 الترجمة", CallbackData: "cnt_translate", Style: "primary"}},
			{{Text: "ℹ️ معلومات البوت", CallbackData: "cnt_bot_info", Style: "primary"}},
			{{Text: "❓ المساعدة", CallbackData: "cnt_help", Style: "primary"}},
			{{Text: "• رجوع •", CallbackData: "nav_main", Style: "danger"}},
		},
	}
}

func buildSubscriptionKeyboard() *InlineKeyboardMarkup {
	return &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{{Text: "➕ إضافة اشتراك جديد (0/10)", CallbackData: "sub_add", Style: "success"}},
			{{Text: "—— الإعدادات ——", CallbackData: "dummy_header1", Style: "primary"}},
			{{Text: "❌ الإشعار", CallbackData: "sub_notif", Style: "danger"}},
			{{Text: "—— العرض ——", CallbackData: "dummy_header2", Style: "primary"}},
			{
				{Text: "⚪ زر التحقق", CallbackData: "sub_check_btn", Style: "primary"},
				{Text: "📋 العرض: مجمّعة", CallbackData: "sub_display_type", Style: "primary"},
			},
			{
				{Text: "🎨 النمط ❌", CallbackData: "sub_style", Style: "danger"},
				{Text: "❌ أيقونة زر التحقق 😀", CallbackData: "sub_icon", Style: "danger"},
			},
			{{Text: "✏️ رسالة العرض المُجمَّع", CallbackData: "sub_group_msg", Style: "primary"}},
			{
				{Text: "• رجوع •", CallbackData: "nav_main", Style: "danger"},
				{Text: "❓ شرح القسم", CallbackData: "sub_explain", Style: "primary"},
			},
		},
	}
}

func buildSettingsKeyboard() *InlineKeyboardMarkup {
	return &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{{Text: "🤖 عمل البوت", CallbackData: "stg_work", Style: "primary"}},
			{{Text: "🔐 قسم التحقق من العضوية", CallbackData: "stg_membership", Style: "primary"}},
			{{Text: "🔒 حماية المحتوى", CallbackData: "stg_protection", Style: "primary"}},
			{{Text: "🔔 الإشعارات", CallbackData: "stg_notifs", Style: "primary"}},
			{{Text: "🔴 الحذف التلقائي ⏱️", CallbackData: "stg_autodelete", Style: "danger"}},
			{{Text: "🔴 تذكير غير النشطين 🔔", CallbackData: "stg_remind", Style: "danger"}},
			{{Text: "📎 ردود سريعة (0)", CallbackData: "stg_quick_replies", Style: "primary"}},
			{{Text: "• رجوع •", CallbackData: "nav_main", Style: "danger"}},
		},
	}
}

func buildUsersKeyboard() *InlineKeyboardMarkup {
	return &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{
				{Text: "📊 الإحصائيات", CallbackData: "usr_stats", Style: "primary"},
				{Text: "👤 المسؤولون", CallbackData: "usr_admins", Style: "primary"},
			},
			{{Text: "🚫 إدارة الحظر", CallbackData: "usr_bans", Style: "danger"}},
			{{Text: "📋 سجل النشاط", CallbackData: "usr_logs", Style: "primary"}},
			{{Text: "• رجوع •", CallbackData: "nav_main", Style: "danger"}},
		},
	}
}

func buildSystemKeyboard() *InlineKeyboardMarkup {
	return &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{{Text: "🛠️ النظام والدعم", CallbackData: "sys_status", Style: "success"}},
			{
				{Text: "❌ إشعار الدخول 🔔", CallbackData: "toggle_login_notif", Style: "danger"},
				{Text: "❌ إشعار الحظر 🚫", CallbackData: "toggle_ban_notif", Style: "danger"},
			},
			{{Text: "❓ دليل الاستخدام", CallbackData: "dummy_guide", Style: "primary"}},
			{{Text: "• رجوع •", CallbackData: "nav_main", Style: "danger"}},
		},
	}
}

// --- Main Handler ---

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Sayat Bot Webhook Running!"))
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

	// 1. معالجة ضغطات الأزرار (Callback Queries)
	if update.CallbackQuery != nil {
		cb := update.CallbackQuery
		if isAdmin(cb.From.ID) {
			handleAdminNavigation(cb)
		} else {
			answerCallbackQuery(cb.ID, "هذه اللوحة مخصصة للمشرفين فقط.")
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	// 2. معالجة الرسائل الواردة
	if update.Message != nil {
		msg := update.Message
		userID := msg.From.ID

		// أمر التشغيل /start
		if msg.Text == "/start" {
			if isAdmin(userID) {
				sendMessage(userID, getAdminHeader(), buildMainPanelKeyboard())
			} else {
				sendMessage(userID, "🔒 <b>مرحباً بك في بوت الصراحة والسايت!</b>\nأرسل رسالتك أو ملاحظتك بصراحة وبدون كشف هويتك، وسأوصلها مباشرة إلى الإدارة.", nil)
			}
			w.WriteHeader(http.StatusOK)
			return
		}

		// التعامل مع المشرفين
		if isAdmin(userID) {
			if msg.Text == "/admin" {
				sendMessage(userID, getAdminHeader(), buildMainPanelKeyboard())
				w.WriteHeader(http.StatusOK)
				return
			}

			// عند رد المشرف (Reply) على رسالة العميل
			if msg.ReplyTo != nil {
				targetUserID := extractUserIDFromReply(msg.ReplyTo)
				if targetUserID != 0 {
					copyMessage(targetUserID, userID, msg.MessageID)
					sendMessage(userID, "✅ تم إرسال ردك إلى العميل بنجاح.", nil)
				} else {
					sendMessage(userID, "⚠️ تعذر استخراج آيدي العميل. اختر Reply على الهيدر الذي يحتوي على <code>🆔 ID:</code>.", nil)
				}
				w.WriteHeader(http.StatusOK)
				return
			}
		}

		// التعامل مع رسائل العملاء -> توجيهها لكلا المشرفين 1 و 2
		if !isAdmin(userID) {
			header := fmt.Sprintf("📩 <b>رسالة جديدة من صراحة/سايت:</b>\n👤 <b>الاسم:</b> %s\n🆔 ID: <code>%d</code>", msg.From.FirstName, userID)
			if msg.From.Username != "" {
				header += fmt.Sprintf("\n🔗 <b>المعرف:</b> @%s", msg.From.Username)
			}

			// إرسال التنبيه والرسالة لكل مشرف موجود بالقائمة
			for _, adminID := range getAdmins() {
				sendMessage(adminID, header, nil)
				copyMessage(adminID, msg.Chat.ID, msg.MessageID)
			}

			sendMessage(userID, "✅ تم إرسال رسالتك السرية بنجاح إلى الإدارة.", nil)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func handleAdminNavigation(cb *CallbackQuery) {
	chatID := cb.Message.Chat.ID
	msgID := cb.Message.MessageID

	switch cb.Data {
	case "nav_main":
		editMessageText(chatID, msgID, getAdminHeader(), buildMainPanelKeyboard())
	case "nav_content":
		editMessageText(chatID, msgID, "<b>📝 قسم إدارة المحتوى:</b>", buildContentKeyboard())
	case "nav_sub":
		editMessageText(chatID, msgID, "<b>🔐 قسم إدارة الاشتراك الإجباري:</b>", buildSubscriptionKeyboard())
	case "nav_settings":
		editMessageText(chatID, msgID, "<b>⚙️ قسم إعدادات البوت:</b>", buildSettingsKeyboard())
	case "nav_users":
		editMessageText(chatID, msgID, "<b>👥 قسم إدارة المستخدمين والمسؤولين:</b>", buildUsersKeyboard())
	case "nav_system":
		editMessageText(chatID, msgID, "<b>🛠️ قسم النظام والدعم الفني:</b>", buildSystemKeyboard())
	default:
		answerCallbackQuery(cb.ID, "تم الضغط على الزر.")
	}
}
