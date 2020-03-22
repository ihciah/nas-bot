package main

import (
	"flag"

	"github.com/ihciah/nas-bot/aria2"
	"github.com/ihciah/nas-bot/ipmi"
	"github.com/ihciah/telebotex"
	"github.com/ihciah/telebotex/interceptor"
	"github.com/ihciah/telebotex/interceptor/auth"
	"github.com/ihciah/telebotex/plugin"
	"github.com/ihciah/telebotex/plugin/id_bot"
)

func main() {
	var configFile = flag.String("config", "config.json", "Config file")
	flag.Parse()

	bot := telebotex.MustNewBot(*configFile)
	authInterceptor, err := auth.NewAuthenticator(bot)
	if err != nil {
		panic(err)
	}
	plugins := []plugin.Plugin{
		id_bot.NewBot(),
		interceptor.NewInterceptedPlugin(ipmi.NewBot(), authInterceptor),
		interceptor.NewInterceptedPlugin(aria2.NewBot(), authInterceptor),
	}

	if err := bot.Register(plugins...); err != nil {
		panic(err)
	}

	bot.Start()
}
