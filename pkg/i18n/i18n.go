package i18n

import (
	"encoding/json"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var Bundle *i18n.Bundle

func InitI18n(localesPath string) {
	Bundle = i18n.NewBundle(language.English)

	Bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	Bundle.MustLoadMessageFile(localesPath + "/active.en.json")
	Bundle.MustLoadMessageFile(localesPath + "/active.vi.json")
	Bundle.MustLoadMessageFile(localesPath + "/active.jp.json")
}

func GetMessage(langHeader string, messageID string) string {
	localizer := i18n.NewLocalizer(Bundle, langHeader)
	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID: messageID,
	})

	if err != nil {
		return messageID
	}

	return msg
}
