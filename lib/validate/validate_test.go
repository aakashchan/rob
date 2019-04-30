package validate

import "testing"

func TestFeed(t *testing.T) {
	type params struct {
		LastSync string
		MascotId string
		Flag     string
	}

	var invalidParams = []params{
		{"abc", "1", "1"},
		{"", "1", "1"},
		{"1", "abc", "1"},
		{"1", "", "1"},
		{"1", "1", "abc"},
		{"1", "1", ""},
		{"1", "1", "0"},
		{"1", "1", "4"},
	}

	for _, p := range invalidParams {
		_, _, _, err := Feed(p.LastSync, p.MascotId, p.Flag)
		if err == nil {
			t.Errorf("Feed validate failed. Expected=error but received "+
				"nil for values LastSync=%s MascotId=%s Flag=%s", p.LastSync, p.MascotId, p.Flag)
		}
	}

	type expected struct {
		params
		ExpectedLastSync int64
		ExpectedMascotId int
		ExpectedFlag     int
	}

	var validParams = []expected{
		{params{"1", "1", "1"}, 1, 1, 1},
		{params{"-1", "1", "2"}, -1, 1, 2},
		{params{"1000000000000000", "1", "3"}, 1000000000000000, 1, 3},
	}

	for _, p := range validParams {
		ls, ms, f, err := Feed(p.LastSync, p.MascotId, p.Flag)
		if err != nil {
			t.Errorf("Feed validate failed. Expected=nil but received '%s'", err.Error())
		}
		if ls != p.ExpectedLastSync {
			t.Errorf("Feed validate failed. ExpectedLastSync=%d but received %d",
				p.ExpectedLastSync, ls)
		}
		if ms != p.ExpectedMascotId {
			t.Errorf("Feed validate failed. ExpectedMascotId=%d but received %d",
				p.ExpectedMascotId, ms)
		}
		if f != p.ExpectedFlag {
			t.Errorf("Feed validate failed. ExpectedFlag=%d but received %d",
				p.ExpectedFlag, f)
		}
	}
}

func TestCreatePost(t *testing.T) {
	type Param struct {
		CardType      string
		Src           string
		DpSrc         string
		Title         string
		Desc          string
		ButtonText    string
		Url           string
		Icon          string
		GradientStart string
		GradientEnd   string
		ChildPosts    []string
	}

	// TODO: Finish this once the init-creds issue is solved
	invalidParams := []Param{
		{},
	}

	for _, p := range invalidParams {
		_, err := CreatePost(p.CardType, p.Src, p.DpSrc, p.Title,
			p.Desc, p.ButtonText, p.Url, p.Icon, p.GradientStart, p.GradientEnd, p.ChildPosts)
		if err == nil {
			t.Errorf("Expected CreatePost validate to fail but it passed for Params=%+v", p)
		}
	}
}
