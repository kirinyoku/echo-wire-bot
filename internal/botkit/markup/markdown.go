package markup

import "strings"

var (
	// replacer is a strings.Replacer that escapes special characters required for MarkdownV2.
	// Each special character in MarkdownV2 needs to be prefixed with a backslash (`\`) to ensure
	// it is displayed as plain text rather than being interpreted as formatting syntax.
	replacer = strings.NewReplacer(
		"-",
		"\\-",
		"_",
		"\\_",
		"*",
		"\\*",
		"[",
		"\\[",
		"]",
		"\\]",
		"(",
		"\\(",
		")",
		"\\)",
		"~",
		"\\~",
		"`",
		"\\`",
		">",
		"\\>",
		"#",
		"\\#",
		"+",
		"\\+",
		"=",
		"\\=",
		"|",
		"\\|",
		"{",
		"\\{",
		"}",
		"\\}",
		".",
		"\\.",
		"!",
		"\\!",
	)
)

// EscapeForMarkdown escapes all special characters in the input string `src`
// to ensure it is safe for use with Telegram's MarkdownV2 formatting.
func EscapeForMarkdown(src string) string {
	return replacer.Replace(src)
}
