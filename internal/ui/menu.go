package ui

import (
	"encoding/csv"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	cliOrange = lipgloss.Color("#FF8000")
	cliGreen  = lipgloss.Color("#00E054")
	cliBlue   = lipgloss.Color("#40B0FF")
	dimGrey   = lipgloss.Color("240")
	midGrey   = lipgloss.Color("244")
	lightGrey = lipgloss.Color("248")

	menuLogoStyle = lipgloss.NewStyle().
			MarginBottom(1)
	headerSubtitleStyle = lipgloss.NewStyle().
				Foreground(midGrey).
				SetString("A beautiful terminal interface for Letterboxd")

	gradientStart = lipgloss.NewStyle().Foreground(cliOrange)
	gradientMid   = lipgloss.NewStyle().Foreground(cliGreen)
	gradientEnd   = lipgloss.NewStyle().Foreground(cliBlue)

	panelTitleStyle = lipgloss.NewStyle().
			Foreground(midGrey).
			Bold(true).
			Margin(1, 0, 1, 0)

	navItemStyle = lipgloss.NewStyle().
			Foreground(midGrey).
			PaddingLeft(2)
	navItemSelectedStyle = lipgloss.NewStyle().
				Foreground(lightGrey).
				Bold(true).
				PaddingLeft(2)
	navCircleStyle = func(color lipgloss.Color, number string) lipgloss.Style {
		return lipgloss.NewStyle().
			Foreground(color).
			SetString(fmt.Sprintf("%s %s", number, "●"))
	}

	menuItemSelectedBgStyle = lipgloss.NewStyle().Background(lipgloss.Color("236"))

	colorBlockStyle    = lipgloss.NewStyle().Width(18).Height(1)
	colorHexStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	colorNameStyle     = lipgloss.NewStyle().Width(16).Foreground(dimGrey)
	colorBlockBoxStyle = lipgloss.NewStyle().MarginBottom(1)

	quoteTextStyle = lipgloss.NewStyle().
			Foreground(lightGrey).
			Italic(true).
			Bold(true).
			MarginLeft(2)
	quoteAuthorStyle = lipgloss.NewStyle().
				Foreground(cliOrange).
				Bold(true).
				Margin(0, 0, 0, 4)

	tipBulletStyle  = lipgloss.NewStyle().Foreground(cliGreen).SetString("•")
	tipTextStyle    = lipgloss.NewStyle().Foreground(midGrey)
	footerTextStyle = lipgloss.NewStyle().Foreground(dimGrey).MarginTop(1).Italic(true)

	menuQuitTextStyle = lipgloss.NewStyle().Margin(1, 0, 2, 4)
	appLayoutStyle    = lipgloss.NewStyle().Margin(1, 2)
	menuShortcutStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	menuBorderBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(dimGrey).
			Padding(1, 2)
	menuHelpTitleStyle = lipgloss.NewStyle().
				Foreground(cliGreen).
				Bold(true).
				MarginBottom(1)
	menuHelpKeyStyle = lipgloss.NewStyle().
				Foreground(cliOrange).
				Width(15)
	menuHelpDescStyle = lipgloss.NewStyle().
				Foreground(midGrey)
)

func renderGradientLine(width int) string {
	start := "─"
	mid := "─"
	end := "─"
	oneThird := width / 3
	twoThird := width * 2 / 3
	line := gradientStart.Render(strings.Repeat(start, oneThird)) +
		gradientMid.Render(strings.Repeat(mid, twoThird-oneThird)) +
		gradientEnd.Render(strings.Repeat(end, width-twoThird))
	return line
}

var menuLogo = `
		[38;2;102;187;255m░[0m[38;2;102;187;253m█[0m[38;2;103;188;252m█[0m[38;2;103;189;251m█[0m[38;2;104;189;250m█[0m                     [38;2;116;205;226m░[0m[38;2;116;206;225m█[0m[38;2;117;207;224m█[0m[38;2;117;207;223m█[0m[38;2;118;208;222m█[0m      [38;2;122;213;214m░[0m[38;2;122;214;213m█[0m[38;2;123;215;212m█[0m[38;2;123;215;211m█[0m[38;2;124;216;210m█[0m                            [38;2;139;237;179m░[0m[38;2;140;238;177m█[0m[38;2;141;239;176m█[0m[38;2;141;239;175m█[0m[38;2;142;240;174m█[0m[38;2;142;241;173m█[0m[38;2;143;241;172m█[0m   [38;2;145;244;168m░[0m[38;2;145;245;167m█[0m[38;2;146;246;166m█[0m[38;2;147;247;164m█[0m[38;2;147;247;163m█[0m   [38;2;149;250;159m░[0m[38;2;150;251;158m█[0m[38;2;150;252;157m█[0m[38;2;151;252;156m█[0m[38;2;151;253;155m█[0m 
		[38;2;102;187;255m░[0m[38;2;102;187;253m█[0m[38;2;103;188;252m█[0m[38;2;103;189;251m█[0m[38;2;104;189;250m█[0m                     [38;2;116;205;226m░[0m[38;2;116;206;225m█[0m[38;2;117;207;224m█[0m[38;2;117;207;223m█[0m[38;2;118;208;222m█[0m      [38;2;122;213;214m░[0m[38;2;122;214;213m█[0m[38;2;123;215;212m█[0m[38;2;123;215;211m█[0m[38;2;124;216;210m█[0m                          [38;2;138;236;181m░[0m[38;2;139;236;180m█[0m[38;2;139;237;179m█[0m[38;2;140;238;177m█[0m[38;2;141;239;176m█[0m [38;2;142;240;174m░[0m[38;2;142;241;173m█[0m[38;2;143;241;172m█[0m[38;2;143;242;171m█[0m[38;2;144;243;170m█[0m [38;2;145;244;168m░[0m[38;2;145;245;167m█[0m[38;2;146;246;166m█[0m[38;2;147;247;164m█[0m[38;2;147;247;163m█[0m         
		[38;2;102;187;255m░[0m[38;2;102;187;253m█[0m[38;2;103;188;252m█[0m[38;2;103;189;251m█[0m[38;2;104;189;250m█[0m         [38;2;109;197;239m░[0m[38;2;110;197;238m█[0m[38;2;110;198;237m█[0m[38;2;111;199;236m█[0m[38;2;111;200;235m█[0m[38;2;112;200;234m█[0m[38;2;112;201;233m█[0m   [38;2;115;204;228m░[0m[38;2;115;205;227m█[0m[38;2;116;205;226m█[0m[38;2;116;206;225m█[0m[38;2;117;207;224m█[0m[38;2;117;207;223m█[0m[38;2;118;208;222m█[0m[38;2;118;209;221m█[0m[38;2;119;210;220m█[0m  [38;2;120;212;217m░[0m[38;2;121;213;215m█[0m[38;2;122;213;214m█[0m[38;2;122;214;213m█[0m[38;2;123;215;212m█[0m[38;2;123;215;211m█[0m[38;2;124;216;210m█[0m[38;2;124;217;209m█[0m[38;2;125;218;208m█[0m  [38;2;126;220;205m░[0m[38;2;127;221;204m█[0m[38;2;128;221;202m█[0m[38;2;128;222;201m█[0m[38;2;129;223;200m█[0m[38;2;129;223;199m█[0m[38;2;130;224;198m█[0m   [38;2;132;227;194m░[0m[38;2;132;228;193m█[0m[38;2;133;228;192m█[0m[38;2;134;229;190m█[0m[38;2;134;230;189m█[0m[38;2;135;231;188m█[0m[38;2;135;231;187m█[0m[38;2;136;232;186m█[0m[38;2;136;233;185m█[0m   [38;2;138;236;181m░[0m[38;2;139;236;180m█[0m[38;2;139;237;179m█[0m[38;2;140;238;177m█[0m[38;2;141;239;176m█[0m       [38;2;145;244;168m░[0m[38;2;145;245;167m█[0m[38;2;146;246;166m█[0m[38;2;147;247;164m█[0m[38;2;147;247;163m█[0m [38;2;148;249;161m░[0m[38;2;149;249;160m█[0m[38;2;149;250;159m█[0m[38;2;150;251;158m█[0m[38;2;150;252;157m█[0m[38;2;151;252;156m█[0m[38;2;151;253;155m█[0m 
		[38;2;102;187;255m░[0m[38;2;102;187;253m█[0m[38;2;103;188;252m█[0m[38;2;103;189;251m█[0m[38;2;104;189;250m█[0m       [38;2;108;195;241m░[0m[38;2;109;196;240m█[0m[38;2;109;197;239m█[0m[38;2;110;197;238m█[0m[38;2;110;198;237m█[0m [38;2;111;200;235m░[0m[38;2;112;200;234m█[0m[38;2;112;201;233m█[0m[38;2;113;202;232m█[0m[38;2;113;202;231m█[0m   [38;2;116;205;226m░[0m[38;2;116;206;225m█[0m[38;2;117;207;224m█[0m[38;2;117;207;223m█[0m[38;2;118;208;222m█[0m      [38;2;122;213;214m░[0m[38;2;122;214;213m█[0m[38;2;123;215;212m█[0m[38;2;123;215;211m█[0m[38;2;124;216;210m█[0m  [38;2;125;218;207m░[0m[38;2;126;219;206m█[0m[38;2;126;220;205m█[0m[38;2;127;221;204m█[0m[38;2;128;221;202m█[0m [38;2;129;223;200m░[0m[38;2;129;223;199m█[0m[38;2;130;224;198m█[0m[38;2;130;225;197m█[0m[38;2;131;226;196m█[0m [38;2;132;227;194m░[0m[38;2;132;228;193m█[0m[38;2;133;228;192m█[0m[38;2;134;229;190m█[0m[38;2;134;230;189m█[0m [38;2;135;231;187m░[0m[38;2;136;232;186m█[0m[38;2;136;233;185m█[0m[38;2;137;234;184m█[0m[38;2;137;234;183m█[0m [38;2;138;236;181m░[0m[38;2;139;236;180m█[0m[38;2;139;237;179m█[0m[38;2;140;238;177m█[0m[38;2;141;239;176m█[0m       [38;2;145;244;168m░[0m[38;2;145;245;167m█[0m[38;2;146;246;166m█[0m[38;2;147;247;164m█[0m[38;2;147;247;163m█[0m   [38;2;149;250;159m░[0m[38;2;150;251;158m█[0m[38;2;150;252;157m█[0m[38;2;151;252;156m█[0m[38;2;151;253;155m█[0m 
		[38;2;102;187;255m░[0m[38;2;102;187;253m█[0m[38;2;103;188;252m█[0m[38;2;103;189;251m█[0m[38;2;104;189;250m█[0m       [38;2;108;195;241m░[0m[38;2;109;196;240m█[0m[38;2;109;197;239m█[0m[38;2;110;197;238m█[0m[38;2;110;198;237m█[0m [38;2;111;200;235m░[0m[38;2;112;200;234m█[0m[38;2;112;201;233m█[0m[38;2;113;202;232m█[0m[38;2;113;202;231m█[0m   [38;2;116;205;226m░[0m[38;2;116;206;225m█[0m[38;2;117;207;224m█[0m[38;2;117;207;223m█[0m[38;2;118;208;222m█[0m      [38;2;122;213;214m░[0m[38;2;122;214;213m█[0m[38;2;123;215;212m█[0m[38;2;123;215;211m█[0m[38;2;124;216;210m█[0m  [38;2;125;218;207m░[0m[38;2;126;219;206m█[0m[38;2;126;220;205m█[0m[38;2;127;221;204m█[0m[38;2;128;221;202m█[0m [38;2;129;223;200m░[0m[38;2;129;223;199m█[0m[38;2;130;224;198m█[0m[38;2;130;225;197m█[0m[38;2;131;226;196m█[0m [38;2;132;227;194m░[0m[38;2;132;228;193m█[0m[38;2;133;228;192m█[0m[38;2;134;229;190m█[0m[38;2;134;230;189m█[0m       [38;2;138;236;181m░[0m[38;2;139;236;180m█[0m[38;2;139;237;179m█[0m[38;2;140;238;177m█[0m[38;2;141;239;176m█[0m       [38;2;145;244;168m░[0m[38;2;145;245;167m█[0m[38;2;146;246;166m█[0m[38;2;147;247;164m█[0m[38;2;147;247;163m█[0m   [38;2;149;250;159m░[0m[38;2;150;251;158m█[0m[38;2;150;252;157m█[0m[38;2;151;252;156m█[0m[38;2;151;253;155m█[0m 
		[38;2;102;187;255m░[0m[38;2;102;187;253m█[0m[38;2;103;188;252m█[0m[38;2;103;189;251m█[0m[38;2;104;189;250m█[0m       [38;2;108;195;241m░[0m[38;2;109;196;240m█[0m[38;2;109;197;239m█[0m[38;2;110;197;238m█[0m[38;2;110;198;237m█[0m[38;2;111;199;236m█[0m[38;2;111;200;235m█[0m[38;2;112;200;234m█[0m[38;2;112;201;233m█[0m[38;2;113;202;232m█[0m[38;2;113;202;231m█[0m   [38;2;116;205;226m░[0m[38;2;116;206;225m█[0m[38;2;117;207;224m█[0m[38;2;117;207;223m█[0m[38;2;118;208;222m█[0m      [38;2;122;213;214m░[0m[38;2;122;214;213m█[0m[38;2;123;215;212m█[0m[38;2;123;215;211m█[0m[38;2;124;216;210m█[0m  [38;2;125;218;207m░[0m[38;2;126;219;206m█[0m[38;2;126;220;205m█[0m[38;2;127;221;204m█[0m[38;2;128;221;202m█[0m[38;2;128;222;201m█[0m[38;2;129;223;200m█[0m[38;2;129;223;199m█[0m[38;2;130;224;198m█[0m[38;2;130;225;197m█[0m[38;2;131;226;196m█[0m [38;2;132;227;194m░[0m[38;2;132;228;193m█[0m[38;2;133;228;192m█[0m[38;2;134;229;190m█[0m[38;2;134;230;189m█[0m       [38;2;138;236;181m░[0m[38;2;139;236;180m█[0m[38;2;139;237;179m█[0m[38;2;140;238;177m█[0m[38;2;141;239;176m█[0m       [38;2;145;244;168m░[0m[38;2;145;245;167m█[0m[38;2;146;246;166m█[0m[38;2;147;247;164m█[0m[38;2;147;247;163m█[0m   [38;2;149;250;159m░[0m[38;2;150;251;158m█[0m[38;2;150;252;157m█[0m[38;2;151;252;156m█[0m[38;2;151;253;155m█[0m
		[38;2;102;187;255m░[0m[38;2;102;187;253m█[0m[38;2;103;188;252m█[0m[38;2;103;189;251m█[0m[38;2;104;189;250m█[0m       [38;2;108;195;241m░[0m[38;2;109;196;240m█[0m[38;2;109;197;239m█[0m[38;2;110;197;238m█[0m[38;2;110;198;237m█[0m         [38;2;116;205;226m░[0m[38;2;116;206;225m█[0m[38;2;117;207;224m█[0m[38;2;117;207;223m█[0m[38;2;118;208;222m█[0m      [38;2;122;213;214m░[0m[38;2;122;214;213m█[0m[38;2;123;215;212m█[0m[38;2;123;215;211m█[0m[38;2;124;216;210m█[0m  [38;2;125;218;207m░[0m[38;2;126;219;206m█[0m[38;2;126;220;205m█[0m[38;2;127;221;204m█[0m[38;2;128;221;202m█[0m       [38;2;132;227;194m░[0m[38;2;132;228;193m█[0m[38;2;133;228;192m█[0m[38;2;134;229;190m█[0m[38;2;134;230;189m█[0m       [38;2;138;236;181m░[0m[38;2;139;236;180m█[0m[38;2;139;237;179m█[0m[38;2;140;238;177m█[0m[38;2;141;239;176m█[0m       [38;2;145;244;168m░[0m[38;2;145;245;167m█[0m[38;2;146;246;166m█[0m[38;2;147;247;164m█[0m[38;2;147;247;163m█[0m   [38;2;149;250;159m░[0m[38;2;150;251;158m█[0m[38;2;150;252;157m█[0m[38;2;151;252;156m█[0m[38;2;151;253;155m█[0m 
		[38;2;102;187;255m░[0m[38;2;102;187;253m█[0m[38;2;103;188;252m█[0m[38;2;103;189;251m█[0m[38;2;104;189;250m█[0m       [38;2;108;195;241m░[0m[38;2;109;196;240m█[0m[38;2;109;197;239m█[0m[38;2;110;197;238m█[0m[38;2;110;198;237m█[0m [38;2;111;200;235m░[0m[38;2;112;200;234m█[0m[38;2;112;201;233m█[0m[38;2;113;202;232m█[0m[38;2;113;202;231m█[0m   [38;2;116;205;226m░[0m[38;2;116;206;225m█[0m[38;2;117;207;224m█[0m[38;2;117;207;223m█[0m[38;2;118;208;222m█[0m      [38;2;122;213;214m░[0m[38;2;122;214;213m█[0m[38;2;123;215;212m█[0m[38;2;123;215;211m█[0m[38;2;124;216;210m█[0m  [38;2;125;218;207m░[0m[38;2;126;219;206m█[0m[38;2;126;220;205m█[0m[38;2;127;221;204m█[0m[38;2;128;221;202m█[0m [38;2;129;223;200m░[0m[38;2;129;223;199m█[0m[38;2;130;224;198m█[0m[38;2;130;225;197m█[0m[38;2;131;226;196m█[0m [38;2;132;227;194m░[0m[38;2;132;228;193m█[0m[38;2;133;228;192m█[0m[38;2;134;229;190m█[0m[38;2;134;230;189m█[0m       [38;2;138;236;181m░[0m[38;2;139;236;180m█[0m[38;2;139;237;179m█[0m[38;2;140;238;177m█[0m[38;2;141;239;176m█[0m [38;2;142;240;174m░[0m[38;2;142;241;173m█[0m[38;2;143;241;172m█[0m[38;2;143;242;171m█[0m[38;2;144;243;170m█[0m [38;2;145;244;168m░[0m[38;2;145;245;167m█[0m[38;2;146;246;166m█[0m[38;2;147;247;164m█[0m[38;2;147;247;163m█[0m   [38;2;149;250;159m░[0m[38;2;150;251;158m█[0m[38;2;150;252;157m█[0m[38;2;151;252;156m█[0m[38;2;151;253;155m█[0m  
		[38;2;102;187;255m░[0m[38;2;102;187;253m█[0m[38;2;103;188;252m█[0m[38;2;103;189;251m█[0m[38;2;104;189;250m█[0m[38;2;104;190;249m█[0m[38;2;105;191;248m█[0m[38;2;105;192;247m█[0m[38;2;106;192;246m█[0m[38;2;106;193;245m█[0m[38;2;107;194;244m█[0m   [38;2;109;197;239m░[0m[38;2;110;197;238m█[0m[38;2;110;198;237m█[0m[38;2;111;199;236m█[0m[38;2;111;200;235m█[0m[38;2;112;200;234m█[0m[38;2;112;201;233m█[0m       [38;2;117;207;224m░[0m[38;2;117;207;223m█[0m[38;2;118;208;222m█[0m[38;2;118;209;221m█[0m[38;2;119;210;220m█[0m      [38;2;123;215;212m░[0m[38;2;123;215;211m█[0m[38;2;124;216;210m█[0m[38;2;124;217;209m█[0m[38;2;125;218;208m█[0m  [38;2;126;220;205m░[0m[38;2;127;221;204m█[0m[38;2;128;221;202m█[0m[38;2;128;222;201m█[0m[38;2;129;223;200m█[0m[38;2;129;223;199m█[0m[38;2;130;224;198m█[0m   [38;2;132;227;194m░[0m[38;2;132;228;193m█[0m[38;2;133;228;192m█[0m[38;2;134;229;190m█[0m[38;2;134;230;189m█[0m         [38;2;139;237;179m░[0m[38;2;140;238;177m█[0m[38;2;141;239;176m█[0m[38;2;141;239;175m█[0m[38;2;142;240;174m█[0m[38;2;142;241;173m█[0m[38;2;143;241;172m█[0m   [38;2;145;244;168m░[0m[38;2;145;245;167m█[0m[38;2;146;246;166m█[0m[38;2;147;247;164m█[0m[38;2;147;247;163m█[0m [38;2;148;249;161m░[0m[38;2;149;249;160m█[0m[38;2;149;250;159m█[0m[38;2;150;251;158m█[0m[38;2;150;252;157m█[0m[38;2;151;252;156m█[0m[38;2;151;253;155m█[0m[38;2;152;254;154m█[0m[38;2;153;255;153m█[0m
	`

type Quote struct {
	Text   string
	Source string
}

type quoteLoadedMsg struct {
	quote Quote
	err   error
}

type item string
type menuItem item

func (i menuItem) FilterValue() string { return string(i) }

type menuItemDelegate struct{}

func (d menuItemDelegate) Height() int                             { return 1 }
func (d menuItemDelegate) Spacing() int                            { return 0 }
func (d menuItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d menuItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(menuItem)
	if !ok {
		return
	}
	itemText := string(i)
	numberStr := fmt.Sprintf("%d", index+1)
	var circleColor lipgloss.Color
	colorIndex := index % 4
	switch colorIndex {
	case 0:
		circleColor = cliBlue
	case 1:
		circleColor = cliGreen
	case 2:
		circleColor = cliOrange
	case 3:
		circleColor = cliBlue
	default:
		circleColor = cliBlue
	}

	numCircle := navCircleStyle(circleColor, numberStr).String()
	numCircleStyled := lipgloss.NewStyle().Width(5).Render(numCircle)

	shortcut := menuShortcutStyle.Render(fmt.Sprintf("[%d]", index+1))
	shortcutWidth := lipgloss.Width(shortcut)

	if index == m.Index() {
		itemTextStyled := navItemSelectedStyle.Render(itemText)

		paddingNeeded := m.Width() - lipgloss.Width(numCircleStyled) - lipgloss.Width(itemTextStyled) - shortcutWidth
		padding := strings.Repeat(" ", max(1, paddingNeeded))

		line := lipgloss.JoinHorizontal(lipgloss.Left,
			numCircleStyled,
			itemTextStyled,
			padding,
			shortcut,
		)
		fmt.Fprint(w, menuItemSelectedBgStyle.Width(m.Width()).Render(line))
	} else {
		itemTextStyled := navItemStyle.Render(itemText)

		paddingNeeded := m.Width() - lipgloss.Width(numCircleStyled) - lipgloss.Width(itemTextStyled) - shortcutWidth
		padding := strings.Repeat(" ", max(1, paddingNeeded))

		line := lipgloss.JoinHorizontal(lipgloss.Left,
			numCircleStyled,
			itemTextStyled,
			padding,
			shortcut,
		)
		fmt.Fprint(w, line)
	}
}

type MenuModel struct {
	list             list.Model
	Choice           string
	Quitting         bool
	termWidth        int
	termHeight       int
	quoteText        string
	quoteSource      string
	err              error
	showHelp         bool
	quoteLoadStarted bool
}

func NewMenuModel() MenuModel {
	items := []list.Item{
		menuItem(item("search movie")),
		menuItem(item("user profile")),
		menuItem(item("diary")),
		menuItem(item("watchlist")),
		menuItem(item("view lists")),
	}
	const defaultWidth = 35
	listHeight := len(items)
	l := list.New(items, menuItemDelegate{}, defaultWidth, listHeight)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	l.Styles = list.Styles{}
	return MenuModel{
		list:             l,
		termWidth:        80,
		termHeight:       24,
		quoteText:        "Loading quote...",
		quoteSource:      "",
		showHelp:         false,
		quoteLoadStarted: false,
	}
}

func loadRandomQuote() tea.Msg {
	var csvPath string
	snapDir := os.Getenv("SNAP")

	if snapDir != "" {
		csvPath = filepath.Join(snapDir, "assets", "movie_quotes.csv")
		if _, err := os.Stat(csvPath); os.IsNotExist(err) {
			return quoteLoadedMsg{err: fmt.Errorf("in snap, file not found at %s", csvPath)}
		}
	} else {
		goExecPath, err := os.Executable()
		if err != nil {
			return quoteLoadedMsg{err: fmt.Errorf("could not get executable path: %w", err)}
		}
		baseDir := filepath.Dir(goExecPath)
		csvPath = filepath.Join(baseDir, "assets", "movie_quotes.csv")

		if _, err := os.Stat(csvPath); os.IsNotExist(err) {
			wd, _ := os.Getwd()
			altCsvPath := filepath.Join(wd, "assets", "movie_quotes.csv")
			if _, altErr := os.Stat(altCsvPath); os.IsNotExist(altErr) {
				altCsvPath2 := filepath.Join(wd, "..", "..", "assets", "movie_quotes.csv")
				if _, altErr2 := os.Stat(altCsvPath2); os.IsNotExist(altErr2) {
					return quoteLoadedMsg{err: fmt.Errorf("quotes.csv not found at %s or %s or %s", csvPath, altCsvPath, altCsvPath2)}
				}
				csvPath = altCsvPath2
			} else {
				csvPath = altCsvPath
			}
		}
	}

	file, err := os.Open(csvPath)
	if err != nil {
		return quoteLoadedMsg{err: fmt.Errorf("could not open movie_quotes.csv: %w", err)}
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return quoteLoadedMsg{err: fmt.Errorf("failed to parse movie_quotes.csv: %w", err)}
	}

	if len(records) <= 1 {
		return quoteLoadedMsg{err: fmt.Errorf("movie_quotes.csv is empty or has no header")}
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomIndex := r.Intn(len(records)-1) + 1
	record := records[randomIndex]

	if len(record) < 4 {
		return quoteLoadedMsg{err: fmt.Errorf("invalid record in quotes.csv: expected 4 columns, got %d", len(record))}
	}

	quote := Quote{
		Text:   fmt.Sprintf(`"%s"`, record[0]),
		Source: fmt.Sprintf("%s (%s)", record[1], record[3]),
	}

	return quoteLoadedMsg{quote: quote}
}

func (m MenuModel) Init() tea.Cmd {
	return loadRandomQuote
}

func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if !m.quoteLoadStarted && m.quoteText == "Loading quote..." {
		m.quoteLoadStarted = true
		cmds = append(cmds, loadRandomQuote)
	}

	switch msg := msg.(type) {
	case quoteLoadedMsg:
		m.quoteLoadStarted = true
		if msg.err != nil {
			m.err = msg.err
			m.quoteText = "Failed to load quote."
			m.quoteSource = msg.err.Error()
		} else {
			m.quoteText = msg.quote.Text
			m.quoteSource = msg.quote.Source
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height
		listWidth := max(35, (m.termWidth*4/10)-4)
		m.list.SetSize(listWidth, len(m.list.Items()))
		return m, nil

	case tea.KeyMsg:
		if m.showHelp {
			switch msg.String() {
			case "?", "esc":
				m.showHelp = false
				return m, nil
			case "q", "ctrl+c":
				m.Quitting = true
				return m, tea.Quit
			default:
				return m, nil
			}
		}

		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.Quitting = true
			return m, tea.Quit

		case "?":
			m.showHelp = true
			return m, nil

		case "esc":
			return m, nil

		case "1", "2", "3", "4", "5":
			index := int(keypress[0] - '1')
			if index < len(m.list.Items()) {
				m.list.Select(index)
				i, ok := m.list.SelectedItem().(menuItem)
				if ok {
					switch string(i) {
					case "search movie":
						m.Choice = "Search a movie"
					case "user profile":
						m.Choice = "View a person's profile"
					case "diary":
						m.Choice = "Get diary of a person"
					case "watchlist":
						m.Choice = "Get Watchlist"
					case "view lists":
						m.Choice = "View Lists of Letterboxd"
					default:
						m.Choice = ""
					}
					if m.Choice != "" {
						return m, tea.Quit
					}
				}
				return m, nil
			}

		case "enter":
			i, ok := m.list.SelectedItem().(menuItem)
			if ok {
				switch string(i) {
				case "search movie":
					m.Choice = "Search a movie"
				case "user profile":
					m.Choice = "View a person's profile"
				case "diary":
					m.Choice = "Get diary of a person"
				case "watchlist":
					m.Choice = "Get Watchlist"
				case "view lists":
					m.Choice = "View Lists of Letterboxd"
				default:
					m.Choice = ""
				}
			}
			if m.Choice != "" {
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	if !m.Quitting && m.Choice == "" {
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m MenuModel) renderHelpView() string {
	helpTitle := menuHelpTitleStyle.Render("Help / Keybindings")

	keys := []string{
		"↑ / k", "Navigate Up",
		"↓ / j", "Navigate Down",
		"1-5", "Quick Select Item",
		"enter", "Confirm Selection",
		"?", "Toggle This Help Menu",
		"esc", "Close Help Menu / Go Back",
		"q / ctrl+c", "Quit LetterCLI",
	}

	var helpLines []string
	for i := 0; i < len(keys); i += 2 {
		line := lipgloss.JoinHorizontal(lipgloss.Left,
			menuHelpKeyStyle.Render(keys[i]),
			menuHelpDescStyle.Render(keys[i+1]),
		)
		helpLines = append(helpLines, line)
	}

	helpBlock := lipgloss.JoinVertical(lipgloss.Left,
		helpTitle,
		"\n",
		strings.Join(helpLines, "\n"),
		"\n\n(Press '?' or 'esc' to close)",
	)

	helpBox := menuBorderBox.Render(helpBlock)
	return lipgloss.Place(
		m.termWidth, m.termHeight,
		lipgloss.Center, lipgloss.Center,
		helpBox,
	)
}

func (m MenuModel) View() string {
	if m.Quitting {
		return menuQuitTextStyle.Render("Exiting LetterCLI...")
	}
	if m.Choice != "" {
		return menuQuitTextStyle.Render(fmt.Sprintf("Selected: %s", m.Choice))
	}
	if m.showHelp {
		return appLayoutStyle.Render(m.renderHelpView())
	}

	header := lipgloss.JoinVertical(lipgloss.Left,
		menuLogoStyle.Render(menuLogo),
		headerSubtitleStyle.String(),
		renderGradientLine(m.termWidth-4),
	)

	colorSystem := lipgloss.JoinVertical(lipgloss.Left,
		panelTitleStyle.Render("COLOR SYSTEM"),
		colorBlockBoxStyle.Render(
			lipgloss.JoinVertical(lipgloss.Left,
				colorBlockStyle.Copy().Background(cliOrange).Render(),
				lipgloss.JoinHorizontal(lipgloss.Left, colorNameStyle.Render("Primary Orange"), colorHexStyle.Render("#FF8000")),
			),
		),
		colorBlockBoxStyle.Render(
			lipgloss.JoinVertical(lipgloss.Left,
				colorBlockStyle.Copy().Background(cliGreen).Render(),
				lipgloss.JoinHorizontal(lipgloss.Left, colorNameStyle.Render("Success Green"), colorHexStyle.Render("#00E054")),
			),
		),
		colorBlockBoxStyle.Render(
			lipgloss.JoinVertical(lipgloss.Left,
				colorBlockStyle.Copy().Background(cliBlue).Render(),
				lipgloss.JoinHorizontal(lipgloss.Left, colorNameStyle.Render("Info Blue"), colorHexStyle.Render("#40B0FF")),
			),
		),
	)

	leftContent := lipgloss.JoinVertical(lipgloss.Left,
		panelTitleStyle.Render("NAVIGATION"),
		m.list.View(),
		colorSystem,
	)

	quote := lipgloss.JoinVertical(lipgloss.Left,
		panelTitleStyle.Render("FEATURED QUOTE"),
		quoteTextStyle.Render(m.quoteText),
		quoteAuthorStyle.Render(m.quoteSource),
	)

	quickTips := lipgloss.JoinVertical(lipgloss.Left,
		panelTitleStyle.Render("QUICK TIPS"),
		lipgloss.JoinHorizontal(lipgloss.Left, tipBulletStyle.String(), " ", tipTextStyle.Render("Press [1-5] for quick navigation")),
		lipgloss.JoinHorizontal(lipgloss.Left, tipBulletStyle.String(), " ", tipTextStyle.Render("Use ESC to return to menu")),
		lipgloss.JoinHorizontal(lipgloss.Left, tipBulletStyle.String(), " ", tipTextStyle.Render("Press '?' for help")),
	)

	rightFooter := footerTextStyle.Render("Terminal-based interface for Letterboxd\n")

	rightContent := lipgloss.JoinVertical(lipgloss.Left,
		quote,
		"\n",
		quickTips,
		"\n",
		rightFooter,
	)

	leftWidth := max(35, (m.termWidth * 4 / 10))
	rightWidth := m.termWidth - leftWidth - 6
	if rightWidth < 45 {
		rightWidth = 45
	}
	if leftWidth+rightWidth+6 > m.termWidth {
		delta := (leftWidth + rightWidth + 6) - m.termWidth
		if leftWidth-delta > 35 {
			leftWidth -= delta
		} else {
			rightWidth -= delta
		}
	}

	mainHeight := m.termHeight - lipgloss.Height(header) - 4
	if mainHeight < 15 {
		mainHeight = 15
	}

	leftColRender := lipgloss.NewStyle().Width(leftWidth).Height(mainHeight).Padding(0, 1).Render(leftContent)
	rightColRender := lipgloss.NewStyle().Width(rightWidth).Height(mainHeight).Padding(0, 1).Render(rightContent)

	mainContent := lipgloss.JoinHorizontal(lipgloss.Top,
		leftColRender,
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(dimGrey).
			Height(mainHeight).
			Render(""),
		rightColRender,
	)

	finalView := lipgloss.JoinVertical(lipgloss.Left,
		header,
		mainContent,
	)

	return appLayoutStyle.Render(finalView)
}
