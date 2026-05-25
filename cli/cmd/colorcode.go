package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	symbolFlag string
	sampleFlag string
)

var colorcodeCmd = &cobra.Command{
	Use:   "colorcode",
	Short: "Show Minecraft Colorcodes",
	Long:  "Display Minecraft formatting and color codes with samples.",
	Run: func(cmd *cobra.Command, args []string) {
		symbol := "§"
		if symbolFlag == "amp" {
			symbol = "&"
		}

		infoColor := color.New(color.BgBlue, color.FgWhite).Sprint(" INFO ")
		greenBold := color.New(color.FgGreen, color.Bold)
		blueBold := color.New(color.FgBlue, color.Bold)

		fmt.Printf("%s %s %s\n", infoColor, greenBold.Sprint("The color code will be displayed with the symbol:"), blueBold.Sprint(symbol))
		fmt.Printf("%s %s %s\n\n", infoColor, greenBold.Sprint("The sample text will be displayed in this text:"), blueBold.Sprint(sampleFlag))

		// テーブルのスタイル定義 (外枠をスタイリッシュなパープルで装飾)
		borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))

		// lipgloss/tableの作成（エスケープシーケンスの文字数を自動で無視して幅計算してくれます）
		t := table.New().
			Border(lipgloss.NormalBorder()).
			BorderStyle(borderStyle).
			Headers("Code", "Color/Decoration Name", "Sample")

		// 1. Colorsのデータを追加
		colors := []struct {
			code, name, fgHex, bgHex string
		}{
			{"0", "black", "#000000", "#FFFFFF"},
			{"1", "dark_blue", "#0000AA", "#000000"},
			{"2", "dark_green", "#00AA00", "#000000"},
			{"3", "dark_aqua", "#00AAAA", "#000000"},
			{"4", "dark_red", "#AA0000", "#000000"},
			{"5", "dark_purple", "#AA00AA", "#000000"},
			{"6", "gold", "#FFAA00", "#000000"},
			{"7", "gray", "#AAAAAA", "#FFFFFF"},
			{"8", "dark_gray", "#555555", "#FFFFFF"},
			{"9", "blue", "#5555FF", "#000000"},
			{"a", "green", "#55FF55", "#000000"},
			{"b", "aqua", "#55FFFF", "#000000"},
			{"c", "red", "#FF5555", "#000000"},
			{"d", "light_purple", "#FF55FF", "#000000"},
			{"e", "yellow", "#FFFF55", "#000000"},
			{"f", "white", "#FFFFFF", "#000000"},
		}

		for _, c := range colors {
			codeStr := fmt.Sprintf("%s%s", symbol, c.code)
			sampleStr := paintHex(sampleFlag, c.fgHex, c.bgHex)
			t.Row(codeStr, c.name, sampleStr)
		}

		// 2. Decorations（装飾コード）のデータを追加
		decorations := []struct {
			code, name, format string
		}{
			{"k", "Obfuscated", "zalgo"},
			{"l", "Bold", "\x1b[1m"},
			{"m", "StrikeThrough", "\x1b[9m"},
			{"n", "Underline", "\x1b[4m"},
			{"o", "Italic", "\x1b[3m"},
			{"r", "Reset", "\x1b[0m"},
		}

		for _, d := range decorations {
			codeStr := fmt.Sprintf("%s%s", symbol, d.code)
			var sampleStr string
			if d.format == "zalgo" {
				sampleStr = fmt.Sprintf("%s (Garbled Characters)", zalgoText(sampleFlag))
			} else {
				sampleStr = fmt.Sprintf("%s%s\x1b[0m", d.format, sampleFlag)
			}
			t.Row(codeStr, d.name, sampleStr)
		}

		// 美しくレンダリングされたテーブルを標準出力に表示
		fmt.Fprintln(os.Stdout, t)
		fmt.Println()
	},
}

func init() {
	colorcodeCmd.Flags().StringVar(&symbolFlag, "symbol", "section", "Sets which symbol to display; 'amp' (&) or 'section' (§)")
	colorcodeCmd.Flags().StringVarP(&sampleFlag, "sampletxt", "t", "This is a sample text.", "Text which display to sample text")
	rootCmd.AddCommand(colorcodeCmd)
}

func paintHex(text, fgHex, bgHex string) string {
	var fr, fg, fb int
	fmt.Sscanf(fgHex, "#%02x%02x%02x", &fr, &fg, &fb)
	var br, bg, bb int
	fmt.Sscanf(bgHex, "#%02x%02x%02x", &br, &bg, &bb)

	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm%s\x1b[0m", fr, fg, fb, br, bg, bb, text)
}

func zalgoText(text string) string {
	combiningMarks := []rune{
		'\u0300', '\u0301', '\u0302', '\u0303', '\u0304', '\u0305', '\u0306', '\u0307', '\u0308', '\u0309', '\u030a', '\u030b', '\u030c', '\u030d', '\u030e', '\u030f',
		'\u0310', '\u0311', '\u0312', '\u0313', '\u0314', '\u0315', '\u0316', '\u0317', '\u0318', '\u0319', '\u031a', '\u031b', '\u031c', '\u031d', '\u031e', '\u031f',
	}
	var sb strings.Builder
	for idx, r := range text {
		sb.WriteRune(r)
		sb.WriteRune(combiningMarks[idx%len(combiningMarks)])
		sb.WriteRune(combiningMarks[(idx+1)%len(combiningMarks)])
	}
	return sb.String()
}
