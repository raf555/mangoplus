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

// func toMangaPlusLang(lang Language) string {
// 	switch lang {
// 	case LanguageEnglish:
// 		return "eng"
// 	case LanguageSpanish:
// 		return "esp"
// 	case LanguageFrench:
// 		return "fra"
// 	case LanguageIndonesian:
// 		return "ind"
// 	case LanguagePortugueseBR:
// 		return "ptb"
// 	case LanguageRussian:
// 		return "rus"
// 	case LanguageThai:
// 		return "tha"
// 	case LanguageVietnamese:
// 		return "vie"
// 	case LanguageGerman:
// 		return "deu"
// 	default:
// 		return ""
// 	}
// }
