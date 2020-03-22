package aria2

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/ihciah/telebotex"
	"github.com/ihciah/telebotex/bot"
	"github.com/ihciah/telebotex/plugin"
	jsoniter "github.com/json-iterator/go"
	tb "gopkg.in/tucnak/telebot.v2"
)

const (
	aria2Key       = "aria2"
	maxTorrentSize = 1024 * 1024
)

type Bot struct {
	plugin.BasePlugin
	cache *cache
	aria2 *aria2Cli
}

func (b *Bot) LoadConfig(config map[string]jsoniter.RawMessage) error {
	return telebotex.UnmarshalFromConfig(config, aria2Key, b.aria2)
}

func (b *Bot) Register(bot bot.TelegramBotExt) {
	inlineYesMoviesBtn := tb.InlineButton{
		Unique: "DownloadToMovies",
		Text:   "Yes to Movies",
	}
	inlineYesBangumiBtn := tb.InlineButton{
		Unique: "DownloadToBangumi",
		Text:   "Yes to Bangumi",
	}
	inlineNoBtn := tb.InlineButton{
		Unique: "Deny",
		Text:   "No",
	}
	inlineKeys := [][]tb.InlineButton{
		{inlineYesMoviesBtn, inlineYesBangumiBtn},
		{inlineNoBtn},
	}

	bot.Handle(&inlineYesMoviesBtn, b.downloadHandler(bot, moviesFolder))
	bot.Handle(&inlineYesBangumiBtn, b.downloadHandler(bot, bangumiFolder))

	bot.Handle(&inlineNoBtn, func(c *tb.Callback) {
		b.cache.Del(c.Message.ID)
		_ = bot.Respond(c, &tb.CallbackResponse{Text: "Downloading canceled."})
		_, _ = bot.EditReplyMarkup(c.Message, new(tb.ReplyMarkup))
	})

	bot.ConditionalHandle(tb.OnText, func(m *tb.Message) {
		link := strings.TrimSpace(m.Text)
		msg, err := bot.Reply(m, "Confirm downloading ?\n"+link, &tb.ReplyMarkup{
			InlineKeyboard: inlineKeys,
		})
		if err != nil {
			log.Printf("unable to send comfirm message")
			return
		}

		if err = b.cache.SetLink(msg.ID, link); err != nil {
			log.Printf("cache set error: %v", err)
			_, _ = bot.Reply(m, "Cache set error.")
		}
	}, isDownloadLink)

	bot.ConditionalHandle(tb.OnDocument, func(m *tb.Message) {
		fileName := m.Document.FileName
		fileSize := m.Document.FileSize
		if fileSize > maxTorrentSize {
			_, _ = bot.Reply(m, "torrent size is too big to handle")
			return
		}
		fileReader, err := bot.GetFile(&m.Document.File)
		if err != nil {
			log.Printf("read telegram file error: %v", err)
			_, _ = bot.Reply(m, fmt.Sprintf("read telegram file error: %v", err))
			return
		}

		fileData, err := ioutil.ReadAll(fileReader)
		if err != nil {
			log.Printf("read file content error: %v", err)
			_, _ = bot.Reply(m, fmt.Sprintf("read file content error: %v", err))
			return
		}

		msg, _ := bot.Reply(m, "Confirm downloading ?\n"+fileName, &tb.ReplyMarkup{
			InlineKeyboard: inlineKeys,
		})
		err = b.cache.SetFile(msg.ID, fileName, fileData)
		if err != nil {
			log.Printf("cache set error: %v", err)
			_, _ = bot.Reply(m, "Cache set error.")
		}
	}, isTorrent)
}

func isDownloadLink(m *tb.Message) bool {
	prefix := []string{"http://", "https://", "magnet:?xt=urn:btih:"}
	maybeLink := strings.ToLower(m.Text)
	for _, p := range prefix {
		if strings.HasPrefix(maybeLink, p) {
			return true
		}
	}
	return false
}

func isTorrent(m *tb.Message) bool {
	return m.Document != nil && strings.HasSuffix(strings.ToLower(m.Document.FileName), ".torrent")
}

func (b *Bot) downloadHandler(bot bot.TelegramBotExt, folder string) func(c *tb.Callback) {
	return func(c *tb.Callback) {
		union, err := b.cache.GetAndDel(c.Message.ID)
		if err != nil {
			_ = bot.Respond(c, &tb.CallbackResponse{Text: "Cache get error, link or file may expired."})
			return
		}
		go func() { bot.Respond(c, &tb.CallbackResponse{Text: "Downloading confirmed."}) }()

		caption := union.Link
		if union.IsFile() {
			caption = union.FileName
		}
		submittedMsg, err := bot.Edit(c.Message, fmt.Sprintf("Submiting task %s to %s", caption, folder))

		if union.IsLink() {
			err = b.aria2.AddUri(folder, union.Link)
		} else {
			err = b.aria2.AddTorrent(folder, union.FileData)
		}

		if err == nil {
			_, _ = bot.Edit(submittedMsg, fmt.Sprintf("Submitted task %s to %s", caption, folder))
		} else {
			_, _ = bot.Edit(submittedMsg, fmt.Sprintf("Task %s error: %v", caption, err))
		}
	}
}

func NewBot() *Bot {
	b := new(Bot)
	b.cache = newCache()
	b.aria2 = new(aria2Cli)
	return b
}
