package service

import (
	"context"
	"testing"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
)

// A tenant holds any number of connections of any type; they are listed and managed through one resource.
func TestConnectionsAcrossTypesAreListedAndManagedTogether(t *testing.T) {
	f, s, _ := telegramFixture(t)
	ctx := context.Background()
	pages, _ := connectPages(t, s, f.tenant, f.owner, "page-a", "page-b")
	bot, err := s.ConnectTelegramBot(ctx, f.tenant, f.owner, tgToken)
	if err != nil {
		t.Fatal(err)
	}

	all, err := s.ListConnections(ctx, f.tenant, f.sale, "")
	if err != nil || len(all) != 3 {
		t.Fatal("members see every connection of every type", err, len(all))
	}
	if fb, _ := s.ListConnections(ctx, f.tenant, f.sale, types.ChannelPlatformFacebook); len(fb) != 2 {
		t.Fatal("facebook filter", len(fb))
	}
	if tg, _ := s.ListConnections(ctx, f.tenant, f.sale, types.ChannelPlatformTelegram); len(tg) != 1 || tg[0].ID != bot.ID {
		t.Fatal("telegram filter", tg)
	}
	_, err = s.ListConnections(ctx, f.tenant, f.sale, "sms")
	requireSalesStatus(t, err, 400)
	_, err = s.ListConnections(ctx, f.other, f.sale, "")
	requireSalesStatus(t, err, 403)
	if other, err := s.ListConnections(ctx, f.other, f.owner, ""); err != nil || len(other) != 0 {
		t.Fatal("another workspace must not see these connections", err, other)
	}

	page := pages[0]
	_, err = s.UpdateConnection(ctx, f.tenant, f.sale, page.ID, "pause", "")
	requireSalesStatus(t, err, 403)
	_, err = s.UpdateConnection(ctx, f.other, f.owner, page.ID, "pause", "")
	requireSalesStatus(t, err, 404)
	_, err = s.UpdateConnection(ctx, f.tenant, f.owner, "not-a-uuid", "pause", "")
	requireSalesStatus(t, err, 400)
	_, err = s.UpdateConnection(ctx, f.tenant, f.owner, page.ID, "rotate_token", tgRotated)
	requireSalesStatus(t, err, 400)
	_, err = s.UpdateConnection(ctx, f.tenant, f.owner, page.ID, "pause", tgRotated)
	requireSalesStatus(t, err, 400)
	_, err = s.UpdateConnection(ctx, f.tenant, f.owner, page.ID, "archive", "")
	requireSalesStatus(t, err, 400)

	paused, err := s.UpdateConnection(ctx, f.tenant, f.owner, page.ID, "pause", "")
	if err != nil || paused.Status != types.ChannelStatusPaused || paused.Platform != types.ChannelPlatformFacebook {
		t.Fatal("pause a Page through the shared resource", err, paused)
	}
	rotated, err := s.UpdateConnection(ctx, f.tenant, f.owner, bot.ID, "rotate_token", tgRotated)
	if err != nil || rotated.Status != types.ChannelStatusActive || rotated.AccountName != "Shop Bot 2" {
		t.Fatal("rotate a bot token through the shared resource", err, rotated)
	}
	if _, err = s.UpdateConnection(ctx, f.tenant, f.owner, bot.ID, "disconnect", ""); err != nil {
		t.Fatal("disconnect a bot through the shared resource", err)
	}
	remaining, _ := s.ListConnections(ctx, f.tenant, f.owner, "")
	if len(remaining) != 2 {
		t.Fatal("a disconnected connection leaves the list", len(remaining))
	}
}
