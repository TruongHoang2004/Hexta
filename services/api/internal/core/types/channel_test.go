package types

import (
	"slices"
	"testing"
)

func TestChannelConnectionActionsFollowStateAndType(t *testing.T) {
	cases := []struct {
		platform, status string
		want             []string
	}{
		{ChannelPlatformFacebook, ChannelStatusActive, []string{"pause", "disconnect"}},
		{ChannelPlatformFacebook, ChannelStatusPaused, []string{"resume", "disconnect"}},
		{ChannelPlatformFacebook, ChannelStatusAuthExpired, []string{"pause", "disconnect"}},
		{ChannelPlatformTelegram, ChannelStatusActive, []string{"pause", "rotate_token", "disconnect"}},
		{ChannelPlatformTelegram, ChannelStatusPaused, []string{"resume", "rotate_token", "disconnect"}},
		{ChannelPlatformTelegram, ChannelStatusDisconnected, []string{}},
		{"sms", ChannelStatusActive, []string{}},
	}
	for _, c := range cases {
		got := ChannelConnectionActions(ChannelConnection{Platform: c.platform, Status: c.status})
		if !slices.Equal(got, c.want) {
			t.Errorf("%s/%s: got %v want %v", c.platform, c.status, got, c.want)
		}
	}
}

func TestEveryPlatformHasInboxAndExpiryCopy(t *testing.T) {
	for platform, spec := range ChannelPlatforms {
		if spec.ThreadName == "" || spec.AuthExpiredMessage == "" || ChannelThreadNames[platform] != spec.ThreadName {
			t.Errorf("%s is missing display copy", platform)
		}
	}
	if ChannelThreadPlatform("telegram:abc") != ChannelPlatformTelegram || ChannelThreadPlatform("sim:abc") != "" {
		t.Fatal("thread platform")
	}
}
