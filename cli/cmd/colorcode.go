package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

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

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.Debug)
		fmt.Fprintln(w, "Code\tColor/Decoration Name\t\tSample\t")
		fmt.Fprintln(w, "----\t---------------------\t\t------\t")

		// Colors
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
			fmt.Fprintf(w, "%s\t%s\t-->\t%s\t\n", codeStr, c.name, sampleStr)
		}

		// Spacer
		fmt.Fprintln(w, "\t\t\t\t")

		// Decorations
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
			fmt.Fprintf(w, "%s\t%s\t-->\t%s\t\n", codeStr, d.name, sampleStr)
		}

		w.Flush()
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
		// add some mock combining marks
		sb.WriteRune(combiningMarks[idx%len(combiningMarks)])
		sb.WriteRune(combiningMarks[(idx+1)%len(combiningMarks)])
	}
	return sb.String()
}
