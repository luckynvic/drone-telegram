package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	tgmd "github.com/Mad-Pixels/goldmark-tgmd"
	"github.com/appleboy/drone-template-lib/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func skipIfNoTelegramSecrets(t *testing.T) {
	t.Helper()
	if os.Getenv("TELEGRAM_TOKEN") == "" || os.Getenv("TELEGRAM_TO") == "" {
		t.Skip("TELEGRAM_TOKEN/TELEGRAM_TO not set; skipping integration test")
	}
}

func TestMissingDefaultConfig(t *testing.T) {
	var plugin Plugin

	err := plugin.Exec()

	assert.Error(t, err)
}

func TestMissingUserConfig(t *testing.T) {
	plugin := Plugin{
		Config: Config{
			Token: "123456789",
		},
	}

	err := plugin.Exec()

	assert.Error(t, err)
}

func TestDefaultMessageFormat(t *testing.T) {
	plugin := Plugin{
		Repo: Repo{
			FullName:  "appleboy/go-hello",
			Name:      "go-hello",
			Namespace: "appleboy",
		},
		Commit: Commit{
			Sha:     "e7c4f0a63ceeb42a39ac7806f7b51f3f0d204fd2",
			Author:  "Bo-Yi Wu",
			Branch:  "master",
			Message: "update travis",
		},
		Build: Build{
			Number: 101,
			Status: "success",
			Link:   "https://github.com/appleboy/go-hello",
		},
	}

	message := plugin.Message()

	assert.Equal(
		t,
		[]string{
			"✅ Build #101 of `appleboy/go-hello` success.\n\n📝 Commit by Bo-Yi Wu on `master`:\n``` update travis ```\n\n🌐 https://github.com/appleboy/go-hello",
		},
		message,
	)
}

func TestDefaultMessageFormatFromGitHub(t *testing.T) {
	plugin := Plugin{
		Config: Config{
			GitHub: true,
		},
		Repo: Repo{
			FullName:  "appleboy/go-hello",
			Name:      "go-hello",
			Namespace: "appleboy",
		},
		GitHub: GitHub{
			Workflow:  "test-workflow",
			Action:    "send notification",
			EventName: "push",
		},
	}

	message := plugin.Message()

	assert.Equal(
		t,
		[]string{"appleboy/go-hello/test-workflow triggered by appleboy (push)"},
		message,
	)
}

func TestSendMessage(t *testing.T) {
	plugin := Plugin{
		Repo: Repo{
			Name:      "go-hello",
			Namespace: "appleboy",
		},
		Commit: Commit{
			Sha:     "e7c4f0a63ceeb42a39ac7806f7b51f3f0d204fd2",
			Author:  "Bo-Yi Wu",
			Branch:  "master",
			Message: "update travis by drone plugin",
			Email:   "test@gmail.com",
		},
		Build: Build{
			Tag:    "1.0.0",
			Number: 101,
			Status: "success",
			Link:   "https://github.com/appleboy/go-hello",
		},

		Config: Config{
			Token: os.Getenv("TELEGRAM_TOKEN"),
			To: []string{
				os.Getenv("TELEGRAM_TO"),
				os.Getenv("TELEGRAM_TO") + ":appleboy@gmail.com",
				"中文ID",
				"1234567890",
			},
			Message:  "Test Telegram Chat Bot From Travis or Local, commit message: 『{{ build.message }}』",
			Photo:    []string{"tests/github.png", "1234", " "},
			Document: []string{"tests/gophercolor.png", "1234", " "},
			Sticker:  []string{"tests/github-logo.png", "tests/github.png", "1234", " "},
			Audio:    []string{"tests/audio.mp3", "1234", " "},
			Voice:    []string{"tests/voice.ogg", "1234", " "},
			Location: []string{"24.9163213 121.1424972", "1", " "},
			Venue: []string{
				"35.661777 139.704051 竹北體育館 新竹縣竹北市",
				"24.9163213 121.1424972",
				"1",
				" ",
			},
			Video: []string{"tests/video.mp4", "1234", " "},
			Debug: false,
		},
	}

	err := plugin.Exec()
	require.Error(t, err)

	plugin.Config.Format = formatMarkdown
	plugin.Config.Message = "Test escape under_score"
	err = plugin.Exec()
	require.Error(t, err)

	// disable message
	plugin.Config.Message = ""
	err = plugin.Exec()
	assert.Error(t, err)
}

func TestDisableWebPagePreviewMessage(t *testing.T) {
	skipIfNoTelegramSecrets(t)
	plugin := Plugin{
		Config: Config{
			Token:                 os.Getenv("TELEGRAM_TOKEN"),
			To:                    []string{os.Getenv("TELEGRAM_TO")},
			DisableWebPagePreview: true,
			Debug:                 false,
		},
	}

	plugin.Config.Message = "DisableWebPagePreview https://www.google.com.tw"
	err := plugin.Exec()
	require.NoError(t, err)

	// disable message
	plugin.Config.DisableWebPagePreview = false
	plugin.Config.Message = "EnableWebPagePreview https://www.google.com.tw"
	err = plugin.Exec()
	assert.NoError(t, err)
}

func TestDisableNotificationMessage(t *testing.T) {
	skipIfNoTelegramSecrets(t)
	plugin := Plugin{
		Config: Config{
			Token:               os.Getenv("TELEGRAM_TOKEN"),
			To:                  []string{os.Getenv("TELEGRAM_TO")},
			DisableNotification: true,
			Debug:               false,
		},
	}

	plugin.Config.Message = "DisableNotification https://www.google.com.tw"
	err := plugin.Exec()
	require.NoError(t, err)

	// disable message
	plugin.Config.DisableNotification = false
	plugin.Config.Message = "EnableNotification https://www.google.com.tw"
	err = plugin.Exec()
	assert.NoError(t, err)
}

func TestBotError(t *testing.T) {
	plugin := Plugin{
		Repo: Repo{
			Name:      "go-hello",
			Namespace: "appleboy",
		},
		Commit: Commit{
			Sha:     "e7c4f0a63ceeb42a39ac7806f7b51f3f0d204fd2",
			Author:  "Bo-Yi Wu",
			Branch:  "master",
			Message: "update travis by drone plugin",
		},
		Build: Build{
			Number: 101,
			Status: "success",
			Link:   "https://github.com/appleboy/go-hello",
		},

		Config: Config{
			Token:   "appleboy",
			To:      []string{os.Getenv("TELEGRAM_TO"), "中文ID", "1234567890"},
			Message: "Test Telegram Chat Bot From Travis or Local",
		},
	}

	err := plugin.Exec()
	assert.Error(t, err)
}

func TestTrimElement(t *testing.T) {
	var input, result []string

	input = []string{"1", "     ", "3"}
	result = []string{"1", "3"}

	assert.Equal(t, result, trimElement(input))

	input = []string{"1", "2"}
	result = []string{"1", "2"}

	assert.Equal(t, result, trimElement(input))
}

func TestEscapeMarkdown(t *testing.T) {
	provider := [][][]string{
		{
			{"user", "repo"},
			{"user", "repo"},
		},
		{
			{"user_name", "repo_name"},
			{`user\_name`, `repo\_name`},
		},
		{
			{"user_name_long", "user_name_long"},
			{`user\_name\_long`, `user\_name\_long`},
		},
		{
			{`user\_name\_long`, `repo\_name\_long`},
			{`user\_name\_long`, `repo\_name\_long`},
		},
		{
			{`user\_name\_long`, `repo\_name\_long`, ""},
			{`user\_name\_long`, `repo\_name\_long`},
		},
	}

	for _, testCase := range provider {
		assert.Equal(t, testCase[1], escapeMarkdown(testCase[0]))
	}
}

func TestEscapeMarkdownOne(t *testing.T) {
	provider := [][]string{
		{"user", "user"},
		{"user_name", `user\_name`},
		{"user_name_long", `user\_name\_long`},
		{`user\_name\_escaped`, `user\_name\_escaped`},
	}

	for _, testCase := range provider {
		assert.Equal(t, testCase[1], escapeMarkdownOne(testCase[0]))
	}
}

func TestConvertMarkdownV2One(t *testing.T) {
	provider := [][]string{
		{"user", "user"},
		{"user_name", `user\_name`},
		{"user_name_long", `user\_name\_long`},
		{"**bold** text", `***bold*** text`},
		{"*italic* text", `_italic_ text`},
		{"`code` block", "`code` block"},
		{"[link](url)", "[link](url)"},
		{"text (with parens)", `text \(with parens\)`},
		{"Changes from (luckynvic fork) to upstream", `Changes from \(luckynvic fork\) to upstream`},
		{"Issue #123 is fixed", `Issue \#123 is fixed`},
		{"PR #695 in 4s", `PR \#695 in 4s`},
	}

	md := tgmd.TGMD()
	var buf bytes.Buffer
	for _, testCase := range provider {
		buf.Reset()
		if err := md.Convert([]byte(testCase[0]), &buf); err != nil {
			t.Fatalf("failed to convert markdown: %v", err)
		}
		assert.Equal(t, testCase[1], strings.TrimSpace(buf.String()))
	}
}

func TestConvertMarkdownV2Fields(t *testing.T) {
	userName := "user_name"
	repoName := "repo_name"
	convertMarkdownV2Fields(&userName, &repoName)
	assert.Equal(t, `user\_name`, userName)
	assert.Equal(t, `repo\_name`, repoName)
}

func TestConvertMarkdownV2FieldsNewlines(t *testing.T) {
	provider := [][]string{
		{"Line1\nLine2", "Line1\nLine2"},
		{"Hello\n\nWorld", "Hello\n\nWorld"},
		{"**bold**\n_italic_\n`code`", "***bold***\n_italic_\n`code`"},
		{"\nLeading newline", "Leading newline"},
		{"Trailing newline\n", "Trailing newline"},
	}

	for _, testCase := range provider {
		input := testCase[0]
		expected := testCase[1]
		convertMarkdownV2Fields(&input)
		assert.Equal(t, expected, input, "input: %q", testCase[0])
	}
}

func TestConvertMarkdownV2OneNewlines(t *testing.T) {
	provider := [][]string{
		{"Hello\nWorld", "Hello\nWorld"},
		{"Hello\n\nWorld", "Hello\n\nWorld"},
		{"Line1\nLine2\nLine3", "Line1\nLine2\nLine3"},
		{"**bold**\n\n_italic_\n`code`", "***bold***\n\n_italic_\n`code`"},
	}

	md := tgmd.TGMD()
	var buf bytes.Buffer
	for _, testCase := range provider {
		buf.Reset()
		processed := strings.ReplaceAll(testCase[0], "\n", "  \n")
		if err := md.Convert([]byte(processed), &buf); err != nil {
			t.Fatalf("failed to convert markdown: %v", err)
		}
		assert.Equal(t, testCase[1], strings.TrimSpace(buf.String()))
	}
}

func TestParseTo(t *testing.T) {
	input := []string{"0", "1:1@gmail.com", "2:2@gmail.com", "3:3@gmail.com", "4", "5"}

	ids := parseTo(input, "1@gmail.com", false)
	assert.Equal(t, []int64{0, 4, 5, 1}, ids)

	ids = parseTo(input, "1@gmail.com", true)
	assert.Equal(t, []int64{1}, ids)

	ids = parseTo(input, "a@gmail.com", false)
	assert.Equal(t, []int64{0, 4, 5}, ids)

	ids = parseTo(input, "a@gmail.com", true)
	assert.Equal(t, []int64{0, 4, 5}, ids)

	// test empty ids
	ids = parseTo([]string{"", " ", "   "}, "a@gmail.com", true)
	assert.Empty(t, ids)
}

func TestGlobList(t *testing.T) {
	var input []string
	var result []string

	input = []string{"tests/gophercolor.png", "測試", "3"}
	result = []string{"tests/gophercolor.png"}
	assert.Equal(t, result, globList(input))

	input = []string{"tests/*.mp3"}
	result = []string{"tests/audio.mp3"}
	assert.Equal(t, result, globList(input))
}

func TestConvertLocation(t *testing.T) {
	var input string
	var result Location
	var empty bool

	input = "1"
	result, empty = convertLocation(input)

	assert.True(t, empty)
	assert.Equal(t, Location{}, result)

	// strconv.ParseInt: parsing "測試": invalid syntax
	input = "測試 139.704051"
	result, empty = convertLocation(input)

	assert.True(t, empty)
	assert.Equal(t, Location{}, result)

	// strconv.ParseInt: parsing "測試": invalid syntax
	input = "35.661777 測試"
	result, empty = convertLocation(input)

	assert.True(t, empty)
	assert.Equal(t, Location{}, result)

	input = "35.661777 139.704051"
	result, empty = convertLocation(input)

	assert.False(t, empty)
	assert.Equal(t, Location{
		Latitude:  float64(35.661777),
		Longitude: float64(139.704051),
	}, result)

	input = "35.661777 139.704051 title"
	result, empty = convertLocation(input)

	assert.False(t, empty)
	assert.Equal(t, Location{
		Title:     "title",
		Address:   "",
		Latitude:  float64(35.661777),
		Longitude: float64(139.704051),
	}, result)

	input = "35.661777 139.704051 title address"
	result, empty = convertLocation(input)

	assert.False(t, empty)
	assert.Equal(t, Location{
		Title:     "title",
		Address:   "address",
		Latitude:  float64(35.661777),
		Longitude: float64(139.704051),
	}, result)
}

func TestHTMLMessage(t *testing.T) {
	skipIfNoTelegramSecrets(t)
	plugin := Plugin{
		Repo: Repo{
			Name:      "go-hello",
			Namespace: "appleboy",
		},
		Commit: Commit{
			Sha:     "e7c4f0a63ceeb42a39ac7806f7b51f3f0d204fd2",
			Author:  "Bo-Yi Wu",
			Branch:  "master",
			Message: "test",
		},
		Build: Build{
			Number: 101,
			Status: "success",
			Link:   "https://github.com/appleboy/go-hello",
		},

		Config: Config{
			Token: os.Getenv("TELEGRAM_TOKEN"),
			To:    []string{os.Getenv("TELEGRAM_TO")},
			Message: `
Test HTML Format
<a href='https://google.com'>Google .com 1</a>
<a href='https://google.com'>Google .com 2</a>
<a href='https://google.com'>Google .com 3</a>
`,
			Format: formatHTML,
		},
	}

	assert.NoError(t, plugin.Exec())

	plugin.Config.MessageFile = "tests/message_html.txt"
	assert.NoError(t, plugin.Exec())
}

func TestMessageFile(t *testing.T) {
	skipIfNoTelegramSecrets(t)
	plugin := Plugin{
		Repo: Repo{
			Name:      "go-hello",
			Namespace: "appleboy",
		},
		Commit: Commit{
			Sha:     "e7c4f0a63ceeb42a39ac7806f7b51f3f0d204fd2",
			Author:  "Bo-Yi Wu",
			Branch:  "master",
			Message: "Freakin' macOS isn't fully case-sensitive..",
		},
		Build: Build{
			Number:   101,
			Status:   "success",
			Link:     "https://github.com/appleboy/go-hello",
			Started:  time.Now().Unix(),
			Finished: time.Now().Add(180 * time.Second).Unix(),
		},

		Config: Config{
			Token:       os.Getenv("TELEGRAM_TOKEN"),
			To:          []string{os.Getenv("TELEGRAM_TO")},
			MessageFile: "tests/message.txt",
		},
	}

	err := plugin.Exec()
	assert.NoError(t, err)
}

func TestTemplateVars(t *testing.T) {
	skipIfNoTelegramSecrets(t)
	plugin := Plugin{
		Repo: Repo{
			Name:      "go-hello",
			Namespace: "appleboy",
		},
		Commit: Commit{
			Sha:     "e7c4f0a63ceeb42a39ac7806f7b51f3f0d204fd2",
			Author:  "Bo-Yi Wu",
			Branch:  "master",
			Message: "This is a test commit msg",
		},
		Build: Build{
			Number:   101,
			Status:   "success",
			Link:     "https://github.com/appleboy/go-hello",
			Started:  time.Now().Unix(),
			Finished: time.Now().Add(180 * time.Second).Unix(),
		},

		Config: Config{
			Token:        os.Getenv("TELEGRAM_TOKEN"),
			To:           []string{os.Getenv("TELEGRAM_TO")},
			Format:       formatMarkdown,
			MessageFile:  "tests/message_template.txt",
			TemplateVars: `{"env":"testing","version":"1.2.0-SNAPSHOT"}`,
		},
	}

	err := plugin.Exec()
	assert.NoError(t, err)
}

func TestTemplateVarsFile(t *testing.T) {
	skipIfNoTelegramSecrets(t)
	plugin := Plugin{
		Repo: Repo{
			Name:      "go-hello",
			Namespace: "appleboy",
		},
		Commit: Commit{
			Sha:     "e7c4f0a63ceeb42a39ac7806f7b51f3f0d204fd2",
			Author:  "Bo-Yi Wu",
			Branch:  "master",
			Message: "This is a test commit msg",
		},
		Build: Build{
			Number: 101,
			Status: "success",
			Link:   "https://github.com/appleboy/go-hello",
		},

		Config: Config{
			Token:            os.Getenv("TELEGRAM_TOKEN"),
			To:               []string{os.Getenv("TELEGRAM_TO")},
			Format:           formatMarkdown,
			MessageFile:      "tests/message_template.txt",
			TemplateVarsFile: "tests/vars.json",
		},
	}

	err := plugin.Exec()
	assert.NoError(t, err)
}

func TestProxySendMessage(t *testing.T) {
	skipIfNoTelegramSecrets(t)
	plugin := Plugin{
		Repo: Repo{
			Name:      "go-hello",
			Namespace: "appleboy",
		},
		Commit: Commit{
			Sha:     "e7c4f0a63ceeb42a39ac7806f7b51f3f0d204fd2",
			Author:  "Bo-Yi Wu",
			Branch:  "master",
			Message: "start use proxy",
			Email:   "test@gmail.com",
		},
		Build: Build{
			Tag:    "1.0.0",
			Number: 101,
			Status: "success",
			Link:   "https://github.com/appleboy/go-hello",
		},

		Config: Config{
			Token:   os.Getenv("TELEGRAM_TOKEN"),
			To:      []string{os.Getenv("TELEGRAM_TO")},
			Message: "Send message from socks5 proxy URL.",
			Debug:   false,
			Socks5:  os.Getenv("SOCKS5"),
		},
	}

	err := plugin.Exec()
	assert.NoError(t, err)
}

func TestBuildTemplate(t *testing.T) {
	plugin := Plugin{
		Commit: Commit{
			Sha:     "e7c4f0a63ceeb42a39ac7806f7b51f3f0d204fd2",
			Author:  "Bo-Yi Wu",
			Branch:  "master",
			Message: "This is a test commit msg",
		},
		Build: Build{
			Number:   101,
			Status:   "success",
			Link:     "https://github.com/appleboy/go-hello",
			Started:  time.Now().Unix(),
			Finished: time.Now().Add(180 * time.Second).Unix(),
		},
	}

	_, err := template.RenderTrim(
		`
Sample message loaded from file.

Commit msg:  {{uppercasefirst commit.message}}

duration: {{duration build.started build.finished}}
`, plugin)
	assert.NoError(t, err)
}

func TestSuccessBlockTemplate(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected string
	}{
		{"success status", "success", "GOOD"},
		{"failure status", "failure", "BAD"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			plugin := Plugin{
				Build: Build{
					Status: tc.status,
				},
			}

			result, err := template.RenderTrim(
				`{{#success build.status}}GOOD{{else}}BAD{{/success}}`,
				plugin,
			)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestComplexTemplate(t *testing.T) {
	plugin := Plugin{
		Commit: Commit{
			Message: "fix: update default template",
			Author:  "luckynvic",
		},
		Build: Build{
			Number:   692,
			Status:   "success",
			Event:    "pull_request",
			Link:     "https://github.com/example/repo/actions/runs/123",
			Started:  1234567890,
		},
		Repo: Repo{
			Namespace: "monitoring-pendapatan",
			Name:      "master",
		},
	}

	tmpl := `{{#success build.status}}
✅ Build succeeded: {{build.number}}
🧙 : {{commit.author}}
🕐 : {{build.event}}
📚 : {{repo.namespace}}:{{repo.name}}
{{else}}
🔥 {{commit.message}}
🧙 : {{commit.author}}
❗ Failed. Check {{build.link}} for more details.
{{/success}}`

	result, err := template.RenderTrim(tmpl, plugin)
	assert.NoError(t, err)
	t.Logf("Result:\n%s", result)
}

func TestMarkdownV2FieldsEscapedBeforeTemplateRender(t *testing.T) {
	plugin := Plugin{
		Commit: Commit{
			Author:  "luckynvic",
			Branch:  "bugfix/invoice-footer",
			Message: "Changes from `bugfix/invoice-footer` (luckynvic fork) to upstream/master:",
		},
		Build: Build{
			Number:  695,
			Status:  "success",
			Event:   "pull_request",
			Link:    "https://github.com/example/repo/actions/runs/123",
			Started: 1234567890,
		},
		Repo: Repo{
			Namespace: "monitoring-pendapatan",
			Name:      "master",
		},
		Config: Config{
			Format: formatMarkdownV2,
		},
	}

	tmpl := `✅ {{commit.message}}
🧙 : {{commit.author}}
🕐 : {{build.event}}
📚 : {{repo.name}}:{{commit.branch}}`

	result, err := template.RenderTrim(tmpl, plugin)
	assert.NoError(t, err)

	result = convertMarkdownV2String(result)

	assert.Contains(t, result, `\(luckynvic fork\)`)
	assert.NotContains(t, result, `(luckynvic fork)`)
}

func TestMarkdownV2CommitMessageWithMarkdownSyntax(t *testing.T) {
	plugin := Plugin{
		Commit: Commit{
			Author:  "developer",
			Branch:  "main",
			Message: "feat: add new feature\n\nThis PR includes:\n- **feature1**\n- *feature2*\n- See [docs](https://example.com)",
		},
		Build: Build{
			Number:  100,
			Status:  "success",
			Event:   "pull_request",
			Link:    "https://github.com/example/repo/actions/runs/123",
			Started: 1234567890,
		},
		Repo: Repo{
			Namespace: "example",
			Name:      "repo",
		},
		Config: Config{
			Format: formatMarkdownV2,
		},
	}

	tmpl := `{{commit.message}}`

	result, err := template.RenderTrim(tmpl, plugin)
	assert.NoError(t, err)

	result = convertMarkdownV2String(result)

	assert.Contains(t, result, `***feature1***`)
	assert.Contains(t, result, `_feature2_`)
	assert.Contains(t, result, `[docs](https://example.com)`)
}

func TestTemplatePlaceholderReplacement(t *testing.T) {
	plugin := Plugin{
		Commit: Commit{
			Author:  "testuser",
			Branch:  "feature-branch",
			Message: "test commit message",
		},
		Build: Build{
			Number:  123,
			Status:  "success",
			Event:   "push",
			Link:    "https://github.com/test/repo/actions/runs/456",
			Started: 1234567890,
		},
		Repo: Repo{
			Namespace: "testorg",
			Name:      "testrepo",
		},
	}

	tmpl := `Author: {{commit.author}}
Branch: {{commit.branch}}
Message: {{commit.message}}
Build: #{{build.number}}
Status: {{build.status}}
Event: {{build.event}}
Repo: {{repo.namespace}}/{{repo.name}}`

	result, err := template.RenderTrim(tmpl, plugin)
	assert.NoError(t, err)

	assert.Contains(t, result, "Author: testuser")
	assert.Contains(t, result, "Branch: feature-branch")
	assert.Contains(t, result, "Message: test commit message")
	assert.Contains(t, result, "Build: #123")
	assert.Contains(t, result, "Status: success")
	assert.Contains(t, result, "Event: push")
	assert.Contains(t, result, "Repo: testorg/testrepo")
}

func TestMarkdownV2TemplateWithHash(t *testing.T) {
	plugin := Plugin{
		Build: Build{
			Number: 695,
			Status: "success",
		},
		Config: Config{
			Format: formatMarkdownV2,
		},
	}

	tmpl := `🥇 : #{{build.number}} in 4s`

	result, err := template.RenderTrim(tmpl, plugin)
	assert.NoError(t, err)

	result = convertMarkdownV2String(result)

	assert.Contains(t, result, `\#695`)
}

func TestMarkdownV2AllReservedCharacters(t *testing.T) {
	reservedChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}

	for _, ch := range reservedChars {
		t.Run("char_"+ch, func(t *testing.T) {
			input := fmt.Sprintf("test %s test", ch)
			result := convertMarkdownV2String(input)

			if ch == "[" || ch == "]" || ch == "(" || ch == ")" {
				assert.Contains(t, result, `\`+ch+``)
			} else {
				assert.Contains(t, result, `\`+ch+``)
			}
		})
	}
}

func TestCommitMessageWithEmojisMarkdownV2(t *testing.T) {
	commitMessage := "<!-- :bulb: TIP: Delete sections or comments that are not relevant to this PR -->\n\n" +
		"## :memo: Description\n" +
		"Implements manual payment functionality for invoices.\n\n" +
		"## :link: Related Issues & Tickets\n" +
		"- Fixes: None\n" +
		"- Related Ticket: None\n\n" +
		"## :hammer_and_wrench: Type of Change\n" +
		"- [ ] :bug: Bug fix (non-breaking change which fixes an issue)\n" +
		"- [x] :sparkles: New feature (non-breaking change which adds functionality)\n" +
		"- [ ] :collision: Breaking change (fix or feature that would cause existing functionality to not work as expected)\n" +
		"- [ ] :zap: Performance / Code refactoring (no functional changes)\n" +
		"- [ ] :books: Documentation update\n\n" +
		"## :test_tube: How to Test & Verify\n" +
		"<!-- Step-by-step instructions for the reviewer to verify your changes. -->\n\n" +
		"### 1. Prerequisites\n" +
		"- [ ] PHP version checked ('php -v')\n" +
		"- [ ] Composer dependencies installed ('composer install')\n" +
		"- [ ] Database migrations applied ('php yii migrate')\n" +
		"- [ ] Codeception configured (if applicable)\n\n" +
		"### 2. Steps to Reproduce / Verify\n" +
		"1. Run relevant functional/unit tests ('./vendor/bin/codecept run')\n" +
		"2. Check affected module/controller behavior in local/UAT\n" +
		"3. Verify database schema changes if any migration included\n\n" +
		"### 3. Automated Tests Run\n" +
		"- [ ] Unit tests\n" +
		"- [ ] Functional / Acceptance tests (Codeception)\n" +
		"- [ ] Database migration tested (if applicable)\n\n" +
		"## :rocket: Pre-Merge Checklist\n" +
		"- [ ] Code follows Yii2 conventions and project coding standards\n" +
		"- [ ] I have performed a self-review of my own code\n" +
		"- [ ] I have commented on my code, particularly in complex logic\n" +
		"- [ ] I have made corresponding changes to the documentation\n" +
		"- [ ] No new PHP warnings/errors introduced\n" +
		"- [ ] Database migrations are backward-compatible (if applicable)\n" +
		"- [ ] No hardcoded credentials or secrets introduced"

	plugin := Plugin{
		Commit: Commit{
			Message: commitMessage,
		},
		Config: Config{
			Format: formatMarkdownV2,
		},
	}

	tmpl := "{{commit.message}}"

	result, err := template.RenderTrim(tmpl, plugin)
	assert.NoError(t, err)

	result = convertMarkdownV2String(result)

	assert.Contains(t, result, "💡")
	assert.Contains(t, result, "📝")
	assert.Contains(t, result, "🔗")
	assert.Contains(t, result, "🔧")
	assert.Contains(t, result, "✨")
	assert.Contains(t, result, "💥")
	assert.Contains(t, result, "⚡")
	assert.Contains(t, result, "📚")
	assert.Contains(t, result, "🧪")
	assert.Contains(t, result, "🚀")
}
