package aria2

import (
	"fmt"
	"log"
	"strings"

	"github.com/ihciah/telebotex"
	"github.com/ihciah/telebotex/bot"
	"github.com/ihciah/telebotex/plugin"
	jsoniter "github.com/json-iterator/go"
	tb "gopkg.in/tucnak/telebot.v2"
)

const aria2Key = "aria2"

type Bot struct {
	plugin.BasePlugin
	cache *linkCache
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
		msg, _ := bot.Send(m.Sender, "Confirm downloading ?\n"+link, &tb.ReplyMarkup{
			InlineKeyboard: inlineKeys,
		})
		err := b.cache.Set(msg.ID, link)
		if err != nil {
			log.Printf("cache set error: %v", err)
			_, _ = bot.Send(m.Sender, "Cache set error.")
		}
	}, isDownloadLink)
}

func (b *Bot) downloadHandler(bot bot.TelegramBotExt, folder string) func(c *tb.Callback) {
	return func(c *tb.Callback) {
		link, err := b.cache.GetAndDel(c.Message.ID)
		if err != nil {
			_ = bot.Respond(c, &tb.CallbackResponse{Text: "Cache get error."})
			return
		}

		_ = bot.Respond(c, &tb.CallbackResponse{Text: "Downloading confirmed."})
		submittedMsg, err := bot.Send(c.Sender, fmt.Sprintf("Submiting task %s to %s", link, folder))
		err = b.aria2.AddUri(folder, link)
		if err == nil {
			_, _ = bot.Edit(submittedMsg, fmt.Sprintf("Submitted task %s to %s", link, folder))
		} else {
			_, _ = bot.Edit(submittedMsg, fmt.Sprintf("Task %s error: %v", link, err))
		}
	}
}

func isDownloadLink(m *tb.Message) bool {
	var prefix = []string{"http://", "https://", "magnet:?xt=urn:btih:"}
	for _, p := range prefix {
		if strings.HasPrefix(m.Text, p) {
			return true
		}
	}
	return false
}

func NewBot() *Bot {
	b := new(Bot)
	b.cache = newLinkCache()
	b.aria2 = new(aria2Cli)
	return b
}
