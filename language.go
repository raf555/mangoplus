package mangoplus

import "github.com/raf555/mangoplus/internal/proto"

type Language string

const (
	LanguageEnglish      Language = "ENGLISH"
	LanguageSpanish      Language = "SPANISH"
	LanguageFrench       Language = "FRENCH"
	LanguageIndonesian   Language = "INDONESIAN"
	LanguagePortugueseBR Language = "PORTUGUESE_BR"
	LanguageRussian      Language = "RUSSIAN"
	LanguageThai         Language = "THAI"
	LanguageVietnamese   Language = "VIETNAMESE"
	LanguageGerman       Language = "GERMAN"
)

func languageFromProto(pb proto.Language) Language {
	val, ok := proto.Language_name[int32(pb)]
	if !ok {
		return ""
	}
	return Language(val)
}

func languageFromMangaPlusLang(lang string) Language {
	switch lang {
	case "eng":
		return LanguageEnglish
	case "esp":
		return LanguageSpanish
	case "fra":
		return LanguageFrench
	case "ind":
		return LanguageIndonesian
	case "ptb":
		return LanguagePortugueseBR
	case "rus":
		return LanguageRussian
	case "tha":
		return LanguageThai
	case "vie":
		return LanguageVietnamese
	case "deu":
		return LanguageGerman
	default:
		return ""
	}
}
