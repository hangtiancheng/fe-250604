// Copyright (c) 2026 hangtiancheng
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package service

import (
	"context"
	"log"
	"time"

	"github.com/hangtiancheng/swifty-chat/server/internal/agent_hub"
	"github.com/hangtiancheng/swifty-chat/server/internal/constant"
	"github.com/hangtiancheng/swifty-chat/server/internal/dao"
	"github.com/hangtiancheng/swifty-chat/server/internal/model"
	"github.com/hangtiancheng/swifty-chat/server/internal/util"

	"github.com/hangtiancheng/swifty.go/swifty_orm"
)

// EnsureSwiftyUser creates the reserved Swifty assistant account on startup.
func EnsureSwiftyUser(ctx context.Context) {
	var existing model.UserInfo
	err := dao.Engine.Model(&existing).Where("uuid", constant.SwiftyUUID).First(ctx, &existing)
	if err == nil {
		return
	}
	if err != swifty_orm.ErrNotFound {
		log.Printf("EnsureSwiftyUser: lookup failed: %v", err)
		return
	}
	user := model.UserInfo{
		Uuid:      constant.SwiftyUUID,
		Nickname:  constant.SwiftyName,
		Signature: constant.SwiftySignature,
		CreatedAt: time.Now(),
		Status:    constant.UserStatusNormal,
	}
	if _, err := dao.Engine.Model(&user).Insert(ctx, &user); err != nil {
		log.Printf("EnsureSwiftyUser: insert failed: %v", err)
		return
	}
	log.Printf("Swifty assistant account created (%s)", constant.SwiftyUUID)
}

// EnsureSwiftyContact idempotently gives a user the undeletable Swifty
// contact and a session pointing at it. Called on register and login so
// existing accounts pick it up too.
func EnsureSwiftyContact(ctx context.Context, userId string) {
	if userId == "" || userId == constant.SwiftyUUID {
		return
	}
	if err := ensureUserContact(ctx, dao.Engine, userId, constant.SwiftyUUID); err != nil {
		log.Printf("EnsureSwiftyContact %s: contact failed: %v", userId, err)
		return
	}
	if err := ensureUserContact(ctx, dao.Engine, constant.SwiftyUUID, userId); err != nil {
		log.Printf("EnsureSwiftyContact %s: reverse contact failed: %v", userId, err)
	}
	ensurePeerSession(ctx, userId, constant.SwiftyUUID)
}

// IsSwifty reports whether the given uuid is the built-in assistant.
func IsSwifty(uuid string) bool {
	return uuid == constant.SwiftyUUID
}

// AgentHub runs the assistant behind Swifty. It stays nil when the server is
// started without a swifty configuration, in which case chat keeps working and
// only the assistant thread reports itself unavailable.
var AgentHub *agent_hub.Manager

// InitAgentHub starts the assistant runtime. It is called from main rather than
// init so it runs after Mongo and the cache are up.
func InitAgentHub() {
	AgentHub = agent_hub.NewManager(swiftySink{})
}

func StopAgentHub() {
	if AgentHub != nil {
		AgentHub.Stop()
	}
}

// swiftySink files the assistant's replies as ordinary chat messages. Going
// through the same insert-and-broadcast path a human peer takes is what gives
// Swifty working history, session previews and unread counts for free.
type swiftySink struct{}

func (swiftySink) SaveAssistantText(userID, sessionID, text string) string {
	ctx := bgCtx()
	name, avatar, ok := resolveSender(ctx, constant.SwiftyUUID)
	if !ok {
		name = constant.SwiftyName
	}
	msg := model.Message{
		Uuid:       "M" + util.GetNowAndLenRandomString(11),
		SessionId:  sessionID,
		Type:       constant.MessageText,
		Content:    text,
		SendId:     constant.SwiftyUUID,
		SendName:   name,
		SendAvatar: avatar,
		ReceiveId:  userID,
		Status:     constant.MessageUnsent,
		CreatedAt:  time.Now(),
	}
	if _, err := dao.Engine.Model(&msg).Insert(ctx, &msg); err != nil {
		log.Printf("swiftySink: insert reply failed: %v", err)
		return ""
	}
	ChatServer.broadcast(ChatMessageRequest{SendAvatar: avatar}, msg, false)
	return msg.Uuid
}

// dispatchToSwifty routes a stored direct message into the assistant owned by
// its sender. Group threads are deliberately left out: Swifty only takes part
// in one-to-one conversations.
func dispatchToSwifty(msg *model.Message) {
	if !IsSwifty(msg.ReceiveId) || msg.SendId == constant.SwiftyUUID {
		return
	}
	if AgentHub == nil {
		return
	}
	if msg.Type != constant.MessageText {
		// Uploads are hidden in the assistant thread, so anything else here
		// came from another client and deserves an answer rather than silence.
		swiftySink{}.SaveAssistantText(msg.SendId, msg.SessionId,
			"I can only read text messages — please describe what you need in writing.")
		return
	}
	AgentHub.Dispatch(msg.SendId, msg.SessionId, msg.Uuid, msg.Content)
}
