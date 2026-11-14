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

type Lang string

const (
	LangEnglish      Lang = "eng"
	LangSpanish      Lang = "esp"
	LangFrench       Lang = "fra"
	LangIndonesian   Lang = "ind"
	LangPortugueseBR Lang = "ptb"
	LangRussian      Lang = "rus"
	LangThai         Lang = "tha"
	LangVietnamese   Lang = "vie"
	LangGerman       Lang = "deu"
)

func LangToLanguage(lang Lang) (Language, bool) {
	switch lang {
	case LangEnglish:
		return LanguageEnglish, true
	case LangSpanish:
		return LanguageSpanish, true
	case LangFrench:
		return LanguageFrench, true
	case LangIndonesian:
		return LanguageIndonesian, true
	case LangPortugueseBR:
		return LanguagePortugueseBR, true
	case LangRussian:
		return LanguageRussian, true
	case LangThai:
		return LanguageThai, true
	case LangVietnamese:
		return LanguageVietnamese, true
	case LangGerman:
		return LanguageGerman, true
	default:
		return "", false
	}
}

func LanguageToLang(lang Language) (Lang, bool) {
	switch lang {
	case LanguageEnglish:
		return LangEnglish, true
	case LanguageSpanish:
		return LangSpanish, true
	case LanguageFrench:
		return LangFrench, true
	case LanguageIndonesian:
		return LangIndonesian, true
	case LanguagePortugueseBR:
		return LangPortugueseBR, true
	case LanguageRussian:
		return LangRussian, true
	case LanguageThai:
		return LangThai, true
	case LanguageVietnamese:
		return LangVietnamese, true
	case LanguageGerman:
		return LangGerman, true
	default:
		return "", false
	}
}
