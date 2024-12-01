package translate

import "fmt"

type Translator struct {
	defaultLang string
}

func NewTranslator(defaultLang string) *Translator {
	if defaultLang == "" {
		defaultLang = "en"
	}
	return &Translator{
		defaultLang: defaultLang,
	}
}

func (t *Translator) Translate(lang, key string, params ...any) string {
	template := t.getTranslateTemplate(lang, key)
	if len(params) == 0 {
		return template
	}
	return fmt.Sprintf(template, params...)
}

func (t *Translator) getTranslateTemplate(lang, key string) string {
	langs, ok := translates[key]
	if !ok {
		return key
	}
	templateByLang, ok := langs[lang]
	if !ok {
		templateByLang, ok = langs[t.defaultLang]
		if !ok {
			return key
		}
	}

	return templateByLang
}

var translates = map[string]map[string]string{
	"wishlist_title": {
		"en": "Wishlist",
		"ru": "Вишлист",
	},
	"init_description": {
		"en": "It's my personal Wishlist!\nI can add here different my wishes. " +
			"I can share my wishlist with my friends.\n" +
			"And I will be happy to receive any gift from my wishlist!",
		"ru": "Это мой личный Вишлист!\nЯ могу добавлять сюда разные свои желания. " +
			"Могу делиться своим вишлистом с друзьями.\n" +
			"И любой подарок из этого вишлиста сделает меня немного счастливее!",
	},
	"welcome_message": {
		"en": "Hello, I'm Wishlist Bot!\n\nI can help you to manage your own wishlist.\n\n" +
			"I have already created a personal wishlist for you! " +
			"Press \"Open Wishlist!\" button or \"My Wishlist\" menu.",
		"ru": "Привет, я Wishlist Bot!\n\nЯ могу помочь вам управлять своим личным вишлистом.\n\n" +
			"Я уже создал для вас список желаний! " +
			"Нажмите кнопку \"Открыть вишлист!\" или меню \"My Wishlist\" внизу экрана.",
	},
	"open_wishlist": {
		"en": "Open Wishlist",
		"ru": "Открыть Вишлист",
	},
	"tip_1": {
		"en": "☝️One more tip!\n I can add a Wish to your List by external link!\n" +
			"Just *share* the link with me, and I'll try to create a wish from it.",
		"ru": "☝️Еще один совет!\nЯ могу добавить желание в ваш список по внешней ссылке!\n" +
			"Просто *поделитесь* со мной ссылкой, и я попробую создать из нее желание.",
	},
	"tip_2": {
		"en": " I can add a Wish to your List by external link!\n" +
			"Just *share* the link with me, and I'll try to create a wish from it.",
		"ru": "Я могу добавить желание в ваш список по внешней ссылке!\n" +
			"Просто *поделитесь* со мной ссылкой, и я попробую создать из нее желание!",
	},
	"error_text": {
		"en": "Sorry, I can't do that now. Please, try again later",
		"ru": "Извините, я не могу сейчас этого сделать. Попробуйте позже",
	},
	"wish_added_pattern": {
		"en": "Wish added to your List!\n\n" +
			"*Title:*\n%s\n\n" +
			"*Description:*\n%s\n\n" +
			"Open your new [Wish](%s) to see it.",
		"ru": "Желание добавлено в ваш список!\n\n" +
			"*Название:*\n%s\n\n" +
			"*Описание:*\n%s\n\n" +
			"Откройте свое новое [Желание](%s) чтобы посмотреть его.",
	},
	"empty": {
		"en": "Empty",
		"ru": "Пусто",
	},
}
