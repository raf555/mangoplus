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

func TestHomeService_ViewV6_Success(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t, mangoplus.WithSecret("s"))

	mux.HandleFunc("/home_v6", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testFormValues(t, r, values{
			"secret":      "s",
			"os":          "android",
			"os_ver":      mangoplus.DefaultOSVersion,
			"app_ver":     mangoplus.DefaultAppVersion,
			"lang":        "eng",
			"clang":       "eng",
			"viewer_mode": "horizontal",
		})

		groupPb := &proto.OriginalTitleGroup{
			TheTitle:      "Alpha",
			ChapterNumber: "#012",
			Titles: []*proto.UpdatedTitle{
				{
					Title: &proto.Title{
						TitleId: 1,
						Name:    "Alpha",
					},
					ChapterId:       10,
					ChapterName:     "#012",
					ChapterSubTitle: "Chapter 12",
					IsLatest:        true,
				},
			},
			ViewCount:         1234,
			TitleUpdateStatus: proto.TitleUpdateStatus_UP,
			ChapterStartTime:  1_700_000_000,
		}

		testWriteSuccess(t, w, &proto.SuccessResult{
			Data: &proto.SuccessResult_HomeViewV6{
				HomeViewV6: &proto.HomeViewV6{
					Sections: []*proto.HomeViewV6_Sections{
						{
							Section: &proto.HomeViewV6_Sections_WeeklySection{
								WeeklySection: &proto.HomeViewV6_WeeklySection{
									Contents: []*proto.HomeViewV6_WeeklySection_WeeklyContent{
										{
											IsUpdated:        true,
											UpdatedTimeStamp: 1_700_000_000,
											ContentItems: []*proto.HomeViewV6_WeeklySection_ContentItem{
												{
													Content: &proto.HomeViewV6_WeeklySection_ContentItem_MvBanner{
														MvBanner: &proto.HomeViewV6_WeeklySection_MVBanner{
															ImageUrl:    "https://img/mv",
															TitleGroups: groupPb,
														},
													},
												},
												{
													Content: &proto.HomeViewV6_WeeklySection_ContentItem_TitleGroup{
														TitleGroup: &proto.HomeViewV6_WeeklySection_TitleGroup{
															TitleGroups: []*proto.OriginalTitleGroup{groupPb},
														},
													},
												},
												{
													Content: &proto.HomeViewV6_WeeklySection_ContentItem_CarouselBanners{
														CarouselBanners: &proto.HomeViewV6_WeeklySection_CarouselBanner{
															Banners: []*proto.Banner{
																{
																	Id:       7,
																	ImageUrl: "https://img/carousel",
																	Action: &proto.TransitionAction{
																		Url: "mangaplus://open/detail?id=1",
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						{
							Section: &proto.HomeViewV6_Sections_RankingSection{
								RankingSection: &proto.HomeViewV6_RankingSection{
									RankingTabs: []*proto.HomeViewV6_RankingSection_RankingTab{
										{
											TabType: proto.RankingTabType_TRENDING,
											RankedTitles: []*proto.TitleRankingGroup{
												{
													OriginalTitleId: 100,
													Titles: []*proto.Title{
														{
															TitleId: 1,
															Name:    "Alpha",
														},
													},
													Score: 999,
												},
											},
										},
									},
								},
							},
						},
						{
							Section: &proto.HomeViewV6_Sections_BannerSection{
								BannerSection: &proto.HomeViewV6_BannerSection{
									Banners: []*proto.Banner{
										{
											Id:       8,
											ImageUrl: "https://img/banner",
											Width:    1000,
											Height:   300,
											Action: &proto.TransitionAction{
												Method: proto.TransitionAction_EXTERNAL,
												Url:    "https://example.com",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		})
	})

	got, err := client.Home.ViewV6(t.Context())
	if err != nil {
		t.Fatalf("Home.ViewV6 returned error: %v", err)
	}

	group := mangoplus.OriginalTitleGroup{
		TheTitle:      "Alpha",
		ChapterNumber: "#012",
		Titles: []mangoplus.UpdatedTitle{
			{
				Title: mangoplus.Title{
					TitleID:  1,
					Name:     "Alpha",
					Language: mangoplus.LanguageEnglish,
				},
				ChapterID:       10,
				ChapterName:     "#012",
				ChapterSubTitle: "Chapter 12",
				IsLatest:        true,
			},
		},
		ViewCount:        1234,
		UpdateStatus:     mangoplus.TitleUpdateStatusUp,
		ChapterStartTime: time.Unix(1_700_000_000, 0),
	}
	want := mangoplus.HomeViewV6{
		Sections: []mangoplus.HomeSection{
			{
				Weekly: &mangoplus.WeeklySection{
					Contents: []mangoplus.WeeklyContent{
						{
							IsUpdated: true,
							UpdatedAt: time.Unix(1_700_000_000, 0),
							Items: []mangoplus.WeeklyContentItem{
								{
									MVBanner: &mangoplus.MVBanner{
										ImageURL:   "https://img/mv",
										TitleGroup: group,
									},
								},
								{
									TitleGroups: []mangoplus.OriginalTitleGroup{group},
								},
								{
									CarouselBanners: []mangoplus.Banner{
										{
											ID:        7,
											ImageURL:  "https://img/carousel",
											ActionURL: "mangaplus://open/detail?id=1",
										},
									},
								},
							},
						},
					},
				},
			},
			{
				Ranking: &mangoplus.RankingSection{
					Tabs: []mangoplus.RankingTab{
						{
							Type: mangoplus.RankingTypeTrending,
							RankedTitles: []mangoplus.TitleRankingGroup{
								{
									OriginalTitleID: 100,
									Titles: []mangoplus.Title{
										{
											TitleID:  1,
											Name:     "Alpha",
											Language: mangoplus.LanguageEnglish,
										},
									},
									Score: 999,
								},
							},
						},
					},
				},
			},
			{
				Banners: &mangoplus.BannerSection{
					Banners: []mangoplus.Banner{
						{
							ID:        8,
							ImageURL:  "https://img/banner",
							Width:     1000,
							Height:    300,
							ActionURL: "https://example.com",
						},
					},
				},
			},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Home.ViewV6 mismatch (-want +got):\n%s", diff)
	}
}

func TestHomeService_ViewV6_RequestFailed(t *testing.T) {
	t.Parallel()
	testNewRequestAndDoFailure(t, "Home.ViewV6", func(ctx context.Context, client *mangoplus.Client) error {
		_, err := client.Home.ViewV6(ctx)
		return err
	})
}

func TestHomeService_ViewV6_NilView(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t)

	mux.HandleFunc("/home_v6", func(w http.ResponseWriter, r *http.Request) {
		testWriteSuccess(t, w, &proto.SuccessResult{})
	})

	if _, err := client.Home.ViewV6(t.Context()); err == nil {
		t.Error("Home.ViewV6 returned nil error for empty success body, want error")
	}
}

func TestHomeService_ViewV6_ProtoError(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t)

	mux.HandleFunc("/home_v6", func(w http.ResponseWriter, r *http.Request) {
		testWriteError(t, w, &proto.ErrorResult{
			EnglishPopup: &proto.Popup_OSDefault{
				Subject: "Not Found",
				Body:    "nope",
			},
		})
	})

	_, err := client.Home.ViewV6(t.Context())
	if !errors.Is(err, mangoplus.ErrNotFound) {
		t.Errorf("Home.ViewV6 error = %v, want ErrNotFound", err)
	}
}

func TestHomeService_PreviewV2_Success(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t, mangoplus.WithSecret("s"))

	mux.HandleFunc("/preview_v2", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testFormValues(t, r, values{
			"secret":      "s",
			"os":          "android",
			"os_ver":      mangoplus.DefaultOSVersion,
			"app_ver":     mangoplus.DefaultAppVersion,
			"lang":        "eng",
			"clang":       "ind",
			"viewer_mode": "horizontal",
		})
		testWriteSuccess(t, w, &proto.SuccessResult{
			Data: &proto.SuccessResult_HomePreviewViewV2{
				HomePreviewViewV2: &proto.HomePreviewViewV2{
					PreviewSections: []*proto.HomePreviewViewV2_PreviewSection{
						{
							SectionId:   1,
							SectionName: "Updates Preview",
							SectionType: proto.HomePreviewViewV2_PreviewSection_UPDATE,
							PreviewTabs: []*proto.HomePreviewViewV2_PreviewTab{
								{
									Language: proto.Language_INDONESIAN,
									Titles: []*proto.HomePreviewViewV2_PreviewTitle{
										{
											ChapterId:    10,
											TitleName:    "Alpha",
											Author:       "Author A",
											ThumbnailUrl: "https://img/thumb",
											Pages: []*proto.HomePreviewViewV2_PreviewPage{
												{
													ImageUrl: "https://img/1",
													Width:    739,
													Height:   1080,
												},
											},
										},
									},
								},
							},
						},
						{
							SectionId:   2,
							SectionName: "Chapter 1 Preview",
							SectionType: proto.HomePreviewViewV2_PreviewSection_CHAPTER,
						},
					},
				},
			},
		})
	})

	got, err := client.Home.PreviewV2(t.Context(), mangoplus.LanguageIndonesian)
	if err != nil {
		t.Fatalf("Home.PreviewV2 returned error: %v", err)
	}

	want := mangoplus.HomePreviewViewV2{
		Sections: []mangoplus.PreviewSection{
			{
				ID:   1,
				Name: "Updates Preview",
				Type: mangoplus.PreviewSectionTypeUpdate,
				Tabs: []mangoplus.PreviewTab{
					{
						Language: mangoplus.LanguageIndonesian,
						Titles: []mangoplus.PreviewTitle{
							{
								ChapterID:    10,
								TitleName:    "Alpha",
								Author:       "Author A",
								ThumbnailURL: "https://img/thumb",
								Pages: []mangoplus.PreviewPage{
									{
										ImageURL: "https://img/1",
										Width:    739,
										Height:   1080,
									},
								},
							},
						},
					},
				},
			},
			{
				ID:   2,
				Name: "Chapter 1 Preview",
				Type: mangoplus.PreviewSectionTypeChapter,
			},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Home.PreviewV2 mismatch (-want +got):\n%s", diff)
	}
}

func TestHomeService_PreviewV2_UnsupportedLanguage(t *testing.T) {
	t.Parallel()
	client, _, _ := setup(t)

	if _, err := client.Home.PreviewV2(t.Context(), "KLINGON"); err == nil {
		t.Error("Home.PreviewV2 returned nil error for an unsupported language, want error")
	}
}

func TestHomeService_PreviewV2_RequestFailed(t *testing.T) {
	t.Parallel()
	testNewRequestAndDoFailure(t, "Home.PreviewV2", func(ctx context.Context, client *mangoplus.Client) error {
		_, err := client.Home.PreviewV2(ctx, mangoplus.LanguageEnglish)
		return err
	})
}

func TestHomeService_PreviewV2_NilView(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t)

	mux.HandleFunc("/preview_v2", func(w http.ResponseWriter, r *http.Request) {
		testWriteSuccess(t, w, &proto.SuccessResult{})
	})

	if _, err := client.Home.PreviewV2(t.Context(), mangoplus.LanguageEnglish); err == nil {
		t.Error("Home.PreviewV2 returned nil error for empty success body, want error")
	}
}
