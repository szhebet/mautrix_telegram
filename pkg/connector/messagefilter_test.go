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

import "testing"

func TestParseKeywordRegexp(t *testing.T) {
	cases := []struct {
		name    string
		entry   string
		wantOK  bool
		wantErr bool
	}{
		{name: "plain substring", entry: "casino", wantOK: false, wantErr: false},
		{name: "basic regex", entry: "/foo.*bar/", wantOK: true, wantErr: false},
		{name: "regex with pipe", entry: "/foo|bar/", wantOK: true, wantErr: false},
		{name: "invalid regex", entry: "/[unclosed/", wantOK: false, wantErr: true},
		{name: "empty slash", entry: "//", wantOK: true, wantErr: false},
		{name: "single slash", entry: "/", wantOK: false, wantErr: false},
		{name: "trailing escaped slash", entry: "/foo\\/", wantOK: false, wantErr: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok, err := parseKeywordRegexp(tc.entry)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (ok=%v, re=%v)", ok, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ok != tc.wantOK {
				t.Fatalf("ok mismatch: got %v, want %v", ok, tc.wantOK)
			}
		})
	}
}

func TestParseKeywordRegexpMatching(t *testing.T) {
	re, ok, err := parseKeywordRegexp("/FOO.*bar/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected ok=true for slash-wrapped entry")
	}
	if !re.MatchString("some FOO value bar here") {
		t.Fatal("expected regex to match")
	}
}

func TestMessageFilterRegexAndSubstr(t *testing.T) {
	cfg := &MessageFilterConfig{
		Enabled:        true,
		BlockedSenders: []int64{111},
		AllowedSenders: []int64{222},
		BlockUnknownDMSenders: true,
		BlockedKeywords: []string{"casino", "/win[0-9]+/"},
	}
	if err := cfg.compileRegexps(); err != nil {
		t.Fatalf("compileRegexps error: %v", err)
	}
	if len(cfg.blockedKeywordRegexps) != 1 {
		t.Fatalf("expected 1 compiled regex, got %d", len(cfg.blockedKeywordRegexps))
	}
}
