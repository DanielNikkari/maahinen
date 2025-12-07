package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

const (
	minInputHeight     = 1
	maxInputHeight     = 10
	toolPanelWidth     = 40
	commandMenuMaxShow = 7
)

// Message types for TUI communication
type (
	// SendMessageMsg is sent when user submits a message
	SendMessageMsg struct {
		Content string
	}

	// ResponseMsg is sent when we receive a response from the model
	ResponseMsg struct {
		Role    string
		Content string
	}

	// ToolCallMsg is sent when a tool is being executed (for display)
	ToolCallMsg struct {
		ID        string
		Name      string
		Arguments map[string]any
	}

	// ToolConfirmRequestMsg is sent when agent wants user to confirm a tool
	ToolConfirmRequestMsg struct {
		ID        string
		Name      string
		Arguments map[string]any
	}

	// ToolResultMsg is sent when a tool execution completes
	ToolResultMsg struct {
		ID      string
		Name    string
		Success bool
		Output  string
		Error   string
	}

	// ToolCancelledMsg is sent when a tool is cancelled by the user
	ToolCancelledMsg struct {
		ID        string
		Name      string
		Arguments map[string]any
	}

	// StreamChunkMsg is sent for streaming responses
	StreamChunkMsg struct {
		Content string
		Done    bool
	}

	// ErrorMsg is sent when an error occurs
	ErrorMsg struct {
		Error error
	}

	// ModelChangedMsg is sent when the model changes
	ModelChangedMsg struct {
		Model string
	}

	// UpdateLastMessageMsg updates the last message in place (for progress)
	UpdateLastMessageMsg struct {
		Content string
	}
)

// Command represents an available slash command
type Command struct {
	Name        string
	Description string
	HasSubcmds  bool
}

var availableCommands = []Command{
	{Name: "/model", Description: "Show current model", HasSubcmds: true},
	{Name: "/model/list", Description: "List installed models", HasSubcmds: false},
	{Name: "/spinner", Description: "Show current spinner", HasSubcmds: true},
	{Name: "/spinner/list", Description: "List available spinners", HasSubcmds: false},
	{Name: "/prune", Description: "Clear message history", HasSubcmds: false},
	{Name: "/autoconfirm", Description: "Toggle tool auto-confirm on/off.", HasSubcmds: false},
	{Name: "/help", Description: "Show available commands", HasSubcmds: false},
}

// ChatMessage represents a message in the chat history
type ChatMessage struct {
	Role    string
	Content string
}

// ToolCallRecord represents a tool call in the tool panel
type ToolCallRecord struct {
	ID        string
	Name      string
	Arguments map[string]any
	Status    string // "pending", "running", "success", "error"
	Output    string
	Error     string
}

// Model is the main TUI model
type Model struct {
	// Dimensions
	width  int
	height int

	// Panels
	messageViewport viewport.Model
	chatInput       textarea.Model
	toolCalls       []ToolCallRecord

	// State
	messages         []ChatMessage
	showToolPanel    bool
	showCommandMenu  bool
	commandMenuIndex int
	filteredCommands []Command
	currentModel     string
	isProcessing     bool
	streamBuffer     strings.Builder
	autoConfirmTools bool

	// Confirmation dialog
	showConfirmDialog   bool
	pendingToolCall     *ToolCallMsg
	confirmDialogChoice int // 0 = confirm, 1 = deny

	// Markdown renderer
	mdRenderer *glamour.TermRenderer

	// Spinner for "Thinking..." animation
	spinnerStyle  string
	spinnerFrames []string
	spinnerIndex  int

	// Callbacks (set by the integrating code)
	onSendMessage       func(string)
	onToolConfirm       func(bool)
	onAutoConfirmToggle func(bool)
	onPrune             func()
}

// NewModel creates a new TUI model
func NewModel() *Model {
	// Create textarea for chat input
	ti := textarea.New()
	ti.Placeholder = "Type a message or / for commands..."
	ti.Focus()
	ti.CharLimit = 0 // No limit
	ti.SetHeight(minInputHeight)
	ti.ShowLineNumbers = false
	ti.KeyMap.InsertNewline.SetEnabled(false) // We'll handle enter ourselves

	// Create viewport for message history
	vp := viewport.New(80, 20)
	vp.SetContent("")

	// Create markdown renderer with dark theme (avoid WithAutoStyle which queries terminal)
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithStylePath("dark"),
		glamour.WithWordWrap(80),
	)

	// Default spinner frames
	spinnerFrames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

	return &Model{
		chatInput:        ti,
		messageViewport:  vp,
		messages:         []ChatMessage{},
		toolCalls:        []ToolCallRecord{},
		showToolPanel:    true,
		filteredCommands: availableCommands,
		mdRenderer:       renderer,
		autoConfirmTools: false,
		spinnerStyle:     "dots",
		spinnerFrames:    spinnerFrames,
		spinnerIndex:     0,
	}
}

// SetModel sets the current model name
func (m *Model) SetModel(model string) {
	m.currentModel = model
}

// SetAutoConfirmTools sets whether tools should be auto-confirmed
func (m *Model) SetAutoConfirmTools(auto bool) {
	m.autoConfirmTools = auto
}

// SetOnSendMessage sets the callback for when user sends a message
func (m *Model) SetOnSendMessage(fn func(string)) {
	m.onSendMessage = fn
}

// SetOnToolConfirm sets the callback for when user confirms/denies a tool
func (m *Model) SetOnToolConfirm(fn func(bool)) {
	m.onToolConfirm = fn
}

// SetOnAutoConfirmToggle sets the callback for when auto-confirm is toggled
func (m *Model) SetOnAutoConfirmToggle(fn func(bool)) {
	m.onAutoConfirmToggle = fn
}

// SetOnPrune sets the callback for when /prune is called
func (m *Model) SetOnPrune(fn func()) {
	m.onPrune = fn
}

// ClearMessages clears all messages and tool calls from the UI
func (m *Model) ClearMessages() {
	m.messages = []ChatMessage{}
	m.toolCalls = []ToolCallRecord{}
	m.streamBuffer.Reset()
	m.renderMessages()
}

// SetSpinnerStyle sets the spinner animation style
func (m *Model) SetSpinnerStyle(style string) {
	m.spinnerStyle = style
	// Map style to frames - only text-based spinners that render well in terminals
	spinnerMap := map[string][]string{
		"dots":    {"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		"bounce":  {"⠁", "⠂", "⠄", "⠂"},
		"classic": {"|", "/", "-", "\\"},
		"line":    {"—", "\\", "|", "/"},
		"pulse":   {"○", "◔", "◑", "◕", "●", "◕", "◑", "◔"},
	}
	if frames, ok := spinnerMap[style]; ok {
		m.spinnerFrames = frames
	}
	m.spinnerIndex = 0
}

// GetAvailableSpinners returns list of available spinner styles
func GetAvailableSpinners() []string {
	return []string{"dots", "bounce", "classic", "line", "pulse"}
}

// spinnerTickMsg is sent to animate the spinner
type spinnerTickMsg struct{}

// tickSpinner returns a command that sends a spinner tick after a delay
func tickSpinner() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return spinnerTickMsg{}
	})
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return textarea.Blink
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case spinnerTickMsg:
		// Animate spinner while processing
		if m.isProcessing {
			m.spinnerIndex = (m.spinnerIndex + 1) % len(m.spinnerFrames)
			m.renderMessages()
			return m, tickSpinner()
		}
		return m, nil

	case ResponseMsg:
		m.isProcessing = false // Must be set BEFORE addMessage which calls renderMessages
		m.addMessage(msg.Role, msg.Content)
		return m, nil

	case ToolCallMsg:
		// Tool is being executed - add to panel with "running" status
		m.toolCalls = append(m.toolCalls, ToolCallRecord{
			ID:        msg.ID,
			Name:      msg.Name,
			Arguments: msg.Arguments,
			Status:    "running",
		})
		return m, nil

	case ToolConfirmRequestMsg:
		// Agent wants user to confirm a tool - show dialog
		m.pendingToolCall = &ToolCallMsg{
			ID:        msg.ID,
			Name:      msg.Name,
			Arguments: msg.Arguments,
		}
		m.showConfirmDialog = true
		m.confirmDialogChoice = 0
		return m, nil

	case ToolResultMsg:
		m.handleToolResult(msg)
		m.isProcessing = true
		m.spinnerIndex = 0
		m.renderMessages()
		return m, tickSpinner()

	case ToolCancelledMsg:
		// Add cancelled tool to the panel
		m.toolCalls = append(m.toolCalls, ToolCallRecord{
			ID:        msg.ID,
			Name:      msg.Name,
			Arguments: msg.Arguments,
			Status:    "cancelled",
		})
		return m, nil

	case StreamChunkMsg:
		m.streamBuffer.WriteString(msg.Content)
		if msg.Done {
			m.isProcessing = false // Must be set BEFORE addMessage which calls renderMessages
			m.addMessage("assistant", m.streamBuffer.String())
			m.streamBuffer.Reset()
		} else {
			m.updateStreamingMessage()
		}
		return m, nil

	case ErrorMsg:
		m.isProcessing = false // Must be set BEFORE addMessage which calls renderMessages
		m.addMessage("system", fmt.Sprintf("Error: %v", msg.Error))
		return m, nil

	case ModelChangedMsg:
		m.currentModel = msg.Model
		return m, nil

	case UpdateLastMessageMsg:
		// Update the last message in place (for progress updates)
		if len(m.messages) > 0 {
			m.messages[len(m.messages)-1].Content = msg.Content
			m.renderMessages()
		}
		return m, nil
	}

	// Update chat input
	var cmd tea.Cmd
	m.chatInput, cmd = m.chatInput.Update(msg)
	cmds = append(cmds, cmd)

	// Check for command menu trigger
	m.updateCommandMenu()

	// Update input height dynamically
	m.updateInputHeight()

	return m, tea.Batch(cmds...)
}

func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle confirmation dialog
	if m.showConfirmDialog {
		return m.handleConfirmDialogKey(msg)
	}

	// Filter out escape sequences that leak from terminal responses
	// These include OSC sequences like "]11rgb:...", CSI responses like "[56;1R",
	// and partial escape sequences containing terminal response fragments
	keyStr := msg.String()
	if len(keyStr) > 0 {
		first := keyStr[0]
		// Filter: ESC, CSI intro, OSC intro, ST, or sequences containing terminal response patterns
		if first == ']' || first == '[' || first == '\x1b' || first == '\x9c' || first == '\\' {
			return m, nil
		}
		// Also filter strings that look like partial terminal responses (contain rgb:, ;, R at end)
		if strings.Contains(keyStr, "rgb:") || strings.Contains(keyStr, ";") && strings.HasSuffix(keyStr, "R") {
			return m, nil
		}
		// Filter anything with / that looks like color values (e.g., "2727/3a3a")
		if strings.Count(keyStr, "/") >= 2 && !strings.HasPrefix(keyStr, "/") {
			return m, nil
		}
	}

	switch keyStr {
	case "ctrl+c":
		// If there's selected text, copy it; otherwise quit
		// Note: Terminal selection/copy is handled by the terminal itself
		// ctrl+c without selection should quit
		return m, tea.Quit

	case "ctrl+t":
		// Toggle tool panel
		m.showToolPanel = !m.showToolPanel
		m.updateLayout()
		return m, nil

	case "ctrl+a":
		// Toggle auto-confirm tools
		m.autoConfirmTools = !m.autoConfirmTools
		if m.onAutoConfirmToggle != nil {
			m.onAutoConfirmToggle(m.autoConfirmTools)
		}
		return m, nil

	case "ctrl+v":
		// Paste is handled by the textarea component by default
		// Just pass through to the textarea
		var cmd tea.Cmd
		m.chatInput, cmd = m.chatInput.Update(msg)
		m.updateCommandMenu()
		m.updateInputHeight()
		return m, cmd

	case "enter":
		if m.showCommandMenu && len(m.filteredCommands) > 0 {
			cmd := m.filteredCommands[m.commandMenuIndex]
			currentText := strings.TrimSpace(m.chatInput.Value())

			// If current text matches selected command, execute it
			if currentText == cmd.Name {
				m.showCommandMenu = false
				m.commandMenuIndex = 0
				// Fall through to submit the command
			} else {
				// Otherwise, just fill in the command name
				m.chatInput.SetValue(cmd.Name)
				m.showCommandMenu = false
				m.commandMenuIndex = 0
				return m, nil
			}
		}

		// Submit message
		content := strings.TrimSpace(m.chatInput.Value())
		if content != "" && !m.isProcessing {
			m.chatInput.Reset()
			m.chatInput.SetHeight(minInputHeight)
			m.isProcessing = true
			m.spinnerIndex = 0

			if m.onSendMessage != nil {
				m.onSendMessage(content)
			}

			// Add user message to history
			m.addMessage("user", content)

			// Start spinner animation
			return m, tickSpinner()
		}
		return m, nil

	case "shift+enter":
		// Insert newline
		m.chatInput.SetValue(m.chatInput.Value() + "\n")
		m.updateInputHeight()
		return m, nil

	case "up":
		if m.showCommandMenu {
			if m.commandMenuIndex > 0 {
				m.commandMenuIndex--
			}
			return m, nil
		}
		// Scroll viewport up
		m.messageViewport.ScrollUp(1)
		return m, nil

	case "down":
		if m.showCommandMenu {
			if m.commandMenuIndex < len(m.filteredCommands)-1 {
				m.commandMenuIndex++
			}
			return m, nil
		}
		// Scroll viewport down
		m.messageViewport.ScrollDown(1)
		return m, nil

	case "pgup":
		m.messageViewport.HalfPageUp()
		return m, nil

	case "pgdown":
		m.messageViewport.HalfPageDown()
		return m, nil

	case "esc":
		if m.showCommandMenu {
			m.showCommandMenu = false
			m.commandMenuIndex = 0
			return m, nil
		}
		return m, nil

	case "tab":
		if m.showCommandMenu && len(m.filteredCommands) > 0 {
			// Autocomplete selected command
			cmd := m.filteredCommands[m.commandMenuIndex]
			m.chatInput.SetValue(cmd.Name)
			if cmd.HasSubcmds {
				m.chatInput.SetValue(cmd.Name + "/")
			}
			m.showCommandMenu = false
			return m, nil
		}
	}

	// Default: update textarea
	var cmd tea.Cmd
	m.chatInput, cmd = m.chatInput.Update(msg)
	m.updateCommandMenu()
	m.updateInputHeight()
	return m, cmd
}

func (m *Model) handleConfirmDialogKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		// Move selection up (to Yes)
		if m.confirmDialogChoice > 0 {
			m.confirmDialogChoice--
		}
		return m, nil
	case "down", "j":
		// Move selection down (to No)
		if m.confirmDialogChoice < 1 {
			m.confirmDialogChoice++
		}
		return m, nil
	case "enter":
		confirmed := m.confirmDialogChoice == 0
		m.showConfirmDialog = false
		m.pendingToolCall = nil
		if m.onToolConfirm != nil {
			m.onToolConfirm(confirmed)
		}
		return m, nil
	case "y":
		m.showConfirmDialog = false
		m.pendingToolCall = nil
		if m.onToolConfirm != nil {
			m.onToolConfirm(true)
		}
		return m, nil
	case "n", "esc":
		m.showConfirmDialog = false
		m.pendingToolCall = nil
		if m.onToolConfirm != nil {
			m.onToolConfirm(false)
		}
		return m, nil
	}
	return m, nil
}

func (m *Model) updateLayout() {
	if m.width == 0 || m.height == 0 {
		return
	}

	// Calculate panel dimensions
	inputHeight := m.getInputHeight() + 2 // +2 for border
	statusHeight := 1
	headerHeight := 1

	contentWidth := m.width
	if m.showToolPanel {
		contentWidth = m.width - toolPanelWidth - 1 // -1 for gap
	}

	viewportHeight := m.height - inputHeight - statusHeight - headerHeight - 2

	// Update viewport
	m.messageViewport.Width = contentWidth - 2 // -2 for border
	m.messageViewport.Height = viewportHeight

	// Update chat input width
	m.chatInput.SetWidth(contentWidth - 4) // -4 for border and padding

	// Re-render markdown with new width
	if m.mdRenderer != nil {
		m.mdRenderer, _ = glamour.NewTermRenderer(
			glamour.WithStylePath("dark"),
			glamour.WithWordWrap(contentWidth-6),
		)
	}

	// Re-render messages
	m.renderMessages()
}

func (m *Model) updateInputHeight() {
	// Use the textarea's own line count which accounts for wrapping
	lines := m.chatInput.LineCount()
	if lines == 0 {
		lines = 1
	}

	newHeight := min(max(lines, minInputHeight), maxInputHeight)
	oldHeight := m.chatInput.Height()

	if newHeight != oldHeight {
		// Get current cursor position info
		currentLine := m.chatInput.Line()
		currentValue := m.chatInput.Value()

		// Set new height
		m.chatInput.SetHeight(newHeight)

		// If we're expanding and have multiple lines, reset viewport to show all content
		if newHeight > oldHeight && lines > 1 {
			// Move cursor to start to reset viewport, then restore position
			m.chatInput.CursorStart()

			// If we weren't on the first line, move back toward original position
			// by moving down to the original line
			for i := 0; i < currentLine && i < lines-1; i++ {
				m.chatInput.CursorDown()
			}

			// Set cursor to end of value to ensure we're at the typing position
			m.chatInput.SetValue(currentValue)
			m.chatInput.CursorEnd()
		}

		m.updateLayout()
	}
}

func (m *Model) getInputHeight() int {
	lines := m.chatInput.LineCount()
	if lines == 0 {
		lines = 1
	}
	return min(max(lines, minInputHeight), maxInputHeight)
}

func (m *Model) updateCommandMenu() {
	value := m.chatInput.Value()

	if strings.HasPrefix(value, "/") {
		m.showCommandMenu = true
		m.filterCommands(value)
	} else {
		m.showCommandMenu = false
		m.commandMenuIndex = 0
	}
}

func (m *Model) filterCommands(prefix string) {
	m.filteredCommands = []Command{}
	for _, cmd := range availableCommands {
		if strings.HasPrefix(cmd.Name, prefix) {
			m.filteredCommands = append(m.filteredCommands, cmd)
		}
	}

	// Reset index if out of bounds
	if m.commandMenuIndex >= len(m.filteredCommands) {
		m.commandMenuIndex = max(0, len(m.filteredCommands)-1)
	}
}

func (m *Model) addMessage(role, content string) {
	if role == "assistant" {
		content = "\n" + content
	}
	m.messages = append(m.messages, ChatMessage{
		Role:    role,
		Content: content,
	})
	m.renderMessages()
	m.messageViewport.GotoBottom()
}

func (m *Model) updateStreamingMessage() {
	// Update the last message or add streaming indicator
	m.renderMessages()
	m.messageViewport.GotoBottom()
}

func (m *Model) renderMessages() {
	var sb strings.Builder

	// Calculate available width for message content
	contentWidth := m.messageViewport.Width
	if contentWidth <= 0 {
		contentWidth = 80 // Default fallback
	}

	for _, msg := range m.messages {
		switch msg.Role {
		case "user":
			sb.WriteString(UserLabelStyle.Render("You") + "\n")
			// Apply width constraint to enable word wrapping
			userMsgStyled := UserMessageStyle.Width(contentWidth).Render(msg.Content)
			sb.WriteString(userMsgStyled + "\n\n")
		case "assistant":
			sb.WriteString(AssistantLabelStyle.Render("Maahinen") + "\n")
			// Render markdown
			if m.mdRenderer != nil {
				rendered, err := m.mdRenderer.Render(msg.Content)
				if err == nil {
					sb.WriteString(rendered)
				} else {
					sb.WriteString(AssistantMessageStyle.Width(contentWidth).Render(msg.Content) + "\n")
				}
			} else {
				sb.WriteString(AssistantMessageStyle.Width(contentWidth).Render(msg.Content) + "\n")
			}
			sb.WriteString("\n")
		case "system":
			sb.WriteString(SystemMessageStyle.Width(contentWidth).Render(msg.Content) + "\n\n")
		case "tool":
			sb.WriteString(ToolMessageStyle.Width(contentWidth).Render(msg.Content) + "\n\n")
		case "toolcall":
			// Show tool calls as one-liners
			sb.WriteString(ToolCallPrefixStyle.Render("⚡") + " " + ToolCallOneLineStyle.Render(msg.Content) + "\n")
		case "toolcall_failed":
			// Show failed tool calls in red
			sb.WriteString(ToolCallPrefixStyle.Render("⚡") + " " + ToolCallFailedStyle.Render(msg.Content) + "\n")
		case "toolcall_cancelled":
			// Show cancelled tool calls in dim
			sb.WriteString(ToolCallPrefixStyle.Render("⚡") + " " + ToolCallCancelledStyle.Render(msg.Content) + "\n")
		}
	}

	// Add spinner indicator while processing
	if m.isProcessing {
		spinnerFrame := m.spinnerFrames[m.spinnerIndex%len(m.spinnerFrames)]
		if m.streamBuffer.Len() > 0 {
			// Render streaming content as markdown in real-time
			sb.WriteString(AssistantLabelStyle.Render("Maahinen") + "\n")
			streamContent := m.streamBuffer.String()
			if m.mdRenderer != nil {
				rendered, err := m.mdRenderer.Render(streamContent)
				if err == nil {
					sb.WriteString(rendered)
				} else {
					sb.WriteString(AssistantMessageStyle.Width(contentWidth).Render(streamContent) + "\n")
				}
			} else {
				sb.WriteString(AssistantMessageStyle.Width(contentWidth).Render(streamContent) + "\n")
			}
			sb.WriteString(SpinnerStyle.Render(spinnerFrame) + "\n")
		} else {
			sb.WriteString(SpinnerStyle.Render(spinnerFrame+" Thinking...") + "\n")
		}
	}

	m.messageViewport.SetContent(sb.String())
}

func (m *Model) handleToolResult(tr ToolResultMsg) {
	for i := range m.toolCalls {
		if m.toolCalls[i].ID == tr.ID {
			if tr.Success {
				m.toolCalls[i].Status = "success"
				m.toolCalls[i].Output = tr.Output
			} else {
				m.toolCalls[i].Status = "error"
				m.toolCalls[i].Error = tr.Error
			}
			break
		}
	}
}

// View renders the TUI
func (m *Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Build the main content area
	header := m.renderHeader()
	messagePanel := m.renderMessagePanel()
	chatPanel := m.renderChatPanel()
	statusBar := m.renderStatusBar()

	// If confirmation dialog is showing, overlay it on the message panel
	if m.showConfirmDialog && m.pendingToolCall != nil {
		messagePanel = m.overlayConfirmDialog(messagePanel)
	}

	// Stack message panel and chat panel
	leftPanel := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		messagePanel,
		chatPanel,
	)

	var mainView string
	if m.showToolPanel {
		toolPanel := m.renderToolPanel()
		mainView = lipgloss.JoinHorizontal(
			lipgloss.Top,
			leftPanel,
			" ",
			toolPanel,
		)
	} else {
		mainView = leftPanel
	}

	// Add status bar at the bottom
	mainView = lipgloss.JoinVertical(lipgloss.Left, mainView, statusBar)

	// Overlay command menu if visible
	if m.showCommandMenu {
		mainView = m.overlayCommandMenu(mainView)
	}

	return mainView
}

func (m *Model) renderHeader() string {
	title := HeaderStyle.Render("Maahinen")
	model := ModelIndicatorStyle.Render(fmt.Sprintf("[%s]", m.currentModel))

	// Separator style (always dimmed)
	sep := HelpStyle.Render(" | ")

	// Tool panel toggle
	toolPanelHint := ""
	if m.showToolPanel {
		toolPanelHint = ToolPanelOnStyle.Render("ctrl+t: hide tools")
	} else {
		toolPanelHint = HelpStyle.Render("ctrl+t: show tools")
	}

	// Auto-confirm toggle
	autoConfirmHint := ""
	if m.autoConfirmTools {
		autoConfirmHint = AutoConfirmOnStyle.Render("tool auto-confirm (ctrl+a): ON")
	} else {
		autoConfirmHint = HelpStyle.Render("tool auto-confirm (ctrl+a): OFF")
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		title,
		" ",
		model,
		sep,
		toolPanelHint,
		sep,
		autoConfirmHint,
	)
}

func (m *Model) renderMessagePanel() string {
	contentWidth := m.width
	if m.showToolPanel {
		contentWidth = m.width - toolPanelWidth - 1
	}

	return MessagePanelStyle.
		Width(contentWidth - 2).
		Height(m.messageViewport.Height).
		Render(m.messageViewport.View())
}

func (m *Model) renderChatPanel() string {
	contentWidth := m.width
	if m.showToolPanel {
		contentWidth = m.width - toolPanelWidth - 1
	}

	style := ChatInputStyle
	if m.chatInput.Focused() {
		style = ChatInputFocusedStyle
	}

	return style.
		Width(contentWidth - 2).
		Render(m.chatInput.View())
}

func (m *Model) renderToolPanel() string {
	var sb strings.Builder

	sb.WriteString(ToolNameStyle.Render("Tool Calls") + "\n")
	sb.WriteString(strings.Repeat("─", toolPanelWidth-4) + "\n")

	if len(m.toolCalls) == 0 {
		sb.WriteString(HelpStyle.Render("No tool calls yet\n"))
	} else {
		// Show most recent calls (up to fit the panel)
		maxShow := (m.height - 8)
		start := max(0, len(m.toolCalls)-maxShow)

		for _, tc := range m.toolCalls[start:] {
			// Format tool call as one line: "name (status/args)"
			switch tc.Status {
			case "cancelled":
				// Show cancelled tools in red with message
				sb.WriteString(ToolCancelledStyle.Render(tc.Name+" - cancelled") + "\n")
			case "error":
				sb.WriteString(ToolErrorStyle.Render(tc.Name) + "\n")
				if tc.Error != "" {
					errMsg := tc.Error
					if len(errMsg) > toolPanelWidth-6 {
						errMsg = errMsg[:toolPanelWidth-9] + "..."
					}
					sb.WriteString(ToolArgsStyle.Render("  "+errMsg) + "\n")
				}
			default:
				// pending, running, success - show with args
				var nameStyled string
				switch tc.Status {
				case "pending":
					nameStyled = ToolPendingStyle.Render(tc.Name)
				case "running":
					nameStyled = ToolRunningStyle.Render(tc.Name)
				case "success":
					nameStyled = ToolSuccessStyle.Render(tc.Name)
				default:
					nameStyled = tc.Name
				}
				sb.WriteString(nameStyled + "\n")
				if len(tc.Arguments) > 0 {
					// Format each argument on its own line to prevent wrapping issues
					for k, v := range tc.Arguments {
						argStr := fmt.Sprintf("%s=%v", k, v)
						maxArgLen := toolPanelWidth - 8
						if maxArgLen < 10 {
							maxArgLen = 10
						}
						if len(argStr) > maxArgLen {
							argStr = argStr[:maxArgLen-3] + "..."
						}
						sb.WriteString(ToolArgsStyle.Render("  "+argStr) + "\n")
					}
				}
			}
		}
	}

	return ToolPanelStyle.
		Width(toolPanelWidth).
		Height(m.height - 4).
		Render(sb.String())
}

func (m *Model) renderStatusBar() string {
	status := ""
	if m.isProcessing {
		status = SpinnerStyle.Render("Processing...")
	} else {
		status = HelpStyle.Render("Enter: send | Shift+Enter: newline | /: commands")
	}

	return StatusBarStyle.Width(m.width).Render(status)
}

func (m *Model) overlayCommandMenu(base string) string {
	if len(m.filteredCommands) == 0 {
		return base
	}

	var sb strings.Builder

	showCount := min(len(m.filteredCommands), commandMenuMaxShow)
	for i := 0; i < showCount; i++ {
		cmd := m.filteredCommands[i]

		style := CommandItemStyle
		if i == m.commandMenuIndex {
			style = CommandItemSelectedStyle
		}

		line := style.Render(cmd.Name)
		desc := CommandDescStyle.Render(" " + cmd.Description)
		sb.WriteString(line + desc + "\n")
	}

	menu := CommandMenuStyle.Render(strings.TrimRight(sb.String(), "\n"))

	// Position the menu above the input
	// This is a simplified overlay - in practice you'd want proper positioning
	return base + "\n" + menu
}

// overlayConfirmDialog overlays a simple confirmation prompt on the message panel
func (m *Model) overlayConfirmDialog(messagePanel string) string {
	if m.pendingToolCall == nil {
		return messagePanel
	}

	// Get the message panel dimensions
	panelWidth := m.width
	if m.showToolPanel {
		panelWidth = m.width - toolPanelWidth - 1
	}
	panelHeight := m.messageViewport.Height + 2 // +2 for border

	// Calculate max dialog width (leave some margin)
	maxDialogWidth := panelWidth - 10
	if maxDialogWidth < 40 {
		maxDialogWidth = 40
	}

	// Build simple confirmation prompt
	var sb strings.Builder

	// Check if one-liner fits
	argsStr := formatToolArgs(m.pendingToolCall.Arguments, maxDialogWidth)
	oneLiner := fmt.Sprintf("Confirm tool call %s(%s)?", m.pendingToolCall.Name, argsStr)

	if len(oneLiner) <= maxDialogWidth {
		// Fits on one line
		sb.WriteString(DialogTitleStyle.Render(oneLiner) + "\n\n")
	} else {
		// Multi-line format
		sb.WriteString(DialogTitleStyle.Render(fmt.Sprintf("Confirm tool call %s?", m.pendingToolCall.Name)) + "\n")
		// Show each argument on its own line
		for k, v := range m.pendingToolCall.Arguments {
			valStr := fmt.Sprintf("%v", v)
			maxValLen := maxDialogWidth - len(k) - 4
			if maxValLen < 20 {
				maxValLen = 20
			}
			if len(valStr) > maxValLen {
				valStr = valStr[:maxValLen-3] + "..."
			}
			sb.WriteString(ToolArgsStyle.Render(fmt.Sprintf("  %s: %s", k, valStr)) + "\n")
		}
		sb.WriteString("\n")
	}

	// Simple yes/no selection
	if m.confirmDialogChoice == 0 {
		sb.WriteString(ConfirmYesSelectedStyle.Render("> yes") + "\n")
		sb.WriteString(ConfirmNoStyle.Render("  no") + "\n")
	} else {
		sb.WriteString(ConfirmYesStyle.Render("  yes") + "\n")
		sb.WriteString(ConfirmNoSelectedStyle.Render("> no") + "\n")
	}

	// Wrap in simple dialog style
	dialog := SimpleDialogStyle.Render(sb.String())

	// Center the dialog within the message panel area
	return lipgloss.Place(
		panelWidth,
		panelHeight,
		lipgloss.Center,
		lipgloss.Center,
		dialog,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(ColorBackground),
	)
}

func formatToolArgs(args map[string]any, maxWidth int) string {
	if len(args) == 0 {
		return ""
	}

	var parts []string
	for k, v := range args {
		s := fmt.Sprintf("%s=%v", k, v)
		if len(s) > 30 {
			s = s[:27] + "..."
		}
		parts = append(parts, s)
	}

	result := strings.Join(parts, ", ")
	if len(result) > maxWidth {
		result = result[:maxWidth-3] + "..."
	}
	return result
}

// AddToolCall adds a tool call to the panel (called by agent)
func (m *Model) AddToolCall(id, name string, args map[string]any) {
	m.toolCalls = append(m.toolCalls, ToolCallRecord{
		ID:        id,
		Name:      name,
		Arguments: args,
		Status:    "pending",
	})
}

// UpdateToolStatus updates a tool call's status (called by agent)
func (m *Model) UpdateToolStatus(id, status, output, errMsg string) {
	for i := range m.toolCalls {
		if m.toolCalls[i].ID == id {
			m.toolCalls[i].Status = status
			m.toolCalls[i].Output = output
			m.toolCalls[i].Error = errMsg
			break
		}
	}
}

// GetMessages returns all chat messages (for agent integration)
func (m *Model) GetMessages() []ChatMessage {
	return m.messages
}

// UpdateToolCallStatus updates a tool call message status in the history
func (m *Model) UpdateToolCallStatus(toolName, newRole, newContent string) {
	for i := len(m.messages) - 1; i >= 0; i-- {
		if m.messages[i].Role == "toolcall" && strings.HasPrefix(m.messages[i].Content, toolName+"(") {
			m.messages[i].Role = newRole
			m.messages[i].Content = newContent
			m.renderMessages()
			break
		}
	}
}

// StartProgram starts the TUI program
func StartProgram() (*tea.Program, *Model, error) {
	m := NewModel()
	p := tea.NewProgram(m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(), // Enable mouse for selection
	)
	return p, m, nil
}
