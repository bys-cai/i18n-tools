package i18n

import (
	"encoding/json"
	"path/filepath"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/text/language"
)

var bundle *i18n.Bundle
var localizers map[string]*i18n.Localizer

func MustLoadI18n() {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	if err := loadLocaleFiles(); err != nil {
		logx.Must(err)
	}

	initLocalizers()
}

func loadLocaleFiles() error {
	matches, err := filepath.Glob("etc/locale/*.json")
	if err != nil {
		return err
	}

	for _, file := range matches {
		if _, err := bundle.LoadMessageFile(file); err != nil {
			logx.Errorf("Failed to load locale file %s: %v", file, err)
		}
	}
	return nil
}

func initLocalizers() {
	languages := []string{"en", "zh", "it", "hi"}
	localizers = make(map[string]*i18n.Localizer)

	for _, lang := range languages {
		localizers[lang] = i18n.NewLocalizer(bundle, lang)
	}
}

func Trans(lang, messageID string, data ...map[string]interface{}) string {
	localizer := GetLocalizer(lang)

	config := &i18n.LocalizeConfig{
		MessageID: messageID,
	}

	if len(data) > 0 && data[0] != nil {
		config.TemplateData = data[0]
	}

	result, err := localizer.Localize(config)
	if err != nil {
		logx.Errorf("Translation failed for message %s: %v", messageID, err)
		return messageID
	}
	return result
}

func GetLocalizer(lang string) *i18n.Localizer {
	if localizer, exists := localizers[lang]; exists {
		return localizer
	}
	return localizers["en"]
}
