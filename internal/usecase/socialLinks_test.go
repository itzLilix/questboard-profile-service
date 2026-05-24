package usecase

import (
	"testing"

	"github.com/itzLilix/questboard-shared/dtos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateAndNormalize_AllPlatforms(t *testing.T) {
	cases := []struct {
		platform string
		input    string
		want     string
	}{
		{"telegram", "https://t.me/foo", "https://t.me/foo"},
		{"telegram", "http://telegram.me/bar", "https://telegram.me/bar"},
		{"discord", "https://discord.gg/abc", "https://discord.gg/abc"},
		{"discord", "https://discord.com/users/1", "https://discord.com/users/1"},
		{"twitch", "https://twitch.tv/streamer", "https://twitch.tv/streamer"},
		{"youtube", "https://youtube.com/@chan", "https://youtube.com/@chan"},
		{"youtube", "https://youtu.be/abc", "https://youtu.be/abc"},
		{"x", "https://x.com/foo", "https://x.com/foo"},
		{"x", "https://twitter.com/foo", "https://twitter.com/foo"},
		{"patreon", "https://patreon.com/foo", "https://patreon.com/foo"},
		{"vkontakte", "https://vk.com/id1", "https://vk.com/id1"},
		{"odnoklassniki", "https://ok.ru/profile/1", "https://ok.ru/profile/1"},
		{"facebook", "https://facebook.com/foo", "https://facebook.com/foo"},
		{"instagram", "https://instagram.com/foo", "https://instagram.com/foo"},
		{"tiktok", "https://tiktok.com/@foo", "https://tiktok.com/@foo"},
		{"bluesky", "https://bsky.app/profile/foo", "https://bsky.app/profile/foo"},
		{"whatsapp", "https://wa.me/123", "https://wa.me/123"},
		{"whatsapp", "https://whatsapp.com/foo", "https://whatsapp.com/foo"},
		{"pinterest", "https://pinterest.com/foo", "https://pinterest.com/foo"},
		{"pinterest", "https://pinterest.ru/foo", "https://pinterest.ru/foo"},
		{"kofi", "https://ko-fi.com/foo", "https://ko-fi.com/foo"},
		{"roll20", "https://roll20.net/users/1", "https://roll20.net/users/1"},
		{"reddit", "https://reddit.com/r/golang", "https://reddit.com/r/golang"},
		{"snapchat", "https://snapchat.com/add/foo", "https://snapchat.com/add/foo"},
	}
	for _, tc := range cases {
		t.Run(tc.platform+"/"+tc.input, func(t *testing.T) {
			out, err := validateAndNormalize(dtos.Link{Type: tc.platform, URL: tc.input})
			require.NoError(t, err)
			assert.Equal(t, tc.platform, out.Type)
			assert.Equal(t, tc.want, out.URL)
		})
	}
}

func TestValidateAndNormalize_HostPrefixesStripped(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"www prefix", "https://www.twitch.tv/foo"},
		{"m prefix", "https://m.facebook.com/foo"},
		{"old prefix", "https://old.reddit.com/r/golang"},
		{"new prefix", "https://new.reddit.com/r/golang"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			platform := "twitch"
			switch tc.name {
			case "m prefix":
				platform = "facebook"
			case "old prefix", "new prefix":
				platform = "reddit"
			}
			out, err := validateAndNormalize(dtos.Link{Type: platform, URL: tc.url})
			require.NoError(t, err)
			assert.Equal(t, platform, out.Type)
		})
	}
}

func TestValidateAndNormalize_NormalizesSchemeAndStripsQueryFragment(t *testing.T) {
	out, err := validateAndNormalize(dtos.Link{
		Type: "x",
		URL:  "http://x.com/foo?utm=1#frag",
	})
	require.NoError(t, err)
	assert.Equal(t, "https://x.com/foo", out.URL)
}

func TestValidateAndNormalize_Errors(t *testing.T) {
	tests := []struct {
		name string
		link dtos.Link
	}{
		{"empty url", dtos.Link{Type: "telegram", URL: ""}},
		{"unparseable url", dtos.Link{Type: "telegram", URL: "://"}},
		{"unknown platform", dtos.Link{Type: "myspace", URL: "https://myspace.com/foo"}},
		{"host mismatch", dtos.Link{Type: "telegram", URL: "https://example.com/foo"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := validateAndNormalize(tc.link)
			assert.ErrorIs(t, err, ErrInvalidData)
		})
	}
}
