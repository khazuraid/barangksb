package handler

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SettingHandler struct {
	pool *pgxpool.Pool
}

func NewSettingHandler(pool *pgxpool.Pool) *SettingHandler {
	return &SettingHandler{pool: pool}
}

func newTelegramBot(token, apiURL string) (*bot.Bot, error) {
	if apiURL == "" {
		apiURL = os.Getenv("TELEGRAM_API_URL")
	}
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	opts := []bot.Option{
		bot.WithSkipGetMe(),
		bot.WithHTTPClient(20*time.Second, httpClient),
	}
	apiURL = strings.TrimRight(strings.TrimSpace(apiURL), "/")
	if apiURL != "" {
		opts = append(opts, bot.WithServerURL(apiURL))
	}
	return bot.New(token, opts...)
}

// ---- types ----

type subscriber struct {
	ChatID       int64  `json:"chat_id"`
	Type         string `json:"type"`
	Title        string `json:"title"`
	Username     string `json:"username"`
	NotifyIn     bool   `json:"notify_in"`
	NotifyOut    bool   `json:"notify_out"`
	NotifyAdjust bool   `json:"notify_adjust"`
	NotifyLow    bool   `json:"notify_low_stock"`
	Active       bool   `json:"active"`
}

// ---- bot info ----

// GET /api/settings/telegram/bot — returns bot identity via getMe
func (h *SettingHandler) GetBotInfo(c *gin.Context) {
	token := h.getSetting(c, "telegram_bot_token")
	if token == "" {
		c.JSON(http.StatusOK, gin.H{"configured": false})
		return
	}
	apiURL := h.getSetting(c, "telegram_api_url")

	botInst, err := newTelegramBot(token, apiURL)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"configured": true, "error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	me, err := botInst.GetMe(ctx)
	if err != nil {
		// Fallback to cached bot info if live call fails (due to timeout or restricted network)
		cachedID := h.getSetting(c, "telegram_bot_id")
		cachedUser := h.getSetting(c, "telegram_bot_username")
		cachedName := h.getSetting(c, "telegram_bot_first_name")
		if cachedUser != "" {
			c.JSON(http.StatusOK, gin.H{
				"configured":      true,
				"id":              cachedID,
				"username":        cachedUser,
				"first_name":      cachedName,
				"can_join_groups": true,
				"can_read_all":    false,
				"supports_inline": false,
				"warning":         "Koneksi ke Telegram API timeout. Menampilkan identitas bot dari cache.",
				"error":           nil,
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"configured": true,
			"error":      fmt.Sprintf("Gagal menghubungi Telegram API (%v). Periksa koneksi internet atau gunakan Telegram API URL proxy.", err),
		})
		return
	}

	// Cache successful getMe result
	h.pool.Exec(c, `INSERT INTO app_settings (key, value, updated_at) VALUES
		('telegram_bot_id', $1, now()),
		('telegram_bot_username', $2, now()),
		('telegram_bot_first_name', $3, now())
		ON CONFLICT (key) DO UPDATE SET value=EXCLUDED.value, updated_at=now()`,
		strconv.FormatInt(me.ID, 10), me.Username, me.FirstName)

	c.JSON(http.StatusOK, gin.H{
		"configured":      true,
		"id":              me.ID,
		"username":        me.Username,
		"first_name":      me.FirstName,
		"last_name":       me.LastName,
		"can_join_groups": me.CanJoinGroups,
		"can_read_all":    me.CanReadAllGroupMessages,
		"supports_inline": me.SupportInlineQueries,
	})
}

// ---- settings ----

// GET /api/settings/telegram
func (h *SettingHandler) GetTelegram(c *gin.Context) {
	rows, err := h.pool.Query(c, `SELECT key, value FROM app_settings WHERE key LIKE 'telegram_%' ORDER BY key`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	settings := map[string]string{}
	for rows.Next() {
		var k, v string
		rows.Scan(&k, &v)
		settings[k] = v
	}
	c.JSON(http.StatusOK, settings)
}

// PUT /api/settings/telegram
func (h *SettingHandler) UpdateTelegram(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	allowed := map[string]bool{
		"telegram_bot_token":        true,
		"telegram_api_url":          true,
		"telegram_alert_low_stock":  true,
		"telegram_alert_daily_time": true,
		"telegram_webhook_url":      true,
		"telegram_webhook_secret":   true,
		"telegram_mode":             true,
	}
	for k, v := range req {
		if !allowed[k] {
			continue
		}
		h.pool.Exec(c, `INSERT INTO app_settings (key, value, updated_at) VALUES ($1,$2,now()) ON CONFLICT (key) DO UPDATE SET value=$2, updated_at=now()`, k, v)
	}
	c.JSON(http.StatusOK, gin.H{"message": "settings updated"})
}

// ---- subscribers CRUD ----

// GET /api/settings/telegram/subscribers
func (h *SettingHandler) ListSubscribers(c *gin.Context) {
	rows, err := h.pool.Query(c, `SELECT chat_id, type, title, username, notify_in, notify_out, notify_adjust, notify_low_stock, active, created_at FROM telegram_subscribers ORDER BY created_at DESC`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	subs := []subscriber{}
	for rows.Next() {
		var s subscriber
		var ts time.Time
		rows.Scan(&s.ChatID, &s.Type, &s.Title, &s.Username, &s.NotifyIn, &s.NotifyOut, &s.NotifyAdjust, &s.NotifyLow, &s.Active, &ts)
		subs = append(subs, s)
	}
	c.JSON(http.StatusOK, subs)
}

// POST /api/settings/telegram/subscribers
func (h *SettingHandler) AddSubscriber(c *gin.Context) {
	var req struct {
		ChatID int64  `json:"chat_id" binding:"required"`
		Title  string `json:"title"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Try to resolve chat info via bot
	token := h.getSetting(c, "telegram_bot_token")
	chatType := "private"
	chatTitle := req.Title
	chatUsername := ""

	if token != "" {
		apiURL := h.getSetting(c, "telegram_api_url")
		botInst, err := newTelegramBot(token, apiURL)
		if err == nil {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
			defer cancel()
			info, err := botInst.GetChat(ctx, &bot.GetChatParams{ChatID: req.ChatID})
			if err == nil {
				switch info.Type {
				case "private":
					chatType = "private"
					chatTitle = info.FirstName + " " + info.LastName
					chatUsername = info.Username
				case "group", "supergroup":
					chatType = "group"
					chatTitle = info.Title
					chatUsername = info.Username
				case "channel":
					chatType = "channel"
					chatTitle = info.Title
					chatUsername = info.Username
				}
			}
		}
	}

	if chatTitle == "" {
		chatTitle = strconv.FormatInt(req.ChatID, 10)
	}

	_, err := h.pool.Exec(c,
		`INSERT INTO telegram_subscribers (chat_id, type, title, username, active)
		 VALUES ($1,$2,$3,$4,TRUE)
		 ON CONFLICT (chat_id) DO UPDATE SET type=$2, title=$3, username=$4, active=TRUE, updated_at=now()`,
		req.ChatID, chatType, chatTitle, chatUsername)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"chat_id": req.ChatID, "type": chatType, "title": chatTitle, "username": chatUsername})
}

// PUT /api/settings/telegram/subscribers/:chatId
func (h *SettingHandler) UpdateSubscriber(c *gin.Context) {
	chatIDStr := c.Param("chatId")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat_id"})
		return
	}
	var req struct {
		NotifyIn     *bool `json:"notify_in"`
		NotifyOut    *bool `json:"notify_out"`
		NotifyAdjust *bool `json:"notify_adjust"`
		NotifyLow    *bool `json:"notify_low_stock"`
		Active       *bool `json:"active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := []string{}
	args := []any{chatID}
	idx := 2
	if req.NotifyIn != nil {
		updates = append(updates, "notify_in=$"+strconv.Itoa(idx))
		args = append(args, *req.NotifyIn)
		idx++
	}
	if req.NotifyOut != nil {
		updates = append(updates, "notify_out=$"+strconv.Itoa(idx))
		args = append(args, *req.NotifyOut)
		idx++
	}
	if req.NotifyAdjust != nil {
		updates = append(updates, "notify_adjust=$"+strconv.Itoa(idx))
		args = append(args, *req.NotifyAdjust)
		idx++
	}
	if req.NotifyLow != nil {
		updates = append(updates, "notify_low_stock=$"+strconv.Itoa(idx))
		args = append(args, *req.NotifyLow)
		idx++
	}
	if req.Active != nil {
		updates = append(updates, "active=$"+strconv.Itoa(idx))
		args = append(args, *req.Active)
		idx++
	}
	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}
	updates = append(updates, "updated_at=now()")
	query := "UPDATE telegram_subscribers SET " + strings.Join(updates, ", ") + " WHERE chat_id=$1"
	h.pool.Exec(c, query, args...)
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// DELETE /api/settings/telegram/subscribers/:chatId
func (h *SettingHandler) DeleteSubscriber(c *gin.Context) {
	chatIDStr := c.Param("chatId")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat_id"})
		return
	}
	h.pool.Exec(c, `DELETE FROM telegram_subscribers WHERE chat_id=$1`, chatID)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ---- test ----

// POST /api/settings/telegram/test
func (h *SettingHandler) TestTelegram(c *gin.Context) {
	var req struct {
		ChatID  int64  `json:"chat_id"`
		Message string `json:"message"`
		All     bool   `json:"all"`
	}
	_ = c.ShouldBindJSON(&req)

	token := h.getSetting(c, "telegram_bot_token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bot token belum dikonfigurasi"})
		return
	}
	apiURL := h.getSetting(c, "telegram_api_url")

	botInst, err := newTelegramBot(token, apiURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "bot init gagal: " + err.Error()})
		return
	}

	msg := req.Message
	if msg == "" {
		msg = "✅ Test dari Inventaris Kantor — bot terhubung dengan benar."
	}

	type target struct {
		chatID int64
		title  string
	}
	var targets []target

	if req.All || req.ChatID == 0 {
		rows, _ := h.pool.Query(c, `SELECT chat_id, title FROM telegram_subscribers WHERE active=TRUE`)
		for rows.Next() {
			var t target
			rows.Scan(&t.chatID, &t.title)
			targets = append(targets, t)
		}
		rows.Close()
	} else {
		var title string
		h.pool.QueryRow(c, `SELECT title FROM telegram_subscribers WHERE chat_id=$1`, req.ChatID).Scan(&title)
		targets = append(targets, target{chatID: req.ChatID, title: title})
	}

	if len(targets) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tidak ada subscriber aktif"})
		return
	}

	results := []map[string]any{}
	for _, t := range targets {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		_, err := botInst.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: t.chatID,
			Text:   msg,
		})
		cancel()

		status := "ok"
		errMsg := ""
		if err != nil {
			status = "fail"
			errMsg = err.Error()
		}

		// Log to DB
		h.pool.Exec(c, `INSERT INTO telegram_message_log (chat_id, chat_title, message, status, error) VALUES ($1,$2,$3,$4,$5)`,
			t.chatID, t.title, msg, status, errMsg)

		results = append(results, map[string]any{
			"chat_id": strconv.FormatInt(t.chatID, 10),
			"title":   t.title,
			"status":  status,
			"error":   errMsg,
		})
	}

	c.JSON(http.StatusOK, gin.H{"results": results})
}

// ---- message log ----

// GET /api/settings/telegram/logs
func (h *SettingHandler) TelegramLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	perPage := configPerPage(c)
	offset := (page - 1) * perPage

	var total int
	h.pool.QueryRow(c, `SELECT count(*) FROM telegram_message_log`).Scan(&total)

	rows, err := h.pool.Query(c, `SELECT id, chat_id, chat_title, message, status, error, sent_at FROM telegram_message_log ORDER BY sent_at DESC LIMIT $1 OFFSET $2`, perPage, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type logEntry struct {
		ID        int64  `json:"id"`
		ChatID    int64  `json:"chat_id"`
		ChatTitle string `json:"chat_title"`
		Message   string `json:"message"`
		Status    string `json:"status"`
		Error     string `json:"error"`
		SentAt    string `json:"sent_at"`
	}
	logs := []logEntry{}
	for rows.Next() {
		var l logEntry
		var ts time.Time
		var errMsg *string
		rows.Scan(&l.ID, &l.ChatID, &l.ChatTitle, &l.Message, &l.Status, &errMsg, &ts)
		if errMsg != nil {
			l.Error = *errMsg
		}
		l.SentAt = ts.In(time.FixedZone("WIB", 7*3600)).Format("02-01-2006 15:04:05")
		logs = append(logs, l)
	}
	pages := total / perPage
	if total%perPage > 0 {
		pages++
	}
	c.JSON(http.StatusOK, paginatedResp{Data: logs, Total: total, Page: page, PerPage: perPage, Pages: pages})
}

// ---- helper ----

func (h *SettingHandler) getSetting(c context.Context, key string) string {
	var v string
	h.pool.QueryRow(c, `SELECT value FROM app_settings WHERE key=$1`, key).Scan(&v)
	return v
}

// GetActiveSubscribersForType returns subscribers that should receive a given notification type
func (h *SettingHandler) GetActiveSubscribersForType(ctx context.Context, notifType string) []subscriber {
	column := "notify_in"
	switch notifType {
	case "out":
		column = "notify_out"
	case "adjust":
		column = "notify_adjust"
	case "low_stock":
		column = "notify_low_stock"
	}
	rows, _ := h.pool.Query(ctx, `SELECT chat_id, type, title, username, notify_in, notify_out, notify_adjust, notify_low_stock, active FROM telegram_subscribers WHERE active=TRUE AND `+column+`=TRUE`)
	defer rows.Close()
	subs := []subscriber{}
	for rows.Next() {
		var s subscriber
		rows.Scan(&s.ChatID, &s.Type, &s.Title, &s.Username, &s.NotifyIn, &s.NotifyOut, &s.NotifyAdjust, &s.NotifyLow, &s.Active)
		subs = append(subs, s)
	}
	return subs
}

// GetSettingsMap returns all telegram settings (for internal use)
func (h *SettingHandler) GetSettingsMap(ctx context.Context) map[string]string {
	rows, _ := h.pool.Query(ctx, `SELECT key, value FROM app_settings WHERE key LIKE 'telegram_%'`)
	defer rows.Close()
	m := map[string]string{}
	for rows.Next() {
		var k, v string
		rows.Scan(&k, &v)
		m[k] = v
	}
	return m
}

// AutoRegisterSubscriber is called when someone /starts the bot or joins a group
func (h *SettingHandler) AutoRegisterSubscriber(ctx context.Context, chatID int64, info *models.ChatFullInfo) {
	if info == nil {
		return
	}
	chatType := "private"
	chatTitle := ""
	chatUsername := ""

	switch info.Type {
	case "private":
		chatType = "private"
		chatTitle = info.FirstName + " " + info.LastName
		chatUsername = info.Username
	case "group", "supergroup":
		chatType = "group"
		chatTitle = info.Title
		chatUsername = info.Username
	case "channel":
		chatType = "channel"
		chatTitle = info.Title
		chatUsername = info.Username
	}

	if chatTitle == "" {
		chatTitle = strconv.FormatInt(chatID, 10)
	}

	h.pool.Exec(ctx,
		`INSERT INTO telegram_subscribers (chat_id, type, title, username, active)
		 VALUES ($1,$2,$3,$4,TRUE)
		 ON CONFLICT (chat_id) DO UPDATE SET type=$2, title=$3, username=$4, updated_at=now()`,
		chatID, chatType, chatTitle, chatUsername)
}

// ==================== COMMAND CRUD ====================

type tgCommand struct {
	ID          int    `json:"id"`
	Command     string `json:"command"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Response    string `json:"response"`
	IsMenu      bool   `json:"is_menu"`
	SortOrder   int    `json:"sort_order"`
	Active      bool   `json:"active"`
}

// GET /api/settings/telegram/commands
func (h *SettingHandler) ListCommands(c *gin.Context) {
	rows, err := h.pool.Query(c, `SELECT id, command, label, description, response, is_menu, sort_order, active FROM telegram_commands ORDER BY sort_order, id`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	cmds := []tgCommand{}
	for rows.Next() {
		var cmd tgCommand
		rows.Scan(&cmd.ID, &cmd.Command, &cmd.Label, &cmd.Description, &cmd.Response, &cmd.IsMenu, &cmd.SortOrder, &cmd.Active)
		cmds = append(cmds, cmd)
	}
	c.JSON(http.StatusOK, cmds)
}

// POST /api/settings/telegram/commands
func (h *SettingHandler) CreateCommand(c *gin.Context) {
	var req tgCommand
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Command == "" || req.Label == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "command dan label wajib"})
		return
	}
	// strip leading /
	req.Command = strings.TrimPrefix(req.Command, "/")
	var id int
	err := h.pool.QueryRow(c,
		`INSERT INTO telegram_commands (command, label, description, response, is_menu, sort_order, active)
		 VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
		req.Command, req.Label, req.Description, req.Response, req.IsMenu, req.SortOrder, req.Active).
		Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// PUT /api/settings/telegram/commands/:id
func (h *SettingHandler) UpdateCommand(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req tgCommand
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Command = strings.TrimPrefix(req.Command, "/")
	_, err = h.pool.Exec(c,
		`UPDATE telegram_commands SET command=$2, label=$3, description=$4, response=$5, is_menu=$6, sort_order=$7, active=$8, updated_at=now() WHERE id=$1`,
		id, req.Command, req.Label, req.Description, req.Response, req.IsMenu, req.SortOrder, req.Active)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// DELETE /api/settings/telegram/commands/:id
func (h *SettingHandler) DeleteCommand(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	h.pool.Exec(c, `DELETE FROM telegram_commands WHERE id=$1`, id)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// GetMenuCommands returns commands that show in the inline menu (is_menu=true, active=true)
func (h *SettingHandler) GetMenuCommands(ctx context.Context) []tgCommand {
	rows, _ := h.pool.Query(ctx, `SELECT id, command, label, description, response, is_menu, sort_order, active FROM telegram_commands WHERE is_menu=TRUE AND active=TRUE ORDER BY sort_order`)
	defer rows.Close()
	cmds := []tgCommand{}
	for rows.Next() {
		var cmd tgCommand
		rows.Scan(&cmd.ID, &cmd.Command, &cmd.Label, &cmd.Description, &cmd.Response, &cmd.IsMenu, &cmd.SortOrder, &cmd.Active)
		cmds = append(cmds, cmd)
	}
	return cmds
}

// GetCommandByCmd returns a single command by its command string
func (h *SettingHandler) GetCommandByCmd(ctx context.Context, command string) *tgCommand {
	var cmd tgCommand
	err := h.pool.QueryRow(ctx, `SELECT id, command, label, description, response, is_menu, sort_order, active FROM telegram_commands WHERE command=$1 AND active=TRUE`, command).
		Scan(&cmd.ID, &cmd.Command, &cmd.Label, &cmd.Description, &cmd.Response, &cmd.IsMenu, &cmd.SortOrder, &cmd.Active)
	if err != nil {
		return nil
	}
	return &cmd
}