package mangoplus_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/raf555/mangoplus"
	"github.com/raf555/mangoplus/internal/proto"
)

func TestTitleListService_AllV2_Success(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t, mangoplus.WithSecret("s"))

	mux.HandleFunc("/title_list/allV2", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testFormValues(t, r, values{
			"secret":  "s",
			"os":      "android",
			"os_ver":  mangoplus.DefaultOSVersion,
			"app_ver": mangoplus.DefaultAppVersion,
		})
		testWriteSuccess(t, w, &proto.SuccessResult{
			Data: &proto.SuccessResult_AllTitlesViewV2{
				AllTitlesViewV2: &proto.AllTitlesViewV2{
					AllTitlesGroup: []*proto.AllTitlesGroup{
						{
							TheTitle: "A",
							Titles: []*proto.Title{
								{
									TitleId:           1,
									Name:              "Alpha",
									Author:            "Author A",
									PortraitImageUrl:  "https://img/portrait",
									LandscapeImageUrl: "https://img/landscape",
									Language:          proto.Language_ENGLISH,
								},
							},
							Tags: []*proto.Tag{
								{
									Tag:  "Comedy",
									Slug: "comedy",
								},
							},
							Label:                     &proto.Label{Label: proto.LabelCodes_J_PLUS},
							NextChapterStartTimestamp: 1_700_000_000,
						},
						{
							TheTitle: "B",
							Titles: []*proto.Title{
								{TitleId: 2, Name: "Beta", Language: proto.Language_FRENCH},
								{TitleId: 3, Name: "Gamma", Language: proto.Language_SPANISH},
							},
							Label: &proto.Label{Label: proto.LabelCodes_WJ},
						},
					},
				},
			},
		})
	})

	got, err := client.TitleList.AllV2(t.Context())
	if err != nil {
		t.Fatalf("TitleList.AllV2 returned error: %v", err)
	}

	want := mangoplus.AllTitlesViewV2{
		AllTitlesGroup: []mangoplus.AllTitlesGroupV2{
			{
				TheTitle: "A",
				Titles: []mangoplus.Title{
					{
						TitleID:           1,
						Name:              "Alpha",
						Author:            "Author A",
						PortraitImageURL:  "https://img/portrait",
						LandscapeImageURL: "https://img/landscape",
						Language:          mangoplus.LanguageEnglish,
					},
				},
				Tags: []mangoplus.Tag{
					{
						Name: "Comedy",
						Slug: "comedy",
					},
				},
				Label:                     mangoplus.Label{Code: mangoplus.LabelCodeJPlus},
				NextChapterStartTimestamp: time.Unix(1_700_000_000, 0),
			},
			{
				TheTitle: "B",
				Titles: []mangoplus.Title{
					{TitleID: 2, Name: "Beta", Language: mangoplus.LanguageFrench},
					{TitleID: 3, Name: "Gamma", Language: mangoplus.LanguageSpanish},
				},
				Label: mangoplus.Label{Code: mangoplus.LabelCodeWJ},
			},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("TitleList.AllV2 mismatch (-want +got):\n%s", diff)
	}
}

func TestTitleListService_AllV2_RequestFailed(t *testing.T) {
	t.Parallel()
	testNewRequestAndDoFailure(t, "TitleList.AllV2", func(ctx context.Context, client *mangoplus.Client) error {
		_, err := client.TitleList.AllV2(ctx)
		return err
	})
}

func TestTitleListService_AllV2_NilView(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t)

	mux.HandleFunc("/title_list/allV2", func(w http.ResponseWriter, r *http.Request) {
		testWriteSuccess(t, w, &proto.SuccessResult{})
	})

	if _, err := client.TitleList.AllV2(t.Context()); err == nil {
		t.Error("TitleList.AllV2 returned nil error for empty success body, want error")
	}
}

func TestTitleListService_AllV2_ProtoError(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t)

	mux.HandleFunc("/title_list/allV2", func(w http.ResponseWriter, r *http.Request) {
		testWriteError(t, w, &proto.ErrorResult{
			EnglishPopup: &proto.Popup_OSDefault{Subject: "Not Found", Body: "nope"},
		})
	})

	_, err := client.TitleList.AllV2(t.Context())
	if !errors.Is(err, mangoplus.ErrNotFound) {
		t.Errorf("TitleList.AllV2 error = %v, want ErrNotFound", err)
	}
}

func TestTitleListService_AllV3_Success(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t, mangoplus.WithSecret("s"))

	mux.HandleFunc("/title_list/all_v3", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testFormValues(t, r, values{
			"secret":  "s",
			"os":      "android",
			"os_ver":  mangoplus.DefaultOSVersion,
			"app_ver": mangoplus.DefaultAppVersion,
			"lang":    "eng",
			"clang":   "eng",
			"type":    "completed",
		})
		testWriteSuccess(t, w, &proto.SuccessResult{
			Data: &proto.SuccessResult_SearchView{
				SearchView: &proto.SearchView{
					AllTags: []*proto.Tag{
						{
							Tag:  "Comedy",
							Slug: "comedy",
						},
						{
							Tag:  "Romance",
							Slug: "romance",
						},
					},
					AllTitlesGroup: []*proto.AllTitlesGroup{
						{
							TheTitle: "A",
							Titles: []*proto.Title{
								{
									TitleId: 1,
									Name:    "Alpha",
								},
							},
							Tags: []*proto.Tag{
								{
									Tag:  "Romance",
									Slug: "romance",
								},
							},
							Label: &proto.Label{Label: proto.LabelCodes_J_PLUS},
						},
					},
				},
			},
		})
	})

	got, err := client.TitleList.AllV3(t.Context(), mangoplus.TitleListTypeCompleted)
	if err != nil {
		t.Fatalf("TitleList.AllV3 returned error: %v", err)
	}

	want := mangoplus.SearchView{
		AllTags: []mangoplus.Tag{
			{
				Name: "Comedy",
				Slug: "comedy",
			},
			{
				Name: "Romance",
				Slug: "romance",
			},
		},
		AllTitlesGroup: []mangoplus.AllTitlesGroupV2{
			{
				TheTitle: "A",
				Titles: []mangoplus.Title{
					{
						TitleID:  1,
						Name:     "Alpha",
						Language: mangoplus.LanguageEnglish,
					},
				},
				Tags: []mangoplus.Tag{
					{
						Name: "Romance",
						Slug: "romance",
					},
				},
				Label: mangoplus.Label{Code: mangoplus.LabelCodeJPlus},
			},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("TitleList.AllV3 mismatch (-want +got):\n%s", diff)
	}
}

func TestTitleListService_AllV3_RequestFailed(t *testing.T) {
	t.Parallel()
	testNewRequestAndDoFailure(t, "TitleList.AllV3", func(ctx context.Context, client *mangoplus.Client) error {
		_, err := client.TitleList.AllV3(ctx, mangoplus.TitleListTypeSerializing)
		return err
	})
}

func TestTitleListService_AllV3_NilView(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t)

	mux.HandleFunc("/title_list/all_v3", func(w http.ResponseWriter, r *http.Request) {
		testWriteSuccess(t, w, &proto.SuccessResult{})
	})

	if _, err := client.TitleList.AllV3(t.Context(), mangoplus.TitleListTypeSerializing); err == nil {
		t.Error("TitleList.AllV3 returned nil error for empty success body, want error")
	}
}

func TestTitleListService_AllV3_ProtoError(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t)

	mux.HandleFunc("/title_list/all_v3", func(w http.ResponseWriter, r *http.Request) {
		testWriteError(t, w, &proto.ErrorResult{
			EnglishPopup: &proto.Popup_OSDefault{
				Subject: "Not Found",
				Body:    "nope",
			},
		})
	})

	_, err := client.TitleList.AllV3(t.Context(), mangoplus.TitleListTypeOneShot)
	if !errors.Is(err, mangoplus.ErrNotFound) {
		t.Errorf("TitleList.AllV3 error = %v, want ErrNotFound", err)
	}
}

func TestTitleListService_Search_Success(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t, mangoplus.WithSecret("s"))

	mux.HandleFunc("/title_list/search", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testFormValues(t, r, values{
			"secret":  "s",
			"os":      "android",
			"os_ver":  mangoplus.DefaultOSVersion,
			"app_ver": mangoplus.DefaultAppVersion,
			"lang":    "eng",
			"clang":   "eng",
		})
		testWriteSuccess(t, w, &proto.SuccessResult{
			Data: &proto.SuccessResult_SearchView{
				SearchView: &proto.SearchView{
					TopSearchBanners: []*proto.Banner{
						{
							Id:       3,
							ImageUrl: "https://img/top",
							Action: &proto.TransitionAction{
								Url: "mangaplus://open/detail?id=2",
							},
						},
					},
					Contents: []*proto.SearchView_Contents{
						{
							Banner: &proto.Banner{ImageUrl: "https://img/banner"},
						},
						{
							ChapterPagesList: &proto.ChapterPageList{
								ListName: "Manga Previews",
								ChapterPages: []*proto.ChapterPages{
									{
										ChapterId:        10,
										TitleId:          1,
										Name:             "Alpha",
										Author:           "Author A",
										FavoriteImageUrl: "https://img/fav",
										Pages: []*proto.Page{
											{
												Data: &proto.Page_MangaPage_{
													MangaPage: &proto.Page_MangaPage{
														ImageUrl: "https://img/1",
														Width:    740,
														Height:   1080,
													},
												},
											},
											{
												Data: &proto.Page_LastPage_{LastPage: &proto.Page_LastPage{}},
											},
										},
									},
								},
							},
						},
						{
							TitleList: &proto.TitleList{
								ListName: "JUMP PLUS",
								FeaturedTitles: []*proto.Title{
									{
										TitleId: 2,
										Name:    "Beta",
									},
								},
							},
						},
					},
				},
			},
		})
	})

	got, err := client.TitleList.Search(t.Context(), mangoplus.LanguageEnglish)
	if err != nil {
		t.Fatalf("TitleList.Search returned error: %v", err)
	}

	want := mangoplus.SearchView{
		TopBanners: []mangoplus.Banner{
			{
				ID:        3,
				ImageURL:  "https://img/top",
				ActionURL: "mangaplus://open/detail?id=2",
			},
		},
		Contents: []mangoplus.SearchViewContents{
			{
				Banner: &mangoplus.Banner{ImageURL: "https://img/banner"},
			},
			{
				ChapterPageList: &mangoplus.ChapterPageList{
					ListName: "Manga Previews",
					ChapterPages: []mangoplus.ChapterPages{
						{
							ChapterID:        10,
							TitleID:          1,
							Name:             "Alpha",
							Author:           "Author A",
							FavoriteImageURL: "https://img/fav",
							Pages: []mangoplus.MangaPage{
								{
									ImageURL: "https://img/1",
									Width:    740,
									Height:   1080,
									Type:     mangoplus.PageTypeSingle,
								},
							},
						},
					},
				},
			},
			{
				TitleList: &mangoplus.TitleList{
					ListName: "JUMP PLUS",
					FeaturedTitles: []mangoplus.Title{
						{
							TitleID:  2,
							Name:     "Beta",
							Language: mangoplus.LanguageEnglish,
						},
					},
				},
			},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("TitleList.Search mismatch (-want +got):\n%s", diff)
	}
}

func TestTitleListService_Search_UnsupportedLanguage(t *testing.T) {
	t.Parallel()
	client, _, _ := setup(t)

	if _, err := client.TitleList.Search(t.Context(), "KLINGON"); err == nil {
		t.Error("TitleList.Search returned nil error for an unsupported language, want error")
	}
}

func TestTitleListService_Search_RequestFailed(t *testing.T) {
	t.Parallel()
	testNewRequestAndDoFailure(t, "TitleList.Search", func(ctx context.Context, client *mangoplus.Client) error {
		_, err := client.TitleList.Search(ctx, mangoplus.LanguageEnglish)
		return err
	})
}

func TestTitleListService_Search_NilView(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t)

	mux.HandleFunc("/title_list/search", func(w http.ResponseWriter, r *http.Request) {
		testWriteSuccess(t, w, &proto.SuccessResult{})
	})

	if _, err := client.TitleList.Search(t.Context(), mangoplus.LanguageEnglish); err == nil {
		t.Error("TitleList.Search returned nil error for empty success body, want error")
	}
}

func TestTitleListService_RankingV2_Success(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t, mangoplus.WithSecret("s"))

	mux.HandleFunc("/title_list/rankingV2", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testFormValues(t, r, values{
			"secret":  "s",
			"os":      "android",
			"os_ver":  mangoplus.DefaultOSVersion,
			"app_ver": mangoplus.DefaultAppVersion,
			"lang":    "eng",
			"clang":   "ind",
			"type":    "trending",
		})
		testWriteSuccess(t, w, &proto.SuccessResult{
			Data: &proto.SuccessResult_TitleRankingViewV2{
				TitleRankingViewV2: &proto.TitleRankingViewV2{
					RankingBanners: []*proto.Banner{
						{
							Id:       5,
							ImageUrl: "https://img/ranking",
							Action: &proto.TransitionAction{
								Url: "mangaplus://open/detail?id=1",
							},
						},
					},
					UpdatedTimeStamp: 1_700_000_000,
					RankedTitles: []*proto.TitleRankingGroup{
						{
							OriginalTitleId: 100,
							Titles: []*proto.Title{
								{
									TitleId:  1,
									Name:     "Alpha",
									Language: proto.Language_INDONESIAN,
								},
							},
							Score: 4321,
						},
					},
				},
			},
		})
	})

	got, err := client.TitleList.RankingV2(t.Context(), mangoplus.RankingTypeTrending, mangoplus.LanguageIndonesian)
	if err != nil {
		t.Fatalf("TitleList.RankingV2 returned error: %v", err)
	}

	want := mangoplus.TitleRankingViewV2{
		Banners: []mangoplus.Banner{
			{
				ID:        5,
				ImageURL:  "https://img/ranking",
				ActionURL: "mangaplus://open/detail?id=1",
			},
		},
		UpdatedAt: time.Unix(1_700_000_000, 0),
		RankedTitles: []mangoplus.TitleRankingGroup{
			{
				OriginalTitleID: 100,
				Titles: []mangoplus.Title{
					{
						TitleID:  1,
						Name:     "Alpha",
						Language: mangoplus.LanguageIndonesian,
					},
				},
				Score: 4321,
			},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("TitleList.RankingV2 mismatch (-want +got):\n%s", diff)
	}
}

func TestTitleListService_RankingV2_UnsupportedLanguage(t *testing.T) {
	t.Parallel()
	client, _, _ := setup(t)

	if _, err := client.TitleList.RankingV2(t.Context(), mangoplus.RankingTypeHottest, "KLINGON"); err == nil {
		t.Error("TitleList.RankingV2 returned nil error for an unsupported language, want error")
	}
}

func TestTitleListService_RankingV2_RequestFailed(t *testing.T) {
	t.Parallel()
	testNewRequestAndDoFailure(t, "TitleList.RankingV2", func(ctx context.Context, client *mangoplus.Client) error {
		_, err := client.TitleList.RankingV2(ctx, mangoplus.RankingTypeHottest, mangoplus.LanguageEnglish)
		return err
	})
}

func TestTitleListService_RankingV2_NilView(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t)

	mux.HandleFunc("/title_list/rankingV2", func(w http.ResponseWriter, r *http.Request) {
		testWriteSuccess(t, w, &proto.SuccessResult{})
	})

	if _, err := client.TitleList.RankingV2(t.Context(), mangoplus.RankingTypeHottest, mangoplus.LanguageEnglish); err == nil {
		t.Error("TitleList.RankingV2 returned nil error for empty success body, want error")
	}
}
