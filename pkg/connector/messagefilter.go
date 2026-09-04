// mautrix-telegram - A Matrix-Telegram puppeting bridge.
// Copyright (C) 2026
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package connector

import (
	"fmt"
	"slices"
	"strings"

	"go.mau.fi/mautrix-telegram/pkg/gotd/tg"
)

func getTelegramSenderID(msg *tg.Message) int64 {
	if fromID, ok := msg.GetFromID(); ok {
		if userPeer, ok := fromID.(*tg.PeerUser); ok {
			return userPeer.UserID
		}
	}
	return 0
}

func (tc *TelegramClient) shouldDropMessage(msg *tg.Message) (bool, string) {
	cfg := tc.main.Config.MessageFilter
	if !cfg.Enabled {
		return false, ""
	}

	senderID := getTelegramSenderID(msg)

	if _, isDM := msg.PeerID.(*tg.PeerUser); isDM {
		if cfg.BlockUnknownDMSenders && len(cfg.AllowedSenders) > 0 && !slices.Contains(cfg.AllowedSenders, senderID) {
			return true, fmt.Sprintf("sender %d is not in allowed senders for direct messages", senderID)
		}
	}

	if senderID != 0 && slices.Contains(cfg.BlockedSenders, senderID) {
		return true, fmt.Sprintf("sender %d is blocked", senderID)
	}

	if len(cfg.BlockedKeywords) > 0 || len(cfg.blockedKeywordRegexps) > 0 {
		text := strings.ToLower(msg.Message)
		for _, kw := range cfg.BlockedKeywords {
			if kw == "" || (strings.HasPrefix(kw, "/") && strings.HasSuffix(kw, "/") && !strings.HasSuffix(kw, "\\/")) {
				// Skip empty entries and slash-wrapped regex entries (handled below).
				continue
			}
			if strings.Contains(text, strings.ToLower(kw)) {
				return true, fmt.Sprintf("message matches blocked keyword %q", kw)
			}
		}
		for _, re := range cfg.blockedKeywordRegexps {
			if re.MatchString(text) {
				return true, fmt.Sprintf("message matches blocked regex %q", re.String())
			}
		}
	}

	return false, ""
}
