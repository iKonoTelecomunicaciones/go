// Copyright (c) 2026 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package commands

import (
	"fmt"

	"github.com/iKonoTelecomunicaciones/go/bridgev2"
	"github.com/iKonoTelecomunicaciones/go/bridgev2/provisionutil"
	"github.com/iKonoTelecomunicaciones/go/format"
)

var CommandImportImagePack = &FullHandler{
	Func: fnImportImagePack,
	Name: "import-image-pack",
	Help: HelpMeta{
		Section:     HelpSectionMisc,
		Description: "Import a sticker or emoji pack from the remote network",
		Args:        "<url>",
	},
	RequiresLogin: true,
	NetworkAPI:    NetworkAPIImplements[bridgev2.StickerImportingNetworkAPI],
}

func fnImportImagePack(ce *Event) {
	login, _, args := getClientForStartingChat[bridgev2.StickerImportingNetworkAPI](ce, "importing pack")
	if len(args) == 0 {
		ce.Reply("Usage: `$cmdprefix import-image-pack <url>`")
		return
	}
	resp, err := provisionutil.ImportImagePack(ce.Ctx, login, args[0])
	if err != nil {
		ce.Reply("Failed to import pack: %s", err)
		return
	}
	var footer string
	parts := len(resp.StateKeys)
	if parts > 1 {
		footer = fmt.Sprintf(". Note: the pack was large, so it had to be split up into %d parts", parts)
	}
	ce.Reply(
		"Successfully bridged image pack to %s%s",
		format.MarkdownLink("your personal filtering space",
			resp.RoomID.URI(ce.Bridge.Matrix.ServerName()).MatrixToURL()),
		footer,
	)
}
