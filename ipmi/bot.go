package ipmi

import (
	"log"

	"github.com/ihciah/ipmi-controller/pkg/ipmi"
	"github.com/ihciah/telebotex"
	"github.com/ihciah/telebotex/plugin"
	jsoniter "github.com/json-iterator/go"
	tb "gopkg.in/tucnak/telebot.v2"
)

const ipmiKey = "ipmi"

type Bot struct {
	plugin.BasePlugin
	ipmi *ipmi.IPMI
}

func (b *Bot) LoadConfig(config map[string]jsoniter.RawMessage) error {
	cfg := new(ipmi.IPMIConfig)
	var err error

	if err = telebotex.UnmarshalFromConfig(config, ipmiKey, cfg); err != nil {
		log.Printf("unable to load ipmi config: %v", err)
		return err
	}
	if err = cfg.Validate(); err != nil {
		return err
	}
	b.ipmi = (*ipmi.IPMI)(cfg)
	return nil
}

func (b *Bot) Register(bot plugin.TelegramBot) {
	getStatusBtn := tb.ReplyButton{Text: "GetStatus"}
	getTemperatureBtn := tb.ReplyButton{Text: "GetTemperature"}
	setPowerOnBtn := tb.ReplyButton{Text: "PowerOn"}
	setPowerOffBtn := tb.ReplyButton{Text: "PowerOff"}
	setPowerResetBtn := tb.ReplyButton{Text: "PowerReset"}
	setPowerCycleBtn := tb.ReplyButton{Text: "PowerCycle"}

	replyKeys := [][]tb.ReplyButton{
		{getStatusBtn, getTemperatureBtn},
		{setPowerOnBtn, setPowerOffBtn, setPowerResetBtn, setPowerCycleBtn},
	}

	bot.Handle(&getStatusBtn, b.callbackFactory(bot, b.ipmi.GetStatus))
	bot.Handle(&getTemperatureBtn, b.callbackFactory(bot, b.ipmi.GetTemperature))
	bot.Handle(&setPowerOnBtn, b.callbackFactory(bot, b.ipmi.SetPowerOn))
	bot.Handle(&setPowerOffBtn, b.callbackFactory(bot, b.ipmi.SetPowerOff))
	bot.Handle(&setPowerResetBtn, b.callbackFactory(bot, b.ipmi.SetPowerReset))
	bot.Handle(&setPowerCycleBtn, b.callbackFactory(bot, b.ipmi.SetPowerCycle))
	bot.Handle("/ipmi", func(m *tb.Message) {
		if !m.Private() {
			return
		}
		_, _ = bot.Send(m.Sender, "Please choose the button", &tb.ReplyMarkup{
			ReplyKeyboard: replyKeys,
		})
	})
}

func (b *Bot) callbackFactory(bot plugin.TelegramBot, f func() (string, error)) func(m *tb.Message) {
	fun := f
	return func(m *tb.Message) {
		msg, err := bot.Reply(m, "Command executing...")
		if err != nil {
			return
		}
		output, err := fun()
		if err != nil {
			output = err.Error()
		}
		_, _ = bot.Edit(msg, output)
	}
}

func NewBot() *Bot {
	return new(Bot)
}
