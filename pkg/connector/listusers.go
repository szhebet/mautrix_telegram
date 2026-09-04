// mautrix-telegram - A Matrix-Telegram puppeting bridge.
// Copyright (C) 2025 Sumner Evans
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
	"sort"
	"strings"

	"maunium.net/go/mautrix/bridgev2/commands"
	"maunium.net/go/mautrix/format"
)

type userInfo struct {
	mxid        string
	telegramID  string
	remoteName  string
	phone       string
	bridgeState string
}

var cmdListUsers = &commands.FullHandler{
	Func: fnListUsers,
	Name: "list-users",
	Help: commands.HelpMeta{
		Section:     commands.HelpSectionAdmin,
		Description: "List all users logged in to the bridge who are bound to a Telegram account",
	},
	RequiresAdmin: true,
}

func fnListUsers(ce *commands.Event) {
	userIDs, err := ce.Bridge.DB.UserLogin.GetAllUserIDsWithLogins(ce.Ctx)
	if err != nil {
		ce.Reply("Failed to get users: %v", err)
		return
	}

	infos := make([]userInfo, 0, len(userIDs))
	for _, userID := range userIDs {
		user, err := ce.Bridge.GetExistingUserByMXID(ce.Ctx, userID)
		if err != nil || user == nil {
			continue
		}
		for _, login := range user.GetUserLogins() {
			metadata, ok := login.Metadata.(*UserLoginMetadata)
			if !ok || !metadata.Session.HasAuthKey() {
				// Not actually bound to a Telegram account (no valid session).
				continue
			}
			infos = append(infos, userInfo{
				mxid:        string(user.MXID),
				telegramID:  string(login.ID),
				remoteName:  login.RemoteName,
				phone:       metadata.LoginPhone,
				bridgeState: string(login.BridgeState.GetPrev().StateEvent),
			})
		}
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].mxid < infos[j].mxid
	})

	if len(infos) == 0 {
		ce.Reply("No users are currently logged in to the bridge.")
		return
	}

	var buf strings.Builder
	buf.WriteString("Users bound to Telegram:\n")
	for _, info := range infos {
		buf.WriteString("* ")
		buf.WriteString(format.SafeMarkdownCode(info.mxid))
		buf.WriteString(" - Telegram ")
		buf.WriteString(format.SafeMarkdownCode(info.telegramID))
		if info.remoteName != "" {
			buf.WriteString(" (" + format.SafeMarkdownCode(info.remoteName) + ")")
		}
		if info.phone != "" {
			buf.WriteString(" - phone " + format.SafeMarkdownCode(info.phone))
		}
		if info.bridgeState != "" {
			buf.WriteString(" - " + info.bridgeState)
		}
		buf.WriteByte('\n')
	}
	ce.Reply(buf.String())
}
