package mangoplus_test

import (
	"testing"

	"github.com/raf555/mangoplus"
)

func TestLanguageToLang(t *testing.T) {
	t.Parallel()

	tests := []struct {
		language mangoplus.Language
		want     mangoplus.Lang
		wantOK   bool
	}{
		{
			language: mangoplus.LanguageEnglish,
			want:     mangoplus.LangEnglish,
			wantOK:   true,
		},
		{
			language: mangoplus.LanguageSpanish,
			want:     mangoplus.LangSpanish,
			wantOK:   true,
		},
		{
			language: mangoplus.LanguageFrench,
			want:     mangoplus.LangFrench,
			wantOK:   true,
		},
		{
			language: mangoplus.LanguageIndonesian,
			want:     mangoplus.LangIndonesian,
			wantOK:   true,
		},
		{
			language: mangoplus.LanguagePortugueseBR,
			want:     mangoplus.LangPortugueseBR,
			wantOK:   true,
		},
		{
			language: mangoplus.LanguageRussian,
			want:     mangoplus.LangRussian,
			wantOK:   true,
		},
		{
			language: mangoplus.LanguageThai,
			want:     mangoplus.LangThai,
			wantOK:   true,
		},
		{
			language: mangoplus.LanguageVietnamese,
			want:     mangoplus.LangVietnamese,
			wantOK:   true,
		},
		{
			language: mangoplus.LanguageGerman,
			want:     mangoplus.LangGerman,
			wantOK:   true,
		},
		{
			language: "KLINGON",
		},
		{
			language: "",
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.language), func(t *testing.T) {
			t.Parallel()

			got, ok := mangoplus.LanguageToLang(tt.language)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("LanguageToLang(%q) = %q, %v, want %q, %v", tt.language, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestLangToLanguage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		lang   mangoplus.Lang
		want   mangoplus.Language
		wantOK bool
	}{
		{
			lang:   mangoplus.LangEnglish,
			want:   mangoplus.LanguageEnglish,
			wantOK: true,
		},
		{
			lang:   mangoplus.LangSpanish,
			want:   mangoplus.LanguageSpanish,
			wantOK: true,
		},
		{
			lang:   mangoplus.LangFrench,
			want:   mangoplus.LanguageFrench,
			wantOK: true,
		},
		{
			lang:   mangoplus.LangIndonesian,
			want:   mangoplus.LanguageIndonesian,
			wantOK: true,
		},
		{
			lang:   mangoplus.LangPortugueseBR,
			want:   mangoplus.LanguagePortugueseBR,
			wantOK: true,
		},
		{
			lang:   mangoplus.LangRussian,
			want:   mangoplus.LanguageRussian,
			wantOK: true,
		},
		{
			lang:   mangoplus.LangThai,
			want:   mangoplus.LanguageThai,
			wantOK: true,
		},
		{
			lang:   mangoplus.LangVietnamese,
			want:   mangoplus.LanguageVietnamese,
			wantOK: true,
		},
		{
			lang:   mangoplus.LangGerman,
			want:   mangoplus.LanguageGerman,
			wantOK: true,
		},
		{
			lang: "jpn",
		},
		{
			lang: "",
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.lang), func(t *testing.T) {
			t.Parallel()

			got, ok := mangoplus.LangToLanguage(tt.lang)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("LangToLanguage(%q) = %q, %v, want %q, %v", tt.lang, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}
