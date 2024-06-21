// Copyright (c) 2024 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package database

import (
	"context"
	"database/sql"
	"time"

	"go.mau.fi/util/dbutil"

	"maunium.net/go/mautrix/bridgev2/networkid"
	"maunium.net/go/mautrix/id"
)

type UserPortalQuery struct {
	BridgeID networkid.BridgeID
	*dbutil.QueryHelper[*UserPortal]
}

type UserPortal struct {
	BridgeID  networkid.BridgeID
	UserMXID  id.UserID
	LoginID   networkid.UserLoginID
	Portal    networkid.PortalKey
	InSpace   *bool
	Preferred *bool
	LastRead  time.Time
}

func newUserPortal(_ *dbutil.QueryHelper[*UserPortal]) *UserPortal {
	return &UserPortal{}
}

const (
	getUserPortalBaseQuery = `
		SELECT bridge_id, user_mxid, login_id, portal_id, portal_receiver, in_space, preferred, last_read
		FROM user_portal
	`
	getUserPortalQuery = getUserPortalBaseQuery + `
		WHERE bridge_id=$1 AND user_mxid=$2 AND login_id=$3 AND portal_id=$4 AND portal_receiver=$5
	`
	findUserLoginsByPortalIDQuery = getUserPortalBaseQuery + `
		WHERE bridge_id=$1 AND user_mxid=$2 AND portal_id=$3 AND portal_receiver=$4
		ORDER BY CASE WHEN preferred THEN 0 ELSE 1 END, login_id
	`
	getAllPortalsForLoginQuery = getUserPortalBaseQuery + `
		WHERE bridge_id=$1 AND user_mxid=$2 AND login_id=$3
	`
	insertUserPortalQuery = `
		INSERT INTO user_portal (bridge_id, user_mxid, login_id, portal_id, portal_receiver, in_space, preferred)
		VALUES ($1, $2, $3, $4, $5, false, false)
		ON CONFLICT (bridge_id, user_mxid, login_id, portal_id, portal_receiver) DO NOTHING
	`
	upsertUserPortalQuery = `
		INSERT INTO user_portal (bridge_id, user_mxid, login_id, portal_id, portal_receiver, in_space, preferred, last_read)
		VALUES ($1, $2, $3, $4, $5, COALESCE($6, false), COALESCE($7, false), $8)
		ON CONFLICT (bridge_id, user_mxid, login_id, portal_id, portal_receiver) DO UPDATE
			SET in_space=COALESCE($6, user_portal.in_space),
			    preferred=COALESCE($7, user_portal.preferred),
			    last_read=COALESCE($8, user_portal.last_read)
	`
	markLoginAsPreferredQuery = `
		UPDATE user_portal SET preferred=(login_id=$3) WHERE bridge_id=$1 AND user_mxid=$2 AND portal_id=$4 AND portal_receiver=$5
	`
)

func UserPortalFor(ul *UserLogin, portal networkid.PortalKey) *UserPortal {
	return &UserPortal{
		BridgeID: ul.BridgeID,
		UserMXID: ul.UserMXID,
		LoginID:  ul.ID,
		Portal:   portal,
	}
}

func (upq *UserPortalQuery) GetAllByUser(ctx context.Context, userID id.UserID, portal networkid.PortalKey) ([]*UserPortal, error) {
	return upq.QueryMany(ctx, findUserLoginsByPortalIDQuery, upq.BridgeID, userID, portal.ID, portal.Receiver)
}

func (upq *UserPortalQuery) GetAllForLogin(ctx context.Context, login *UserLogin) ([]*UserPortal, error) {
	return upq.QueryMany(ctx, getUserPortalQuery, upq.BridgeID, login.UserMXID, login.ID)
}

func (upq *UserPortalQuery) Get(ctx context.Context, login *UserLogin, portal networkid.PortalKey) (*UserPortal, error) {
	return upq.QueryOne(ctx, getUserPortalQuery, upq.BridgeID, login.UserMXID, login.ID, portal.ID, portal.Receiver)
}

func (upq *UserPortalQuery) Put(ctx context.Context, up *UserPortal) error {
	ensureBridgeIDMatches(&up.BridgeID, upq.BridgeID)
	return upq.Exec(ctx, upsertUserPortalQuery, up.sqlVariables()...)
}

func (upq *UserPortalQuery) EnsureExists(ctx context.Context, login *UserLogin, portal networkid.PortalKey) error {
	return upq.Exec(ctx, insertUserPortalQuery, upq.BridgeID, login.UserMXID, login.ID, portal.ID, portal.Receiver)
}

func (upq *UserPortalQuery) MarkAsPreferred(ctx context.Context, login *UserLogin, portal networkid.PortalKey) error {
	return upq.Exec(ctx, markLoginAsPreferredQuery, upq.BridgeID, login.UserMXID, login.ID, portal.ID, portal.Receiver)
}

func (up *UserPortal) Scan(row dbutil.Scannable) (*UserPortal, error) {
	var lastRead sql.NullInt64
	err := row.Scan(
		&up.BridgeID, &up.UserMXID, &up.LoginID, &up.Portal.ID, &up.Portal.Receiver,
		&up.InSpace, &up.Preferred, &lastRead,
	)
	if err != nil {
		return nil, err
	}
	if lastRead.Valid {
		up.LastRead = time.Unix(0, lastRead.Int64)
	}
	return up, nil
}

func (up *UserPortal) sqlVariables() []any {
	return []any{
		up.BridgeID, up.UserMXID, up.LoginID, up.Portal.ID, up.Portal.Receiver,
		up.InSpace,
		up.Preferred,
		dbutil.ConvertedPtr(up.LastRead, time.Time.UnixNano),
	}
}

func (up *UserPortal) ResetValues() {
	up.InSpace = nil
	up.Preferred = nil
	up.LastRead = time.Time{}
}
