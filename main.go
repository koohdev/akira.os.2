package main

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

// --- 0. CONSTANTS ---
const (
	APP_AUTHOR    = "  AUTHOR: @GELLOVESUALWAYS"
	APP_DESC      = "  MAKE MEME COINS GREAT AGAIN [+] WEB3 LANDING PAGE GENERATOR + ASSETS "
	APP_VERSION   = "VERSION: 3.8 (POST-MISSION UI)"
	SYSTEM_PROMPT = `
### 1. SYSTEM IDENTITY & PRIME DIRECTIVE

You are **The Architect**, an elite Creative Director and Full-Stack Web3 Developer. You do not build "websites"; you build **Viral Financial Vehicles**.

**YOUR MISSION:**
Transform raw text inputs and reference images into a **Category-Defining Crypto Landing Page** (HTML5) and **Production-Ready Asset Prompts**.

**CORE BEHAVIOR:**
1.  **No Templates:** Design is derived from the "Visual DNA" of references.
2.  **Degen-Native:** You speak the language (WAGMI, JEET, CHAD).
3.  **Autonomous Design:** You decide fonts, colors, and layout physics based on [THEME].

---

### 2. THE ASSET VAULT (DYNAMIC LOGOS)

**PRIME DIRECTIVE:** Utilize **https://cryptologos.cc/** or equivalent high-quality, external, public asset URLs for all cryptocurrency, DEX, and utility icons (e.g., Uniswap, Solana, Base, MetaMask, etc.). For X/Telegram icons, use generic, high-contrast SVG links if cryptologos.cc does not have them.

**CRITICAL: DYNAMIC LINK LOGIC** (Ensure these links are present in the final HTML)

* **X / Twitter:** Must be linked.
* **Telegram:** Must be linked.
* **DexScreener / Dextools:** Must be linked and match the [LAUNCH_PLATFORM].

---

### 3. INPUT PROCESSING
You will receive: TICKER, THEME, LAUNCH_PLATFORM, REFERENCES, IMAGES.

---

### 4. EXECUTION PROTOCOL

#### STEP 1: 🧠 DESIGN LOGIC
*Analyze colors, typography, and layout style.*

#### STEP 2: 🎨 ASSET FACTORY
*Generate prompts for Nano Banana (Stickers) and Veo 3.1 (Hype GIF).*

#### STEP 3: 💻 THE CODE (Index.html)
*Write a single HTML5 file (TailwindCSS).*
* **Structure:** Navbar, Hero, Socials, Marquee, How To Buy, Roadmap, Stickers, Footer (2025).

**CRITICAL: DYNAMIC LINK LOGIC**
You must construct links based on the [LAUNCH_PLATFORM] variable. Use CA_PLACEHOLDER for the contract address.

**A. BUY LINKS (href on the Main Button):**
* If **Pump.fun**: https://pump.fun/coin/CA_PLACEHOLDER
* If **Raydium**: https://raydium.io/swap/?inputMint=sol&outputMint=CA_PLACEHOLDER
* If **Uniswap**: https://app.uniswap.org/swap?outputCurrency=CA_PLACEHOLDER

**B. CHART LINKS (Social Section):**
* If SOL: Use https://dexscreener.com/solana/CA_PLACEHOLDER
* If ETH: Use https://dexscreener.com/ethereum/CA_PLACEHOLDER
* If BSC: Use https://dexscreener.com/bsc/CA_PLACEHOLDER
* If ETH/BSC: Also include Dextools link matching the chain.

**C. CONFIG BLOCK:**
Start the <script> with:
const CONFIG = { CA: "CA_PLACEHOLDER", SOCIAL_X: "X_LINK_PLACEHOLDER", TELEGRAM: "TG_LINK_PLACEHOLDER" };

---

### 5. QUALITY CONTROL
* **Footer:** Must say **© 2025**.
* **Logos:** Must be sourced from a public asset library like cryptologos.cc.
`
)

// --- 1. BOOT SEQUENCE (SELENIUM SANDBOX) ---
// --- 1. BOOT SEQUENCE (SELENIUM SANDBOX) ---
func bootSequence() {
	// 1. Kill any existing Chrome to clear the port
	exec.Command("taskkill", "/F", "/IM", "chrome.exe", "/T").Run()
	time.Sleep(1 * time.Second)

	// 2. Find Chrome Manually
	chromePath := "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe"
	if _, err := os.Stat(chromePath); os.IsNotExist(err) {
		chromePath = "C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe"
	}

	// 3. Ensure the folder exists
	os.MkdirAll("C:\\selenium", 0755)

	// 4. LAUNCH SANDBOX (SILENT & MINIMIZED)
	cmd := exec.Command(chromePath,
		"--remote-debugging-port=9222",
		"--user-data-dir=C:\\selenium\\ChromeProfile",

		// --- WINDOW STATE ---
		"--start-minimized", // Starts in taskbar
		"--disable-session-crashed-bubble",

		// --- SILENCE FLAGS (Prevents Popups that steal focus) ---
		"--no-first-run",
		"--no-default-browser-check",
		"--disable-session-crashed-bubble", // Fixes "Restore Pages" popup
		"--disable-infobars",               // Hides "Chrome is being controlled"
		"--disable-restore-session-state",  // Prevents loading old tabs
		"--password-store=basic",           // Stops keyring prompts
		"--use-mock-keychain",
		// Add these to your command arguments:
		"--hide-crash-restore-bubble",
		"--disable-infobars",
		"--disable-background-networking",
		"--disable-sync",
		"--disable-translate",
	)

	// We removed the invalid syscall.StartupInfo block.
	// The flags above are sufficient to keep it minimized.

	cmd.Start()
}

// --- 2. STYLES ---
var (
	colRed    = lipgloss.Color("#D80000")
	colOrange = lipgloss.Color("#FF4500")
	colBlue   = lipgloss.Color("#00FFFF")
	colBlack  = lipgloss.Color("#080808")
	colGray   = lipgloss.Color("#333333")
	colGreen  = lipgloss.Color("#00FF00")
	colWhite  = lipgloss.Color("#FFFFFF") // New White Color

	txtRed    = lipgloss.NewStyle().Foreground(colRed)
	txtOrange = lipgloss.NewStyle().Foreground(colOrange)
	txtBlue   = lipgloss.NewStyle().Foreground(colBlue)
	txtGray   = lipgloss.NewStyle().Foreground(colGray)
	txtGreen  = lipgloss.NewStyle().Foreground(colGreen)
	txtWhite  = lipgloss.NewStyle().Foreground(colWhite) // New White Style

	bars = []string{" ", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
)

// --- 3. CUSTOM RENDERER ---
type item struct{ title, desc, id string }

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 3 }
func (d itemDelegate) Spacing() int                            { return 1 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	if index == m.Index() {
		fmt.Fprintf(w, " [+] %s\n     %s", txtRed.Bold(true).Render(strings.ToUpper(i.title)), txtOrange.Render(i.desc))
	} else {
		fmt.Fprintf(w, "    %s\n     %s", txtGray.Render(i.title), txtGray.Faint(true).Render(i.desc))
	}
}

// --- 4. STATE MODEL ---
type status int

const (
	statusBoot status = iota
	statusMenu
	statusGenesisInput
	statusGenesisThemeSelect
	statusGenesisCustomTheme
	statusGenesisPlatformSelect
	statusGenesisSocials
	statusAction
	statusProcessing
	statusMusic
	statusDone
	statusPostGen
	statusPatchMenu
	statusHostingerList
	statusManual
	statusAuditReport
)

type model struct {
	state     status
	list      list.Model
	textInput textinput.Model
	spinner   spinner.Model
	viewport  viewport.Model

	inputs         []string
	inputStep      int
	opMode         string
	selectedTarget string
	patchField     string
	lastGenFolder  string

	waveHistory     []int
	globalTick      int
	bootProgress    int
	diagRPM         int
	diagTemp        int
	statusMsg       string
	driverConnected bool
	blinkState      bool

	activePlayer *exec.Cmd
	musicPlaying bool
	currentSong  string

	width, height int
	auditResults  []string
}

// --- 5. INITIALIZATION ---
func initialModel() model {
	items := []list.Item{
		item{title: "PROJECT: AWAKENING", desc: "Genesis Mode (Create Site)", id: "GENESIS"},
		item{title: "PSYCHIC INJECTION", desc: "Smart Patch (Edit Site)", id: "PATCH"},
		item{title: "SATELLITE UPLINK", desc: "Hostinger Warp", id: "HOSTINGER"},
		item{title: "VISUAL CONFIRMATION", desc: "Preview Site", id: "PREVIEW"},
		item{title: "AUDIT: LINK HEALTH", desc: "Check all external dependencies", id: "AUDIT"},
		item{title: "JUKEBOX (SHOJI)", desc: "Audio Deck (.WAV ONLY)", id: "MUSIC"},
		item{title: "NETRUNNER'S CODEX", desc: "System Manual", id: "MANUAL"},
		item{title: "TERMINATE", desc: "Exit System", id: "EXIT"},
	}

	l := list.New(items, itemDelegate{}, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)

	ti := textinput.New()
	ti.Cursor.Style = txtRed
	ti.Prompt = " > "
	ti.TextStyle = txtBlue
	ti.Focus()

	s := spinner.New()
	s.Spinner = spinner.Line
	s.Style = txtRed

	vp := viewport.New(0, 0)
	vp.Style = lipgloss.NewStyle().PaddingLeft(2).PaddingRight(2)

	return model{
		state: statusBoot,
		list:  l, textInput: ti, spinner: s, viewport: vp,
		inputs:      make([]string, 0),
		waveHistory: make([]int, 20),
		diagRPM:     5000, diagTemp: 65,
		auditResults: make([]string, 0),
	}
}

func (m model) Init() tea.Cmd {
	go bootSequence()
	m.setManualText()
	return tea.Batch(tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg { return tickMsg(t) }), m.spinner.Tick)
}

// --- 6. UPDATE ---
type tickMsg time.Time
type folderMsg []string
type hostingerMsg []string
type browserMsg struct{ msg, folder string }
type auditMsg []string

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// --- UPDATE LIST SIZE ---
		m.list.SetWidth(msg.Width / 2)
		m.list.SetHeight(14)

		// --- UPDATE VIEWPORT SIZE ---
		// 1. Calculate the height of the static header
		headerHeight := lipgloss.Height(m.renderHeader())

		// 2. Define the space occupied by the Footer (2 lines) and Top Padding (2 lines)
		// Top Padding: "\n\n"
		// Footer: Text + "\n"
		verticalMargins := 3

		// 3. Calculate remaining space for the viewport content
		vpHeight := msg.Height - headerHeight - verticalMargins

		// 4. Safety check: prevent negative height if window is tiny
		if vpHeight < 1 {
			vpHeight = 1
		}

		m.viewport.Width = msg.Width
		m.viewport.Height = vpHeight

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			killPlayer(m.activePlayer)
			return m, tea.Quit
		}

		if msg.String() == "esc" && m.state != statusBoot {
			m.state = statusMenu
			m.resetMenu()
			return m, nil
		}

		if m.state == statusAuditReport {
			if msg.String() == "enter" || msg.String() == "esc" {
				m.state = statusMenu
				m.resetMenu()
				return m, nil
			}
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}

		if m.state == statusManual {
			if msg.String() == "enter" {
				m.state = statusMenu
				m.resetMenu()
				return m, nil
			}
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}

		// --- MAIN MENU ---
		if m.state == statusMenu {
			if msg.String() == "enter" {
				i, ok := m.list.SelectedItem().(item)
				if ok {
					switch i.id {
					case "GENESIS":
						m.state = statusGenesisInput
						m.opMode = "GENESIS"
						m.inputStep = 0
						m.inputs = []string{}
						m.textInput.Reset()
						m.textInput.Focus()
					case "PATCH", "PREVIEW", "AUDIT":
						m.opMode = i.id
						m.state = statusAction
						return m, loadFolders()
					case "HOSTINGER":
						m.state = statusProcessing
						m.statusMsg = "SCANNING HOSTINGER SATELLITES..."
						return m, runHostingerScan()
					case "MUSIC":
						m.state = statusMusic
						return m, loadMusic()
					case "MANUAL":
						m.state = statusManual
						m.setManualText()
						return m, nil
					case "EXIT":
						killPlayer(m.activePlayer)
						return m, tea.Quit
					}
				}
			}
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		// --- INPUT HANDLERS ---
		if m.state == statusGenesisInput {
			if msg.String() == "enter" {
				val := m.textInput.Value()
				if m.opMode == "PATCH" {
					executePatch(m.selectedTarget, m.patchField, val)
					m.state = statusDone
					m.statusMsg = fmt.Sprintf("PATCH APPLIED: %s -> %s", m.patchField, m.selectedTarget)
					return m, nil
				}
				if val == "" {
					val = "None"
				}
				m.inputs = append(m.inputs, val)
				m.inputStep++
				m.state = statusGenesisThemeSelect
				m.loadThemeOptions()
				return m, nil
			}
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}

		if m.state == statusGenesisThemeSelect {
			if msg.String() == "enter" {
				i, ok := m.list.SelectedItem().(item)
				if ok {
					if i.id == "CUSTOM" {
						m.state = statusGenesisCustomTheme
						m.textInput.Placeholder = "ENTER CUSTOM VISUAL STYLE..."
						m.textInput.Reset()
						return m, nil
					}
					m.inputs = append(m.inputs, i.title)
					m.inputStep++
					m.state = statusGenesisPlatformSelect
					m.loadPlatformOptions()
					return m, nil
				}
			}
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		if m.state == statusGenesisCustomTheme {
			if msg.String() == "enter" {
				val := m.textInput.Value()
				if val == "" {
					val = "Custom"
				}
				m.inputs = append(m.inputs, val)
				m.inputStep++
				m.state = statusGenesisPlatformSelect
				m.loadPlatformOptions()
				return m, nil
			}
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}

		if m.state == statusGenesisPlatformSelect {
			if msg.String() == "enter" {
				i, ok := m.list.SelectedItem().(item)
				if ok {
					m.inputs = append(m.inputs, i.title)
					m.inputStep++
					m.state = statusGenesisSocials
					m.textInput.Placeholder = "X / Twitter Link"
					m.textInput.Reset()
					return m, nil
				}
			}
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		if m.state == statusGenesisSocials {
			labels := []string{"TICKER SYMBOL", "THEME/VIBE", "BLOCKCHAIN PLATFORM", "X / Twitter Link", "Telegram Link", "Discord Link", "Dex Screener Link", "Tools/Utility Link", "Image Assets (URL)", "Extra Context"}
			if msg.String() == "enter" {
				val := m.textInput.Value()
				if val == "" {
					val = "None"
				}
				m.inputs = append(m.inputs, val)
				m.inputStep++
				m.textInput.Reset()
				nextStep := m.inputStep
				if nextStep < len(labels) {
					m.textInput.Placeholder = labels[nextStep]
				}
				if m.inputStep >= 10 {
					m.state = statusProcessing
					m.statusMsg = "TRANSMITTING DATA TO GEMINI..."
					return m, runGenesis(m.inputs)
				}
			}
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}

		// --- POST-GEN MENU ---
		if m.state == statusPostGen {
			if msg.String() == "enter" {
				i, ok := m.list.SelectedItem().(item)
				if ok {
					if i.id == "VIEW" {
						openBrowser(m.lastGenFolder)
					} else if i.id == "FOLDER" {
						openExplorer(m.lastGenFolder)
					} else if i.id == "MENU" {
						m.state = statusMenu
						m.resetMenu()
					}
				}
			}
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		// --- MODULES ---
		if m.state == statusAction {
			if msg.String() == "enter" {
				i, ok := m.list.SelectedItem().(item)
				if ok {
					if m.opMode == "PREVIEW" {
						openBrowser(i.title)
						m.state = statusDone
						m.statusMsg = "VISUALS ONLINE: " + i.title
					} else if m.opMode == "PATCH" {
						m.selectedTarget = i.title
						m.state = statusPatchMenu
						m.loadPatchOptions()
						return m, nil
					} else if m.opMode == "AUDIT" {
						m.selectedTarget = i.title
						m.state = statusProcessing
						m.statusMsg = fmt.Sprintf("AUDITING EXTERNAL LINKS FOR: %s", i.title)
						return m, runAudit(i.title)
					}
				}
			}
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		if m.state == statusPatchMenu {
			if msg.String() == "enter" {
				i, ok := m.list.SelectedItem().(item)
				if ok {
					m.patchField = i.id
					m.state = statusGenesisInput
					m.textInput.Placeholder = "Enter new value..."
					m.textInput.Reset()
					m.textInput.Focus()
					return m, nil
				}
			}
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		if m.state == statusHostingerList {
			if msg.String() == "enter" {
				i, ok := m.list.SelectedItem().(item)
				if ok {
					m.state = statusProcessing
					m.statusMsg = "WARPING TO DASHBOARD..."
					return m, runHostingerNav(i.title)
				}
			}
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		if m.state == statusMusic {
			if msg.String() == "enter" {
				i, ok := m.list.SelectedItem().(item)
				if ok {
					killPlayer(m.activePlayer)
					if i.id != "STOP" {
						m.activePlayer = playAudio(i.title)
						m.musicPlaying = true
						m.currentSong = i.title
					} else {
						m.musicPlaying = false
					}
				}
			}
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		if m.state == statusDone {
			if msg.String() == "enter" {
				m.state = statusMenu
				m.resetMenu()
			}
		}

	case tickMsg:
		m.globalTick++
		if m.state == statusBoot {
			m.bootProgress++
			if m.bootProgress > 25 {
				m.state = statusMenu
			}
		}
		if m.globalTick%5 == 0 {
			if checkDriverPort() {
				m.driverConnected = true
			} else {
				m.driverConnected = false
			}
		}
		if m.globalTick%10 == 0 {
			m.blinkState = !m.blinkState
			m.waveHistory = append(m.waveHistory[1:], rand.Intn(8))
			if rand.Intn(100) > 80 {
				m.diagRPM = 8000 + rand.Intn(6000)
				m.diagTemp = 80 + rand.Intn(40)
			}
		}

		// --- ANIMATE AUDIT REPORT ICONS ---
		if m.state == statusAuditReport {
			iconOk := "●"
			iconFail := "■"
			iconWarn := "▲"
			if !m.blinkState {
				iconOk = "◌"
				iconFail = " "
				iconWarn = " "
			}

			// Add space between Header and Title as requested
			report := "\n\n   AUDIT REPORT: EXTERNAL DEPENDENCY HEALTH\n   ----------------------------------------\n"
			for _, result := range m.auditResults {
				line := strings.ReplaceAll(result, "::OK::", iconOk)
				line = strings.ReplaceAll(line, "::FAIL::", iconFail)
				line = strings.ReplaceAll(line, "::WARN::", iconWarn)
				report += "   " + line + "\n"
			}
			m.viewport.SetContent(report)
		}

		return m, tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg { return tickMsg(t) })

	case folderMsg:
		var items []list.Item
		if m.state == statusMusic {
			items = append(items, item{title: "SILENCE (STOP)", desc: "Stop Audio Signal", id: "STOP"})
		}
		sort.Strings(msg)
		for _, f := range msg {
			items = append(items, item{title: f, desc: "DATA", id: f})
		}
		if len(items) == 0 {
			items = append(items, item{title: "NO DATA FOUND", desc: "Create a project first", id: "NONE"})
		}
		m.list.SetItems(items)

	case hostingerMsg:
		m.state = statusHostingerList
		var items []list.Item
		for _, s := range msg {
			items = append(items, item{title: s, desc: "Active Site", id: s})
		}
		m.list.SetItems(items)

	case auditMsg:
		m.auditResults = msg
		m.state = statusAuditReport
		m.list.SetItems(make([]list.Item, 0))

	case browserMsg:
		if strings.Contains(msg.msg, "SUBJECT CONSTRUCTED") {
			m.state = statusPostGen
			m.lastGenFolder = msg.folder
			options := []list.Item{
				item{title: "VISUAL CONFIRMATION", desc: "Open Site in Browser", id: "VIEW"},
				item{title: "ACCESS DATA", desc: "Open Windows Folder", id: "FOLDER"},
				item{title: "NEW OPERATION", desc: "Return to Main Menu", id: "MENU"},
			}
			m.list.SetItems(options)
			m.list.ResetSelected()
		} else {
			m.state = statusDone
			m.statusMsg = msg.msg
		}

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, cmd
}

// --- 7. VIEW ---
func (m model) renderHeader() string {
	graph := ""
	for _, h := range m.waveHistory {
		graph += bars[h]
	}

	titleBlock := txtRed.Render(fmt.Sprintf(`
    ▄████▄ ██ ▄█▀ ██ █████▄  ▄████▄ ©
    ██▄▄██ ████   ██ ██▄▄██▄ ██▄▄██ 
    ██  ██ ██ ▀█▄ ██ ██   ██ ██  ██ 											[+] 1988.
                                 
  %s // %s`, APP_AUTHOR, APP_VERSION))
	return "\n" + titleBlock
}

// --- NEW HELPER ---
func (m model) renderMusicStatus() string {
	musicStatus := txtGray.Render("AUDIO: STANDBY")
	if m.musicPlaying {
		musicStatus = txtBlue.Render("PLAYING: " + m.currentSong)
	}
	return musicStatus
}

func (m model) View() string {

	if m.state == statusBoot {
		bar := strings.Repeat("█", m.bootProgress)
		return txtOrange.Render(fmt.Sprintf("\n\n  INITIALIZING NEO-TOKYO KERNEL...\n  [%-25s]\n\n  > MOUNTING CHROME DRIVER (SELENIUM MODE)...\n  > LOADING AUDIO DRIVER (.WAV)...", bar))
	}

	header := m.renderHeader()

	// --- FIX: DEFINE CONTENT STYLE HERE FOR ALL BLOCKS TO USE ---
	contentStyle := lipgloss.NewStyle().PaddingLeft(3)

	// LIST SCREENS
	if m.state == statusMenu || m.state == statusAction || m.state == statusMusic || m.state == statusPatchMenu || m.state == statusHostingerList || m.state == statusGenesisThemeSelect || m.state == statusGenesisPlatformSelect || m.state == statusPostGen {

		driverStatus := ""
		if m.driverConnected {
			indicator := "●"
			if !m.blinkState {
				indicator = "◌"
			}
			driverStatus = txtGreen.Render(fmt.Sprintf("%s DRIVER: ONLINE", indicator))
		} else {
			indicator := "■"
			if !m.blinkState {
				indicator = " "
			}
			driverStatus = txtRed.Render(fmt.Sprintf("%s DRIVER: OFFLINE", indicator))
		}

		musicStatus := m.renderMusicStatus()

		graph := ""
		for _, h := range m.waveHistory {
			graph += bars[h]
		}
		diag := fmt.Sprintf("\n  DIAGNOSTICS\n  -----------\n  CPU: %d°C\n  RPM: %d\n\n  %s\n\n  %s\n  %s",
			m.diagTemp, m.diagRPM, txtBlue.Render(graph), driverStatus, musicStatus)

		listHeader := ""
		if m.state == statusPatchMenu {
			listHeader = txtOrange.Render(fmt.Sprintf(" INJECTING INTO: %s", strings.ToUpper(m.selectedTarget))) + "\n"
		}

		if m.state == statusHostingerList {
			listHeader = txtOrange.Render(" SELECT HOSTINGER NODE:") + "\n"
		}
		if m.state == statusGenesisThemeSelect {
			listHeader = txtOrange.Render(" SELECT VISUAL STYLE:") + "\n"
		}
		if m.state == statusGenesisPlatformSelect {
			listHeader = txtOrange.Render(" SELECT BLOCKCHAIN NETWORK:") + "\n"
		}
		if m.state == statusPostGen {
			listHeader = txtGreen.Render(fmt.Sprintf("  SUCCESS: %s", m.lastGenFolder)) + "\n"
		}

		// 1. Create the main body content (List + Diagnostics)
		body := lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(60).Render(listHeader+m.list.View()),
			txtOrange.Render(diag))

		// 2. Logic for App Description (MAIN DASHBOARD ONLY)
		finalHeader := header
		if m.state == statusMenu {
			// Add a newline, then the description in Blue
			finalHeader = lipgloss.JoinVertical(lipgloss.Left, header, txtRed.Render("  "+APP_DESC))
		}

		// 2. Apply the PaddingLeft(4) to match the Genesis Input screen
		contentStyle := lipgloss.NewStyle().PaddingLeft(3)

		return lipgloss.JoinVertical(lipgloss.Left, finalHeader, "\n\n", contentStyle.Render(body))
	}

	// --- AUDIT REPORT VIEW ---
	if m.state == statusAuditReport {
		// 1. Define Footer (Text + Newline for bottom padding)
		footerText := fmt.Sprintf("   [PRESS ESC or ENTER TO RETURN]                       %s\n", m.renderMusicStatus())
		footer := txtWhite.Render(footerText)

		// 2. Render Viewport (Size is already calculated in Update)
		screenContent := m.viewport.View()

		// 3. Stack them: Top Padding + Header + Content + Footer
		return lipgloss.JoinVertical(
			lipgloss.Left,
			header,        // The Header
			screenContent, // The Scrollable Area
			footer,        // The Footer (2 lines)
		)
	}

	// MANUAL SCREEN (SCROLLABLE VIEWPORT)
	if m.state == statusManual {
		return lipgloss.JoinVertical(lipgloss.Left, header, "\n", m.viewport.View())
	}

	// VIEW: GENESIS/PATCH INPUT SCREENS (Genesis & Smart Patch)
	if m.state == statusGenesisInput || m.state == statusGenesisSocials || m.state == statusGenesisCustomTheme {

		labels := []string{"TICKER SYMBOL", "THEME/VIBE", "BLOCKCHAIN PLATFORM", "TWITTER LINK", "TELEGRAM LINK", "DISCORD", "DEX SCREENER", "TOOLS/UTILITY", "IMAGE ASSETS (URL)", "EXTRA CONTEXT"}

		history := ""
		for i, input := range m.inputs {
			if i < len(labels) {
				history += txtGray.Faint(true).Render(fmt.Sprintf("  ▵ %s: %s\n", labels[i], input))
			}
		}

		lbl := "INPUT VALUE"
		if m.state == statusGenesisCustomTheme {
			lbl = "CUSTOM VISUAL STYLE"
		} else if m.opMode == "GENESIS" && m.inputStep < len(labels) {
			lbl = labels[m.inputStep]
		} else if m.opMode == "PATCH" {
			lbl = fmt.Sprintf("NEW VALUE FOR %s", m.patchField)
		}

		rawContent := fmt.Sprintf(`
SUBJECT INITIALIZATION PROTOCOL
-------------------------------
%s
ENTER PARAMETER (%s):
%s
`, history, txtOrange.Render(lbl), m.textInput.View())

		// 2. Apply Indentation using Lipgloss (Fixes the layout bug)
		contentStyle := lipgloss.NewStyle().PaddingLeft(4)
		screenContent := contentStyle.Render(rawContent)

		// --- STICKY FOOTER LOGIC ---
		footerText := fmt.Sprintf("   [ESC] ABORT      [ENTER] INITIATE SEQUENCE                  %s", m.renderMusicStatus())
		footer := txtWhite.Render(footerText)

		// 3. Calculate Height using the STYLED content
		headerH := lipgloss.Height(header)
		contentH := lipgloss.Height(screenContent)
		footerH := lipgloss.Height(footer)

		// Subtract 2 for the header's top spacing + spacers
		gap := m.height - headerH - contentH - footerH - 3

		gapFill := ""
		if gap > 0 {
			gapFill = strings.Repeat("\n", gap)
		}

		return lipgloss.JoinVertical(lipgloss.Left, "\n\n\n", header, screenContent, gapFill, footer, "\n")
	}

	// VIEW: PROCESSING SCREEN
	if m.state == statusProcessing {
		flux := strings.Repeat("█", rand.Intn(15)) + strings.Repeat("▒", 5)

		// 1. Clean Raw Content (Removed manual indentation)
		rawContent := fmt.Sprintf(`
⚠  CRITICAL ALERT  ⚠  //  SYNCHRONIZING...

%s %s

DATA FLUX:
%s  [OVERFLOW]

AWAITING NEURAL LINK RESPONSE...
`, m.spinner.View(), m.statusMsg, flux)

		// 2. Apply Formatting
		// Color it Orange first, then apply the Left Padding style
		screenContent := contentStyle.Render(txtOrange.Render(rawContent))

		// 4. Calculate Height & Gap
		headerH := lipgloss.Height(header)
		contentH := lipgloss.Height(screenContent)

		// Subtract 2 for the top spacing
		gap := m.height - headerH - contentH - 2

		gapFill := ""
		if gap > 0 {
			gapFill = strings.Repeat("\n", gap)
		}

		return lipgloss.JoinVertical(lipgloss.Left, header, screenContent, gapFill)
	}

	if m.state == statusDone {
		// Apply Padding
		return lipgloss.JoinVertical(lipgloss.Left, header, contentStyle.Render(txtBlue.Render(fmt.Sprintf("\n  MISSION REPORT:\n  %s\n\n  [PRESS ENTER TO RETURN]", m.statusMsg))))
	}

	return ""
}

// --- 8. LOGIC HELPERS ---

func (m *model) setManualText() {
	manual := `
 NETRUNNER'S CODEX (MANUAL)
 --------------------------

 1. GENESIS MODE
    Constructs full websites. You will be prompted for Ticker, Theme, etc.
    
 2. SMART PATCH
    Edit existing sites. Select a folder, then choose what to inject.

 3. HOSTINGER WARP
    Scans your Hostinger account.
    NOTE: You MUST log in to Hostinger in the minimized Chrome window!

 4. DRIVER STATUS
    The app uses "C:\selenium\ChromeProfile".
    If driver is offline, maximize the Chrome window and ensure it didn't crash.

 5. AUDIO PROTOCOL
    Place .WAV files in the /music folder.
    
 6. AUDIT MODULE (NEW)
     Checks all external links in index.html (socials, scripts, etc.) for a 200 OK status.

 [ UP / DOWN / MOUSE ] SCROLL     [ ENTER ] RETURN
`
	m.viewport.SetContent(txtBlue.Render(manual))
}

func checkDriverPort() bool {
	conn, err := net.DialTimeout("tcp", "127.0.0.1:9222", time.Second)
	if err == nil {
		conn.Close()
		return true
	}
	return false
}

func loadFolders() tea.Cmd {
	return func() tea.Msg {
		pwd, _ := os.Getwd()
		path := filepath.Join(pwd, "websites")
		os.MkdirAll(path, 0755)
		entries, _ := os.ReadDir(path)
		var folders []string
		for _, e := range entries {
			if e.IsDir() {
				folders = append(folders, e.Name())
			}
		}
		return folderMsg(folders)
	}
}

func (m *model) loadThemeOptions() {
	items := []list.Item{
		item{title: "CYBERPUNK", desc: "High contrast, neon, glitch", id: "CYBERPUNK"},
		item{title: "RETROWAVE", desc: "Sunsets, grids, 80s vibe", id: "RETROWAVE"},
		item{title: "DEGEN / MEME", desc: "Chaotic, comic sans, vibrant", id: "DEGEN"},
		item{title: "DARK MATTER", desc: "Deep black, subtle purple", id: "DARK"},
		item{title: "CUSTOM", desc: "Type your own style...", id: "CUSTOM"},
	}
	m.list.SetItems(items)
}

func (m *model) loadPlatformOptions() {
	items := []list.Item{
		item{title: "SOLANA", desc: "SPL Standard", id: "SOL"},
		item{title: "PUMP.FUN", desc: "Pump Bonding Curve", id: "PUMP"},
		item{title: "ETHEREUM", desc: "ERC-20 Standard", id: "ETH"},
		item{title: "BSC", desc: "Binance Smart Chain", id: "BSC"},
		item{title: "BASE", desc: "Coinbase L2", id: "BASE"},
		item{title: "MONAD", desc: "Next Gen L1", id: "MONAD"},
	}
	m.list.SetItems(items)
}

func (m *model) loadPatchOptions() {
	items := []list.Item{
		item{title: "Inject CA", desc: "Update CA_PLACEHOLDER", id: "Inject CA"},
		item{title: "Update X", desc: "Update X_LINK_PLACEHOLDER", id: "Update X"},
		item{title: "Update Telegram", desc: "Update TG_LINK_PLACEHOLDER", id: "Update Telegram"},
		item{title: "Update Discord", desc: "Update DISCORD_PLACEHOLDER", id: "Update Discord"},
		item{title: "Update DexScreener", desc: "Update DEX_LINK_PLACEHOLDER", id: "Update DexScreener"},
	}
	m.list.SetItems(items)
}

func loadMusic() tea.Cmd {
	return func() tea.Msg {
		pwd, _ := os.Getwd()
		path := filepath.Join(pwd, "music")
		os.MkdirAll(path, 0755)
		entries, _ := os.ReadDir(path)
		var files []string
		for _, e := range entries {
			if strings.HasSuffix(strings.ToLower(e.Name()), ".wav") {
				files = append(files, e.Name())
			}
		}
		if len(files) == 0 {
			files = append(files, "NO .WAV FILES FOUND")
		}
		return folderMsg(files)
	}
}

func killPlayer(cmd *exec.Cmd) {
	if cmd != nil && cmd.Process != nil {
		cmd.Process.Kill()
	}
}

func playAudio(filename string) *exec.Cmd {
	pwd, _ := os.Getwd()
	fullPath := filepath.Join(pwd, "music", filename)
	safePath := strings.ReplaceAll(fullPath, "'", "''")
	psScript := fmt.Sprintf(`$player = New-Object System.Media.SoundPlayer; $player.SoundLocation = '%s'; $player.PlayLooping(); while($true) { Start-Sleep 1 }`, safePath)
	cmd := exec.Command("powershell", "-c", psScript)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Start()
	return cmd
}

func openBrowser(folder string) {
	pwd, _ := os.Getwd()
	path := filepath.Join(pwd, "websites", folder, "index.html")
	exec.Command("cmd", "/c", "start", "", path).Start()
}

func openExplorer(folder string) {
	pwd, _ := os.Getwd()
	path := filepath.Join(pwd, "websites", folder)
	exec.Command("explorer", path).Start()
}

func runHostingerScan() tea.Cmd {
	return func() tea.Msg {
		defer func() { recover() }()
		u := launcher.MustResolveURL("127.0.0.1:9222")
		browser := rod.New().ControlURL(u).MustConnect()
		page := browser.MustPage("https://hpanel.hostinger.com/websites")
		page.MustWaitLoad()
		var sites []string
		page.Race().Element(".website-card").Handle(func(e *rod.Element) error {
			els := page.MustElements(".website-card")
			for _, el := range els {
				txt, _ := el.Text()
				sites = append(sites, strings.Split(txt, "\n")[0])
			}
			return nil
		}).Element("body").MustDo()
		if len(sites) == 0 {
			sites = append(sites, "Manual Dashboard (Scan Failed)")
		}
		return hostingerMsg(sites)
	}
}

func runHostingerNav(domain string) tea.Cmd {
	return func() tea.Msg {
		defer func() { recover() }()
		u := launcher.MustResolveURL("127.0.0.1:9222")
		browser := rod.New().ControlURL(u).MustConnect()
		var page *rod.Page
		if domain == "Manual Dashboard (Scan Failed)" || domain == "" {
			page = browser.MustPage("https://hpanel.hostinger.com/")
		} else {
			page = browser.MustPage(fmt.Sprintf("https://hpanel.hostinger.com/hosting/%s", domain))
		}
		page.MustWaitLoad()
		page.MustActivate()
		return browserMsg{msg: "UPLINK SECURE. DASHBOARD ACTIVE."}
	}
}

func executePatch(folder, action, value string) {
	pwd, _ := os.Getwd()
	path := filepath.Join(pwd, "websites", folder, "index.html")
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}
	text := string(content)
	switch action {
	case "Inject CA":
		text = strings.ReplaceAll(text, "CA_PLACEHOLDER", value)
	case "Update X":
		text = strings.ReplaceAll(text, "X_LINK_PLACEHOLDER", value)
	case "Update Telegram":
		text = strings.ReplaceAll(text, "TG_LINK_PLACEHOLDER", value)
	case "Update Discord":
		text = strings.ReplaceAll(text, "DISCORD_PLACEHOLDER", value)
	case "Update DexScreener":
		text = strings.ReplaceAll(text, "DEX_LINK_PLACEHOLDER", value)
	}
	os.WriteFile(path, []byte(text), 0644)
}

func parseAssetPrompts(htmlContent string, fullPath string) {
	startTag := "### 2. THE ASSET VAULT (DYNAMIC LOGOS)"
	endTag := "#### STEP 3: 💻 THE CODE"

	start := strings.Index(htmlContent, startTag)
	if start == -1 {
		return
	}

	contentAfterStart := htmlContent[start:]
	end := strings.Index(contentAfterStart, endTag)

	promptContent := ""
	if end != -1 {
		promptContent = contentAfterStart[:end]
	} else {
		promptContent = contentAfterStart
	}

	cleanedContent := strings.TrimSpace(promptContent)
	promptPath := filepath.Join(fullPath, "asset_prompts.txt")
	os.WriteFile(promptPath, []byte(cleanedContent), 0644)
}

func runAudit(folder string) tea.Cmd {
	return func() tea.Msg {
		pwd, _ := os.Getwd()
		path := filepath.Join(pwd, "websites", folder, "index.html")
		content, err := os.ReadFile(path)
		if err != nil {
			return auditMsg{fmt.Sprintf("FATAL: Could not read index.html: %v", err)}
		}
		text := string(content)

		re := regexp.MustCompile(`(href|src)="([^"]*http[^"]*)"`)
		matches := re.FindAllStringSubmatch(text, -1)

		var urls []string
		for _, match := range matches {
			if len(match) > 2 {
				urls = append(urls, match[2])
			}
		}

		uniqueURLs := make(map[string]struct{})
		for _, url := range urls {
			uniqueURLs[url] = struct{}{}
		}

		results := []string{}
		client := http.Client{Timeout: 8 * time.Second}

		for url := range uniqueURLs {
			if strings.Contains(url, "PLACEHOLDER") {
				line := fmt.Sprintf("ℹ️ INFO     %-60s [PLACEHOLDER]", url)
				results = append(results, txtBlue.Render(line))
				continue
			}

			if strings.HasPrefix(url, "http") {
				req, err := http.NewRequest("GET", url, nil)
				if err != nil {
					line := fmt.Sprintf("::FAIL:: FAIL     %s - Request Error", url)
					results = append(results, txtRed.Render(line))
					continue
				}

				resp, err := client.Do(req)
				if err != nil {
					line := fmt.Sprintf("::FAIL:: FAIL     %s - Connection Error", url)
					results = append(results, txtRed.Render(line))
					continue
				}
				defer resp.Body.Close()

				statusText := fmt.Sprintf("[STATUS %d]", resp.StatusCode)

				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					line := fmt.Sprintf("::OK:: ONLINE   %-60s %s", url, statusText)
					results = append(results, txtGreen.Render(line))
				} else if resp.StatusCode >= 300 && resp.StatusCode < 400 {
					line := fmt.Sprintf("::WARN:: WARNING %-60s %s", url, statusText)
					results = append(results, txtOrange.Render(line))
				} else {
					line := fmt.Sprintf("::FAIL:: FAIL     %-60s %s", url, statusText)
					results = append(results, txtRed.Render(line))
				}
			}
		}

		if len(results) == 0 {
			results = append(results, txtBlue.Render("No external HTTP links detected in index.html."))
		}

		return auditMsg(results)
	}
}

func runGenesis(inputs []string) tea.Cmd {
	return func() tea.Msg {
		defer func() { recover() }()
		u := launcher.MustResolveURL("127.0.0.1:9222")
		browser := rod.New().ControlURL(u).MustConnect()
		page := browser.MustPage("https://gemini.google.com/app")
		page.MustWaitLoad()

		time.Sleep(1 * time.Second)

		inputBox, err := page.Element("rich-textarea")
		if err != nil {
			inputBox = page.MustElement("div[contenteditable='true']")
		}
		inputBox.MustClick()
		time.Sleep(500 * time.Millisecond)

		data := fmt.Sprintf("\nDATA: TICKER:%s THEME:%s PLATFORM:%s LINKS:%s %s ASSETS:%s", inputs[0], inputs[1], inputs[2], inputs[3], inputs[4], inputs[8])
		inputBox.MustInput(SYSTEM_PROMPT + data)

		time.Sleep(2 * time.Second)

		sendBtn, err := page.Element("button[aria-label*='Send']")
		if err != nil {
			sendBtn = page.MustElement(".send-button")
		}
		sendBtn.MustClick()

		page.MustWaitIdle()

		foundCode := false
		var el *rod.Element
		stableCount := 0
		lastLen := 0

		for i := 0; i < 120; i++ {
			if has, e, _ := page.Has("pre code"); has {
				el = e
				txt, _ := el.Text()
				currLen := len(txt)

				if currLen > 0 && currLen == lastLen {
					stableCount++
					if stableCount >= 3 {
						foundCode = true
						break
					}
				} else {
					stableCount = 0
				}
				lastLen = currLen
			}
			time.Sleep(1 * time.Second)
		}

		if !foundCode {
			return browserMsg{msg: "ERR: GEMINI TIMEOUT (NO CODE GENERATED)"}
		}

		htmlContent := ""
		if el != nil {
			htmlContent, _ = el.Text()
		}

		cleanTicker := strings.ToUpper(strings.ReplaceAll(inputs[0], "$", ""))
		dateStr := time.Now().Format("2006-01-02")
		folderName := fmt.Sprintf("%s_%s", cleanTicker, dateStr)

		pwd, _ := os.Getwd()
		fullPath := filepath.Join(pwd, "websites", folderName)

		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return browserMsg{msg: fmt.Sprintf("FATAL: Cannot create project folder at %s. Error: %v", fullPath, err)}
		}

		assetBase := filepath.Join(fullPath, "images")
		subfolders := []string{"socials", "emojies", "stickers"}
		failedFolders := []string{}
		firstAssetPath := ""

		for i, sf := range subfolders {
			assetPath := filepath.Join(assetBase, sf)
			if err := os.MkdirAll(assetPath, 0755); err != nil {
				failedFolders = append(failedFolders, fmt.Sprintf("%s (Error: %v)", assetPath, err))
			}
			if i == 0 {
				firstAssetPath = assetPath
			}
		}

		if len(failedFolders) > 0 {
			return browserMsg{msg: fmt.Sprintf("CRITICAL WARNING: Asset folder creation failed. Errors: %s", strings.Join(failedFolders, "; "))}
		}

		if _, err := os.Stat(firstAssetPath); os.IsNotExist(err) {
			return browserMsg{msg: fmt.Sprintf("FATAL: Asset folders reported success, but verification failed. Check permissions for %s", firstAssetPath)}
		}

		if htmlContent != "" {
			parseAssetPrompts(htmlContent, fullPath)

			if err := os.WriteFile(filepath.Join(fullPath, "index.html"), []byte(htmlContent), 0644); err != nil {
				return browserMsg{msg: fmt.Sprintf("FATAL: Could not write index.html. Error: %v", err)}
			}
			return browserMsg{msg: "SUBJECT CONSTRUCTED: Asset folders created successfully.", folder: folderName}
		}
		return browserMsg{msg: "ERR: NO CODE DETECTED IN RESPONSE."}
	}
}

func (m *model) resetMenu() {
	m.list.ResetSelected()
	m.list.SetItems(initialModel().list.Items())
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("ERR: %v", err)
		os.Exit(1)
	}
}
