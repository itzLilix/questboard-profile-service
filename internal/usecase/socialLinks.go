package usecase

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/itzLilix/questboard-shared/dtos"
)

var platformHosts = map[string][]string{
	"telegram":  {"t.me", "telegram.me"},
	"discord":   {"discord.gg", "discord.com"},
	"twitch":    {"twitch.tv"},
	"youtube":   {"youtube.com", "youtu.be"},
	"x":         {"twitter.com", "x.com"},
	"patreon":   {"patreon.com"},
	"vkontakte": {"vk.com"},
	"odnoklassniki": {"ok.ru"},
	"facebook":  {"facebook.com"},
	"instagram": {"instagram.com"},
	"tiktok":    {"tiktok.com"},
	"bluesky":   {"bsky.app"},
	"whatsapp":  {"wa.me", "whatsapp.com"},
	"pinterest": {"pinterest.com", "pinterest.ru"},
	"kofi":      {"ko-fi.com"},
	"roll20":    {"roll20.net"},
	"reddit":    {"reddit.com"},
	"snapchat":  {"snapchat.com"},
}

func validateAndNormalize(link dtos.Link) (dtos.Link, error) {
	u, err := url.Parse(link.URL)
	if err != nil || u.Host == "" {
		return dtos.Link{}, ErrInvalidURL
	}

	host := strings.ToLower(u.Host)
	host = strings.TrimPrefix(host, "www.")
	host = strings.TrimPrefix(host, "m.")
	host = strings.TrimPrefix(host, "old.")
	host = strings.TrimPrefix(host, "new.")

	allowedHosts, ok := platformHosts[link.Type]
	if !ok {
		return dtos.Link{}, fmt.Errorf("unknown platform")
	}

	valid := false
	for _, h := range allowedHosts {
		if host == h {
			valid = true
			break
		}
	}
	if !valid {
		return dtos.Link{}, fmt.Errorf("URL does not match platform")
	}

	u.Scheme = "https"
	u.RawQuery = ""
	u.Fragment = ""

	return dtos.Link{
		Type: link.Type,
		URL:  u.String(),
	}, nil
}

