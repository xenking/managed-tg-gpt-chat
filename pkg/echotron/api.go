/*
 * Echotron
 * Copyright (C) 2018-2022 The Echotron Devs
 *
 * Echotron is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Echotron is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package echotron

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// API is the object that contains all the functions that wrap those of the Telegram SessionHandler API.
type API struct {
	client *client
	token  string
	base   string
}

// NewAPI returns a new API object.
func NewAPI(token string) API {
	return API{
		token:  token,
		base:   fmt.Sprintf("https://api.telegram.org/bot%s/", token),
		client: lclient,
	}
}

// NewLocalAPI is like NewAPI but allows to use a local API server.
func NewLocalAPI(url, token string) API {
	return API{
		token:  token,
		base:   url,
		client: lclient,
	}
}

// GetUpdates is used to receive incoming updates using long polling.
func (a API) GetUpdates(ctx context.Context, opts *UpdateOptions) (res APIResponseUpdate, err error) {
	return res, a.client.get(ctx, a.base, "getUpdates", urlValues(opts), &res)
}

// SetWebhook is used to specify a url and receive incoming updates via an outgoing webhook.
func (a API) SetWebhook(ctx context.Context, webhookURL string, dropPendingUpdates bool, opts *WebhookOptions) (res APIResponseBase, err error) {
	var (
		vals   = make(url.Values)
		keyVal = map[string]string{"url": webhookURL}
	)

	url, err := url.JoinPath(a.base, "setWebhook")
	if err != nil {
		return res, err
	}

	vals.Set("drop_pending_updates", btoa(dropPendingUpdates))
	addValues(vals, opts)
	url = fmt.Sprintf("%s?%s", strings.TrimSuffix(url, "/"), vals.Encode())

	cnt, err := a.client.doPostForm(ctx, url, keyVal)
	if err != nil {
		return
	}

	if err = json.Unmarshal(cnt, &res); err != nil {
		return
	}

	err = check(res)
	return
}

// DeleteWebhook is used to remove webhook integration if you decide to switch back to GetUpdates.
func (a API) DeleteWebhook(ctx context.Context, dropPendingUpdates bool) (res APIResponseBase, err error) {
	vals := make(url.Values)
	vals.Set("drop_pending_updates", btoa(dropPendingUpdates))

	return res, a.client.get(ctx, a.base, "deleteWebhook", vals, &res)
}

// GetWebhookInfo is used to get current webhook status.
func (a API) GetWebhookInfo(ctx context.Context) (res APIResponseWebhook, err error) {
	return res, a.client.get(ctx, a.base, "getWebhookInfo", nil, &res)
}

// GetMe is a simple method for testing your bot's auth token.
func (a API) GetMe(ctx context.Context) (res APIResponseUser, err error) {
	return res, a.client.get(ctx, a.base, "getMe", nil, &res)
}

// LogOut is used to log out from the cloud SessionHandler API server before launching the bot locally.
// You MUST log out the bot before running it locally, otherwise there is no guarantee that the bot will receive updates.
// After a successful call, you can immediately log in on a local server,
// but will not be able to log in back to the cloud SessionHandler API server for 10 minutes.
func (a API) LogOut(ctx context.Context) (res APIResponseBool, err error) {
	return res, a.client.get(ctx, a.base, "logOut", nil, &res)
}

// Close is used to close the bot instance before moving it from one local server to another.
// You need to delete the webhook before calling this method to ensure that the bot isn't launched again after server restart.
// The method will return error 429 in the first 10 minutes after the bot is launched.
func (a API) Close(ctx context.Context) (res APIResponseBool, err error) {
	return res, a.client.get(ctx, a.base, "close", nil, &res)
}

// SendMessage is used to send text messages.
func (a API) SendMessage(ctx context.Context, text string, chatID int64, opts *MessageOptions) (res APIResponseMessage, err error) {
	vals := make(url.Values)

	vals.Set("text", text)
	vals.Set("chat_id", itoa(chatID))
	return res, a.client.get(ctx, a.base, "sendMessage", addValues(vals, opts), &res)
}

// ForwardMessage is used to forward messages of any kind.
// Service messages can't be forwarded.
func (a API) ForwardMessage(ctx context.Context, chatID, fromChatID int64, messageID int, opts *ForwardOptions) (res APIResponseMessage, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	vals.Set("from_chat_id", itoa(fromChatID))
	vals.Set("message_id", itoa(int64(messageID)))
	return res, a.client.get(ctx, a.base, "forwardMessage", addValues(vals, opts), &res)
}

// ForwardMessages is used to forward multiple messages of any kind.
// If some of the specified messages can't be found or forwarded, they are skipped.
// Service messages and messages with protected content can't be forwarded.
// Album grouping is kept for forwarded messages.
func (a API) ForwardMessages(ctx context.Context, chatID, fromChatID int64, messageIDs []int, opts *ForwardOptions) (res APIResponseMessageIDs, err error) {
	vals := make(url.Values)

	msgIDs, err := json.Marshal(messageIDs)
	if err != nil {
		return res, err
	}

	vals.Set("chat_id", itoa(chatID))
	vals.Set("from_chat_id", itoa(fromChatID))
	vals.Set("message_ids", string(msgIDs))
	return res, a.client.get(ctx, a.base, "forwardMessages", addValues(vals, opts), &res)
}

// CopyMessage is used to copy messages of any kind.
// Service messages, paid media mesages, giveaway messages, giveaway winners messages, and invoice messages can't be copied.
// The method is analogous to the method ForwardMessage,
// but the copied message doesn't have a link to the original message.
func (a API) CopyMessage(ctx context.Context, chatID, fromChatID int64, messageID int, opts *CopyOptions) (res APIResponseMessageID, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	vals.Set("from_chat_id", itoa(fromChatID))
	vals.Set("message_id", itoa(int64(messageID)))
	return res, a.client.get(ctx, a.base, "copyMessage", addValues(vals, opts), &res)
}

// CopyMessages is used to copy messages of any kind.
// If some of the specified messages can't be found or copied, they are skipped.
// Service messages, paid media mesages, giveaway messages, giveaway winners messages, and invoice messages can't be copied.
// A quiz poll can be copied only if the value of the field correct_option_id is known to the bot.
// The method is analogous to the method forwardMessages, but the copied messages don't have a link to the original message.
// Album grouping is kept for copied messages.
func (a API) CopyMessages(ctx context.Context, chatID, fromChatID int64, messageIDs []int, opts *CopyMessagesOptions) (res APIResponseMessageIDs, err error) {
	vals := make(url.Values)

	msgIDs, err := json.Marshal(messageIDs)
	if err != nil {
		return res, err
	}

	vals.Set("chat_id", itoa(chatID))
	vals.Set("from_chat_id", itoa(fromChatID))
	vals.Set("message_ids", string(msgIDs))
	return res, a.client.get(ctx, a.base, "copyMessages", addValues(vals, opts), &res)
}

// SendPhoto is used to send photos.
func (a API) SendPhoto(ctx context.Context, file InputFile, chatID int64, opts *PhotoOptions) (res APIResponseMessage, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	return res, a.client.postFile(ctx, a.base, "sendPhoto", "photo", file, InputFile{}, addValues(vals, opts), &res)
}

// SendAudio is used to send audio files,
// if you want Telegram clients to display them in the music player.
// Your audio must be in the .MP3 or .M4A format.
func (a API) SendAudio(ctx context.Context, file InputFile, chatID int64, opts *AudioOptions) (res APIResponseMessage, err error) {
	var (
		thumbnail InputFile
		vals      = make(url.Values)
	)

	if opts != nil {
		thumbnail = opts.Thumbnail
	}

	vals.Set("chat_id", itoa(chatID))
	return res, a.client.postFile(ctx, a.base, "sendAudio", "audio", file, thumbnail, addValues(vals, opts), &res)
}

// SendDocument is used to send general files.
func (a API) SendDocument(ctx context.Context, file InputFile, chatID int64, opts *DocumentOptions) (res APIResponseMessage, err error) {
	var (
		thumbnail InputFile
		vals      = make(url.Values)
	)

	if opts != nil {
		thumbnail = opts.Thumbnail
	}

	vals.Set("chat_id", itoa(chatID))
	return res, a.client.postFile(ctx, a.base, "sendDocument", "document", file, thumbnail, addValues(vals, opts), &res)
}

// SendVideo is used to send video files.
// Telegram clients support mp4 videos (other formats may be sent with SendDocument).
func (a API) SendVideo(ctx context.Context, file InputFile, chatID int64, opts *VideoOptions) (res APIResponseMessage, err error) {
	var (
		thumbnail InputFile
		vals      = make(url.Values)
	)

	if opts != nil {
		thumbnail = opts.Thumbnail
	}

	vals.Set("chat_id", itoa(chatID))
	return res, a.client.postFile(ctx, a.base, "sendVideo", "video", file, thumbnail, addValues(vals, opts), &res)
}

// SendAnimation is used to send animation files (GIF or H.264/MPEG-4 AVC video without sound).
func (a API) SendAnimation(ctx context.Context, file InputFile, chatID int64, opts *AnimationOptions) (res APIResponseMessage, err error) {
	var (
		thumbnail InputFile
		vals      = make(url.Values)
	)

	if opts != nil {
		thumbnail = opts.Thumbnail
	}

	vals.Set("chat_id", itoa(chatID))
	return res, a.client.postFile(ctx, a.base, "sendAnimation", "animation", file, thumbnail, addValues(vals, opts), &res)
}

// SendVoice is used to send audio files, if you want Telegram clients to display the file as a playable voice message.
// For this to work, your audio must be in an .OGG file encoded with OPUS (other formats may be sent as Audio or Document).
func (a API) SendVoice(ctx context.Context, file InputFile, chatID int64, opts *VoiceOptions) (res APIResponseMessage, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	return res, a.client.postFile(ctx, a.base, "sendVoice", "voice", file, InputFile{}, addValues(vals, opts), &res)
}

// SendVideoNote is used to send video messages.
func (a API) SendVideoNote(ctx context.Context, file InputFile, chatID int64, opts *VideoNoteOptions) (res APIResponseMessage, err error) {
	var (
		thumbnail InputFile
		vals      = make(url.Values)
	)

	if opts != nil {
		thumbnail = opts.Thumbnail
	}

	vals.Set("chat_id", itoa(chatID))
	return res, a.client.postFile(ctx, a.base, "sendVideoNote", "video_note", file, thumbnail, addValues(vals, opts), &res)
}

// SendPaidMedia is used to send paid media to channel chats.
func (a API) SendPaidMedia(ctx context.Context, chatID, starCount int64, media []GroupableInputMedia, opts *PaidMediaOptions) (res APIResponseMessage, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	vals.Set("star_count", itoa(starCount))
	return res, a.client.postMedia(ctx, a.base, "sendPaidMedia", false, addValues(vals, opts), &res, toInputMedia(media)...)
}

// SendMediaGroup is used to send a group of photos, videos, documents or audios as an album.
// Documents and audio files can be only grouped in an album with messages of the same type.
func (a API) SendMediaGroup(ctx context.Context, chatID int64, media []GroupableInputMedia, opts *MediaGroupOptions) (res APIResponseMessageArray, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	return res, a.client.postMedia(ctx, a.base, "sendMediaGroup", false, addValues(vals, opts), &res, toInputMedia(media)...)
}

// SendLocation is used to send point on the map.
func (a API) SendLocation(ctx context.Context, chatID int64, latitude, longitude float64, opts *LocationOptions) (res APIResponseMessage, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	vals.Set("latitude", ftoa(latitude))
	vals.Set("longitude", ftoa(longitude))
	return res, a.client.get(ctx, a.base, "sendLocation", addValues(vals, opts), &res)
}

// EditMessageLiveLocation is used to edit live location messages.
// A location can be edited until its `LivePeriod` expires or editing is explicitly disabled by a call to `StopMessageLiveLocation`.
func (a API) EditMessageLiveLocation(ctx context.Context, msg MessageIDOptions, latitude, longitude float64, opts *EditLocationOptions) (res APIResponseMessage, err error) {
	vals := make(url.Values)

	vals.Set("latitude", ftoa(latitude))
	vals.Set("longitude", ftoa(longitude))
	return res, a.client.get(ctx, a.base, "editMessageLiveLocation", addValues(addValues(vals, msg), opts), &res)
}

// StopMessageLiveLocation is used to stop updating a live location message before `LivePeriod` expires.
func (a API) StopMessageLiveLocation(ctx context.Context, msg MessageIDOptions, opts *StopLocationOptions) (res APIResponseMessage, err error) {
	return res, a.client.get(ctx, a.base, "stopMessageLiveLocation", addValues(urlValues(msg), opts), &res)
}

// SendVenue is used to send information about a venue.
func (a API) SendVenue(ctx context.Context, chatID int64, latitude, longitude float64, title, address string, opts *VenueOptions) (res APIResponseMessage, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	vals.Set("latitude", ftoa(latitude))
	vals.Set("longitude", ftoa(longitude))
	vals.Set("title", title)
	vals.Set("address", address)
	return res, a.client.get(ctx, a.base, "sendVenue", addValues(vals, opts), &res)
}

// SendContact is used to send phone contacts.
func (a API) SendContact(ctx context.Context, phoneNumber, firstName string, chatID int64, opts *ContactOptions) (res APIResponseMessage, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	vals.Set("phone_number", phoneNumber)
	vals.Set("first_name", firstName)
	return res, a.client.get(ctx, a.base, "sendContact", addValues(vals, opts), &res)
}

// SendPoll is used to send a native poll.
func (a API) SendPoll(ctx context.Context, chatID int64, question string, options []InputPollOption, opts *PollOptions) (res APIResponseMessage, err error) {
	vals := make(url.Values)

	pollOpts, err := json.Marshal(options)
	if err != nil {
		return res, err
	}

	vals.Set("chat_id", itoa(chatID))
	vals.Set("question", question)
	vals.Set("options", string(pollOpts))
	return res, a.client.get(ctx, a.base, "sendPoll", addValues(vals, opts), &res)
}

// SendDice is used to send an animated emoji that will display a random value.
func (a API) SendDice(ctx context.Context, chatID int64, emoji DiceEmoji, opts *BaseOptions) (res APIResponseMessage, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	vals.Set("emoji", string(emoji))
	return res, a.client.get(ctx, a.base, "sendDice", addValues(vals, opts), &res)
}

// SendChatAction is used to tell the user that something is happening on the bot's side.
// The status is set for 5 seconds or less (when a message arrives from your bot, Telegram clients clear its typing status).
func (a API) SendChatAction(ctx context.Context, action ChatAction, chatID int64, opts *ChatActionOptions) (res APIResponseBool, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	vals.Set("action", string(action))
	return res, a.client.get(ctx, a.base, "sendChatAction", addValues(vals, opts), &res)
}

// SetMessageReaction is used to change the chosen reactions on a message.
// Service messages can't be reacted to.
// Automatically forwarded messages from a channel to its discussion group have the same available reactions as messages in the channel.
// In albums, bots must react to the first message.
func (a API) SetMessageReaction(ctx context.Context, chatID int64, messageID int, opts *MessageReactionOptions) (res APIResponseBool, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	vals.Set("message_id", itoa(int64(messageID)))
	return res, a.client.get(ctx, a.base, "setMessageReaction", addValues(vals, opts), &res)
}

// GetUserProfilePhotos is used to get a list of profile pictures for a user.
func (a API) GetUserProfilePhotos(ctx context.Context, userID int64, opts *UserProfileOptions) (res APIResponseUserProfile, err error) {
	vals := make(url.Values)

	vals.Set("user_id", itoa(userID))
	return res, a.client.get(ctx, a.base, "getUserProfilePhotos", addValues(vals, opts), &res)
}

// GetFile returns the basic info about a file and prepares it for downloading.
// For the moment, bots can download files of up to 20MB in size.
// The file can then be downloaded with DownloadFile where filePath is taken from the response.
// It is guaranteed that the file will be downloadable for at least 1 hour.
// When the download file expires, a new one can be requested by calling GetFile again.
func (a API) GetFile(ctx context.Context, fileID string) (res APIResponseFile, err error) {
	vals := make(url.Values)

	vals.Set("file_id", fileID)
	return res, a.client.get(ctx, a.base, "getFile", vals, &res)
}

// DownloadFile returns the bytes of the file corresponding to the given filePath.
// This function is callable for at least 1 hour since the call to GetFile.
// When the download expires a new one can be requested by calling GetFile again.
func (a API) DownloadFile(ctx context.Context, filePath string) ([]byte, error) {
	return a.client.doGet(ctx, fmt.Sprintf(
		"https://api.telegram.org/file/bot%s/%s",
		a.token,
		filePath,
	))
}

// BanChatMember is used to ban a user in a group, a supergroup or a channel.
// In the case of supergroups or channels, the user will not be able to return to the chat
// on their own using invite links, etc., unless unbanned first (through the UnbanChatMember method).
// The bot must be an administrator in the chat for this to work and must have the appropriate admin rights.
func (a API) BanChatMember(ctx context.Context, chatID, userID int64, opts *BanOptions) (res APIResponseBool, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	vals.Set("user_id", itoa(userID))
	return res, a.client.get(ctx, a.base, "banChatMember", addValues(vals, opts), &res)
}

// UnbanChatMember is used to unban a previously banned user in a supergroup or channel.
// The user will NOT return to the group or channel automatically, but will be able to join via link, etc.
// The bot must be an administrator for this to work.
// By default, this method guarantees that after the call the user is not a member of the chat, but will be able to join it.
// So if the user is a member of the chat they will also be REMOVED from the chat.
// If you don't want this, use the parameter `OnlyIfBanned`.
func (a API) UnbanChatMember(ctx context.Context, chatID, userID int64, opts *UnbanOptions) (res APIResponseBool, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	vals.Set("user_id", itoa(userID))
	return res, a.client.get(ctx, a.base, "unbanChatMember", addValues(vals, opts), &res)
}

// RestrictChatMember is used to restrict a user in a supergroup.
// The bot must be an administrator in the supergroup for this to work and must have the appropriate admin rights.
func (a API) RestrictChatMember(ctx context.Context, chatID, userID int64, permissions ChatPermissions, opts *RestrictOptions) (res APIResponseBool, err error) {
	vals := make(url.Values)

	perm, err := json.Marshal(permissions)
	if err != nil {
		return
	}

	vals.Set("chat_id", itoa(chatID))
	vals.Set("user_id", itoa(userID))
	vals.Set("permissions", string(perm))
	return res, a.client.get(ctx, a.base, "restrictChatMember", addValues(vals, opts), &res)
}

// PromoteChatMember is used to promote or demote a user in a supergroup or a channel.
// The bot must be an administrator in the supergroup for this to work and must have the appropriate admin rights.
func (a API) PromoteChatMember(ctx context.Context, chatID, userID int64, opts *PromoteOptions) (res APIResponseBool, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	vals.Set("user_id", itoa(userID))
	return res, a.client.get(ctx, a.base, "promoteChatMember", addValues(vals, opts), &res)
}

// SetChatAdministratorCustomTitle is used to set a custom title for an administrator in a supergroup promoted by the bot.
func (a API) SetChatAdministratorCustomTitle(ctx context.Context, chatID, userID int64, customTitle string) (res APIResponseBool, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	vals.Set("user_id", itoa(userID))
	vals.Set("custom_title", customTitle)
	return res, a.client.get(ctx, a.base, "setChatAdministratorCustomTitle", vals, &res)
}

// BanChatSenderChat is used to ban a channel chat in a supergroup or a channel.
// The owner of the chat will not be able to send messages and join live streams on behalf of the chat, unless it is unbanned first.
// The bot must be an administrator in the supergroup or channel for this to work and must have the appropriate administrator rights.
func (a API) BanChatSenderChat(ctx context.Context, chatID, senderChatID int64) (res APIResponseBool, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	vals.Set("sender_chat_id", itoa(senderChatID))
	return res, a.client.get(ctx, a.base, "banChatSenderChat", vals, &res)
}

// UnbanChatSenderChat is used to unban a previously channel chat in a supergroup or channel.
// The bot must be an administrator for this to work and must have the appropriate administrator rights.
func (a API) UnbanChatSenderChat(ctx context.Context, chatID, senderChatID int64) (res APIResponseBool, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	vals.Set("sender_chat_id", itoa(senderChatID))
	return res, a.client.get(ctx, a.base, "unbanChatSenderChat", vals, &res)
}

// SetChatPermissions is used to set default chat permissions for all members.
// The bot must be an administrator in the supergroup for this to work and must have the can_restrict_members admin rights.
func (a API) SetChatPermissions(ctx context.Context, chatID int64, permissions ChatPermissions, opts *ChatPermissionsOptions) (res APIResponseBool, err error) {
	vals := make(url.Values)

	perm, err := json.Marshal(permissions)
	if err != nil {
		return
	}

	vals.Set("chat_id", itoa(chatID))
	vals.Set("permissions", string(perm))
	return res, a.client.get(ctx, a.base, "setChatPermissions", addValues(vals, opts), &res)
}

// ExportChatInviteLink is used to generate a new primary invite link for a chat;
// any previously generated primary link is revoked.
// The bot must be an administrator in the supergroup for this to work and must have the appropriate admin rights.
func (a API) ExportChatInviteLink(ctx context.Context, chatID int64) (res APIResponseString, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	return res, a.client.get(ctx, a.base, "exportChatInviteLink", vals, &res)
}

// CreateChatInviteLink is used to create an additional invite link for a chat.
// The bot must be an administrator in the supergroup for this to work and must have the appropriate admin rights.
// The link can be revoked using the method RevokeChatInviteLink.
func (a API) CreateChatInviteLink(ctx context.Context, chatID int64, opts *InviteLinkOptions) (res APIResponseInviteLink, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	return res, a.client.get(ctx, a.base, "createChatInviteLink", addValues(vals, opts), &res)
}

// EditChatInviteLink is used to edit a non-primary invite link created by the bot.
// The bot must be an administrator in the supergroup for this to work and must have the appropriate admin rights.
func (a API) EditChatInviteLink(ctx context.Context, chatID int64, inviteLink string, opts *InviteLinkOptions) (res APIResponseInviteLink, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	vals.Set("invite_link", inviteLink)
	return res, a.client.get(ctx, a.base, "editChatInviteLink", addValues(vals, opts), &res)
}

// CreateChatSubscriptionInviteLink is used to create a subscription invite link for a channel chat.
// The bot must have the can_invite_users administrator rights.
// The link can be edited using the method editChatSubscriptionInviteLink or revoked using the method revokeChatInviteLink.
func (a API) CreateChatSubscriptionInviteLink(ctx context.Context, chatID int64, subscriptionPeriod, subscriptionPrice int, opts *ChatSubscriptionInviteOptions) (res APIResponseInviteLink, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	vals.Set("subscription_period", itoa(int64(subscriptionPeriod)))
	vals.Set("subscription_price", itoa(int64(subscriptionPrice)))
	return res, a.client.get(ctx, a.base, "createChatSubscriptionInviteLink", addValues(vals, opts), &res)
}

// EditChatSubscriptionInviteLink is used to creeditate a subscription invite link for a channel chat.
// The bot must have the can_invite_users administrator rights.
func (a API) EditChatSubscriptionInviteLink(ctx context.Context, chatID int64, inviteLink string, opts *ChatSubscriptionInviteOptions) (res APIResponseInviteLink, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	vals.Set("invite_link", inviteLink)
	return res, a.client.get(ctx, a.base, "editChatSubscriptionInviteLink", addValues(vals, opts), &res)
}

// RevokeChatInviteLink is used to revoke an invite link created by the bot.
// If the primary link is revoked, a new link is automatically generated.
// The bot must be an administrator in the supergroup for this to work and must have the appropriate admin rights.
func (a API) RevokeChatInviteLink(ctx context.Context, chatID int64, inviteLink string) (res APIResponseInviteLink, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	vals.Set("invite_link", inviteLink)
	return res, a.client.get(ctx, a.base, "editChatInviteLink", vals, &res)
}

// ApproveChatJoinRequest is used to approve a chat join request.
// The bot must be an administrator in the chat for this to work and must have the CanInviteUsers administrator right.
func (a API) ApproveChatJoinRequest(ctx context.Context, chatID, userID int64) (res APIResponseBool, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	vals.Set("user_id", itoa(userID))
	return res, a.client.get(ctx, a.base, "approveChatJoinRequest", vals, &res)
}

// DeclineChatJoinRequest is used to decline a chat join request.
// The bot must be an administrator in the chat for this to work and must have the CanInviteUsers administrator right.
func (a API) DeclineChatJoinRequest(ctx context.Context, chatID, userID int64) (res APIResponseBool, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	vals.Set("user_id", itoa(userID))
	return res, a.client.get(ctx, a.base, "declineChatJoinRequest", vals, &res)
}

// SetChatPhoto is used to set a new profile photo for the chat.
// Photos can't be changed for private chats.
// The bot must be an administrator in the chat for this to work and must have the appropriate admin rights.
func (a API) SetChatPhoto(ctx context.Context, file InputFile, chatID int64) (res APIResponseBool, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	return res, a.client.postFile(ctx, a.base, "setChatPhoto", "photo", file, InputFile{}, vals, &res)
}

// DeleteChatPhoto is used to delete a chat photo.
// Photos can't be changed for private chats.
// The bot must be an administrator in the chat for this to work and must have the appropriate admin rights.
func (a API) DeleteChatPhoto(ctx context.Context, chatID int64) (res APIResponseBool, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	return res, a.client.get(ctx, a.base, "deleteChatPhoto", vals, &res)
}

// SetChatTitle is used to change the title of a chat.
// Titles can't be changed for private chats.
// The bot must be an administrator in the chat for this to work and must have the appropriate admin rights.
func (a API) SetChatTitle(ctx context.Context, chatID int64, title string) (res APIResponseBool, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	vals.Set("title", title)
	return res, a.client.get(ctx, a.base, "setChatTitle", vals, &res)
}

// SetChatDescription is used to change the description of a group, a supergroup or a channel.
// The bot must be an administrator in the chat for this to work and must have the appropriate admin rights.
func (a API) SetChatDescription(ctx context.Context, chatID int64, description string) (res APIResponseBool, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	vals.Set("description", description)
	return res, a.client.get(ctx, a.base, "setChatDescription", vals, &res)
}

// PinChatMessage is used to add a message to the list of pinned messages in the chat.
// If the chat is not a private chat, the bot must be an administrator in the chat for this to work
// and must have the 'can_pin_messages' admin right in a supergroup or 'can_edit_messages' admin right in a channel.
func (a API) PinChatMessage(ctx context.Context, chatID int64, messageID int, opts *PinMessageOptions) (res APIResponseBool, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	vals.Set("message_id", itoa(int64(messageID)))
	return res, a.client.get(ctx, a.base, "pinChatMessage", addValues(vals, opts), &res)
}

// UnpinChatMessage is used to remove a message from the list of pinned messages in the chat.
// If the chat is not a private chat, the bot must be an administrator in the chat for this to work
// and must have the 'can_pin_messages' admin right in a supergroup or 'can_edit_messages' admin right in a channel.
func (a API) UnpinChatMessage(ctx context.Context, chatID int64, opts *UnpinMessageOptions) (res APIResponseBool, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	return res, a.client.get(ctx, a.base, "unpinChatMessage", addValues(vals, opts), &res)
}

// UnpinAllChatMessages is used to clear the list of pinned messages in a chat.
// If the chat is not a private chat, the bot must be an administrator in the chat for this to work
// and must have the 'can_pin_messages' admin right in a supergroup or 'can_edit_messages' admin right in a channel.
func (a API) UnpinAllChatMessages(ctx context.Context, chatID int64) (res APIResponseBool, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	return res, a.client.get(ctx, a.base, "unpinAllChatMessages", vals, &res)
}

// LeaveChat is used to make the bot leave a group, supergroup or channel.
func (a API) LeaveChat(ctx context.Context, chatID int64) (res APIResponseBool, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	return res, a.client.get(ctx, a.base, "leaveChat", vals, &res)
}

// GetChat is used to get up to date information about the chat.
// (current name of the user for one-on-one conversations, current username of a user, group or channel, etc.)
func (a API) GetChat(ctx context.Context, chatID int64) (res APIResponseChat, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	return res, a.client.get(ctx, a.base, "getChat", vals, &res)
}

// GetChatAdministrators is used to get a list of administrators in a chat.
func (a API) GetChatAdministrators(ctx context.Context, chatID int64) (res APIResponseAdministrators, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	return res, a.client.get(ctx, a.base, "getChatAdministrators", vals, &res)
}

// GetChatMemberCount is used to get the number of members in a chat.
func (a API) GetChatMemberCount(ctx context.Context, chatID int64) (res APIResponseInteger, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	return res, a.client.get(ctx, a.base, "getChatMemberCount", vals, &res)
}

// GetChatMember is used to get information about a member of a chat.
func (a API) GetChatMember(ctx context.Context, chatID, userID int64) (res APIResponseChatMember, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	vals.Set("user_id", itoa(userID))
	return res, a.client.get(ctx, a.base, "getChatMember", vals, &res)
}

// SetChatStickerSet is used to set a new group sticker set for a supergroup.
// The bot must be an administrator in the chat for this to work and must have the appropriate admin rights.
// Use the field `CanSetStickerSet` optionally returned in GetChat requests to check if the bot can use this method.
func (a API) SetChatStickerSet(ctx context.Context, chatID int64, stickerSetName string) (res APIResponseBool, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	vals.Set("sticker_set_name", stickerSetName)
	return res, a.client.get(ctx, a.base, "setChatStickerSet", vals, &res)
}

// DeleteChatStickerSet is used to delete a group sticker set for a supergroup.
// The bot must be an administrator in the chat for this to work and must have the appropriate admin rights.
// Use the field `CanSetStickerSet` optionally returned in GetChat requests to check if the bot can use this method.
func (a API) DeleteChatStickerSet(ctx context.Context, chatID int64) (res APIResponseBool, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	return res, a.client.get(ctx, a.base, "deleteChatStickerSet", vals, &res)
}

// CreateForumTopic is used to create a topic in a forum supergroup chat.
// The bot must be an administrator in the chat for this to work and must have the can_manage_topics administrator rights.
func (a API) CreateForumTopic(ctx context.Context, chatID int64, name string, opts *CreateTopicOptions) (res APIResponseForumTopic, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	vals.Set("name", name)
	return res, a.client.get(ctx, a.base, "createForumTopic", addValues(vals, opts), &res)
}

// EditForumTopic is used to edit name and icon of a topic in a forum supergroup chat.
// The bot must be an administrator in the chat for this to work and must have the can_manage_topics administrator rights.
func (a API) EditForumTopic(ctx context.Context, chatID, messageThreadID int64, opts *EditTopicOptions) (res APIResponseBool, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	vals.Set("message_thread_id", itoa(messageThreadID))
	return res, a.client.get(ctx, a.base, "editForumTopic", addValues(vals, opts), &res)
}

// CloseForumTopic is used to close an open topic in a forum supergroup chat.
// The bot must be an administrator in the chat for this to work and must have the can_manage_topics administrator rights.
func (a API) CloseForumTopic(ctx context.Context, chatID, messageThreadID int64) (res APIResponseBool, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	vals.Set("message_thread_id", itoa(messageThreadID))
	return res, a.client.get(ctx, a.base, "closeForumTopic", vals, &res)
}

// ReopenForumTopic is used to reopen a closed topic in a forum supergroup chat.
// The bot must be an administrator in the chat for this to work and must have the can_manage_topics administrator rights.
func (a API) ReopenForumTopic(ctx context.Context, chatID, messageThreadID int64) (res APIResponseBool, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	vals.Set("message_thread_id", itoa(messageThreadID))
	return res, a.client.get(ctx, a.base, "reopenForumTopic", vals, &res)
}

// DeleteForumTopic is used to delete a forum topic along with all its messages in a forum supergroup chat.
// The bot must be an administrator in the chat for this to work and must have the can_manage_topics administrator rights.
func (a API) DeleteForumTopic(ctx context.Context, chatID, messageThreadID int64) (res APIResponseBool, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	vals.Set("message_thread_id", itoa(messageThreadID))
	return res, a.client.get(ctx, a.base, "deleteForumTopic", vals, &res)
}

// UnpinAllForumTopicMessages is used to clear the list of pinned messages in a forum topic.
// The bot must be an administrator in the chat for this to work and must have the can_manage_topics administrator rights.
func (a API) UnpinAllForumTopicMessages(ctx context.Context, chatID, messageThreadID int64) (res APIResponseBool, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	vals.Set("message_thread_id", itoa(messageThreadID))
	return res, a.client.get(ctx, a.base, "unpinAllForumTopicMessages", vals, &res)
}

// EditGeneralForumTopic is used to edit the name of the 'General' topic in a forum supergroup chat.
// The bot must be an administrator in the chat for this to work and must have can_manage_topics administrator rights.
func (a API) EditGeneralForumTopic(ctx context.Context, chatID int64, name string) (res APIResponseBool, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	vals.Set("name", name)
	return res, a.client.get(ctx, a.base, "editGeneralForumTopic", vals, &res)
}

// CloseGeneralForumTopic is used to close an open 'General' topic in a forum supergroup chat.
// The bot must be an administrator in the chat for this to work and must have can_manage_topics administrator rights.
func (a API) CloseGeneralForumTopic(ctx context.Context, chatID int64) (res APIResponseBool, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	return res, a.client.get(ctx, a.base, "closeGeneralForumTopic", vals, &res)
}

// ReopenGeneralForumTopic is used to reopen a closed 'General' topic in a forum supergroup chat.
// The bot must be an administrator in the chat for this to work and must have can_manage_topics administrator rights.
// The topic will be automatically unhidden if it was hidden.
func (a API) ReopenGeneralForumTopic(ctx context.Context, chatID int64) (res APIResponseBool, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	return res, a.client.get(ctx, a.base, "reopenGeneralForumTopic", vals, &res)
}

// HideGeneralForumTopic is used to hide the 'General' topic in a forum supergroup chat.
// The bot must be an administrator in the chat for this to work and must have can_manage_topics administrator rights.
// The topic will be automatically closed if it was open.
func (a API) HideGeneralForumTopic(ctx context.Context, chatID int64) (res APIResponseBool, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	return res, a.client.get(ctx, a.base, "hideGeneralForumTopic", vals, &res)
}

// UnhideGeneralForumTopic is used to unhide the 'General' topic in a forum supergroup chat.
// The bot must be an administrator in the chat for this to work and must have can_manage_topics administrator rights.
func (a API) UnhideGeneralForumTopic(ctx context.Context, chatID int64) (res APIResponseBool, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	return res, a.client.get(ctx, a.base, "unhideGeneralForumTopic", vals, &res)
}

// UnpinAllGeneralForumTopicMessages is used to clear the list of pinned messages in a General forum topic.
// The bot must be an administrator in the chat for this to work and must have can_pin_messages administrator right in the supergroup.
func (a API) UnpinAllGeneralForumTopicMessages(ctx context.Context, chatID int64) (res APIResponseBool, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	return res, a.client.get(ctx, a.base, "unpinAllGeneralForumTopicMessages", vals, &res)
}

// AnswerCallbackQuery is used to send answers to callback queries sent from inline keyboards.
// The answer will be displayed to the user as a notification at the top of the chat screen or as an alert.
func (a API) AnswerCallbackQuery(ctx context.Context, callbackID string, opts *CallbackQueryOptions) (res APIResponseBool, err error) {
	vals := make(url.Values)

	vals.Set("callback_query_id", callbackID)
	return res, a.client.get(ctx, a.base, "answerCallbackQuery", addValues(vals, opts), &res)
}

// GetUserChatBoosts is used to get the list of boosts added to a chat by a user.
// Requires administrator rights in the chat.
func (a API) GetUserChatBoosts(ctx context.Context, chatID, userID int64) (res APIResponseUserChatBoosts, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	vals.Set("user_id", itoa(userID))
	return res, a.client.get(ctx, a.base, "getUserChatBoosts", vals, &res)
}

// GetBusinessConnection is used to get information about the connection of the bot with a business account.
func (a API) GetBusinessConnection(ctx context.Context, business_connection_id string) (res APIResponseBusinessConnection, err error) {
	vals := make(url.Values)

	vals.Set("business_connection_id", business_connection_id)
	return res, a.client.get(ctx, a.base, "getBusinessConnection", vals, &res)
}

// SetMyCommands is used to change the list of the bot's commands for the given scope and user language.
func (a API) SetMyCommands(ctx context.Context, opts *CommandOptions, commands ...BotCommand) (res APIResponseBool, err error) {
	vals := make(url.Values)

	jsn, _ := json.Marshal(commands)
	vals.Set("commands", string(jsn))
	return res, a.client.get(ctx, a.base, "setMyCommands", addValues(vals, opts), &res)
}

// DeleteMyCommands is used to delete the list of the bot's commands for the given scope and user language.
func (a API) DeleteMyCommands(ctx context.Context, opts *CommandOptions) (res APIResponseBool, err error) {
	return res, a.client.get(ctx, a.base, "deleteMyCommands", urlValues(opts), &res)
}

// GetMyCommands is used to get the current list of the bot's commands for the given scope and user language.
func (a API) GetMyCommands(ctx context.Context, opts *CommandOptions) (res APIResponseCommands, err error) {
	return res, a.client.get(ctx, a.base, "getMyCommands", urlValues(opts), &res)
}

// SetMyName is used to change the bot's name.
func (a API) SetMyName(ctx context.Context, name, languageCode string) (res APIResponseBool, err error) {
	vals := make(url.Values)

	vals.Set("name", name)
	vals.Set("language_code", languageCode)
	return res, a.client.get(ctx, a.base, "setMyName", vals, &res)
}

// GetMyName is used to get the current bot name for the given user language.
func (a API) GetMyName(ctx context.Context, languageCode string) (res APIResponseBotName, err error) {
	vals := make(url.Values)

	vals.Set("language_code", languageCode)
	return res, a.client.get(ctx, a.base, "getMyName", vals, &res)
}

// SetMyDescription is used to to change the bot's description, which is shown in the chat with the bot if the chat is empty.
func (a API) SetMyDescription(ctx context.Context, description, languageCode string) (res APIResponseBool, err error) {
	vals := make(url.Values)

	vals.Set("description", description)
	vals.Set("language_code", languageCode)
	return res, a.client.get(ctx, a.base, "setMyDescription", vals, &res)
}

// GetMyDescription is used to get the current bot description for the given user language.
func (a API) GetMyDescription(ctx context.Context, languageCode string) (res APIResponseBotDescription, err error) {
	vals := make(url.Values)

	vals.Set("language_code", languageCode)
	return res, a.client.get(ctx, a.base, "getMyDescription", vals, &res)
}

// SetMyShortDescription is used to to change the bot's short description,
// which is shown on the bot's profile page and is sent together with the link when users share the bot.
func (a API) SetMyShortDescription(ctx context.Context, shortDescription, languageCode string) (res APIResponseBool, err error) {
	vals := make(url.Values)

	vals.Set("short_description", shortDescription)
	vals.Set("language_code", languageCode)
	return res, a.client.get(ctx, a.base, "setMyShortDescription", vals, &res)
}

// GetMyShortDescription is used to get the current bot short description for the given user language.
func (a API) GetMyShortDescription(ctx context.Context, languageCode string) (res APIResponseBotShortDescription, err error) {
	vals := make(url.Values)

	vals.Set("language_code", languageCode)
	return res, a.client.get(ctx, a.base, "getMyDescription", vals, &res)
}

// EditMessageText is used to edit text and game messages.
func (a API) EditMessageText(ctx context.Context, text string, msg MessageIDOptions, opts *MessageTextOptions) (res APIResponseMessage, err error) {
	vals := make(url.Values)

	vals.Set("text", text)
	return res, a.client.get(ctx, a.base, "editMessageText", addValues(addValues(vals, msg), opts), &res)
}

// EditMessageCaption is used to edit captions of messages.
func (a API) EditMessageCaption(ctx context.Context, msg MessageIDOptions, opts *MessageCaptionOptions) (res APIResponseMessage, err error) {
	return res, a.client.get(ctx, a.base, "editMessageCaption", addValues(urlValues(msg), opts), &res)
}

// EditMessageMedia is used to edit animation, audio, document, photo or video messages.
// If a message is part of a message album, then it can be edited only to an audio for audio albums,
// only to a document for document albums and to a photo or a video otherwise.
// When an inline message is edited, a new file can't be uploaded.
// Use a previously uploaded file via its file_id or specify a URL.
func (a API) EditMessageMedia(ctx context.Context, msg MessageIDOptions, media InputMedia, opts *MessageMediaOptions) (res APIResponseMessage, err error) {
	return res, a.client.postMedia(ctx, a.base, "editMessageMedia", true, addValues(urlValues(msg), opts), &res, media)
}

// EditMessageReplyMarkup is used to edit only the reply markup of messages.
func (a API) EditMessageReplyMarkup(ctx context.Context, msg MessageIDOptions, opts *MessageReplyMarkupOptions) (res APIResponseMessage, err error) {
	return res, a.client.get(ctx, a.base, "editMessageReplyMarkup", addValues(urlValues(msg), opts), &res)
}

// StopPoll is used to stop a poll which was sent by the bot.
func (a API) StopPoll(ctx context.Context, chatID int64, messageID int, opts *StopPollOptions) (res APIResponsePoll, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	vals.Set("message_id", itoa(int64(messageID)))
	return res, a.client.get(ctx, a.base, "stopPoll", addValues(vals, opts), &res)
}

// DeleteMessage is used to delete a message, including service messages, with the following limitations:
// - A message can only be deleted if it was sent less than 48 hours ago.
// - A dice message in a private chat can only be deleted if it was sent more than 24 hours ago.
// - Bots can delete outgoing messages in private chats, groups, and supergroups.
// - Bots can delete incoming messages in private chats.
// - Bots granted can_post_messages permissions can delete outgoing messages in channels.
// - If the bot is an administrator of a group, it can delete any message there.
// - If the bot has can_delete_messages permission in a supergroup or a channel, it can delete any message there.
func (a API) DeleteMessage(ctx context.Context, chatID int64, messageID int) (res APIResponseBase, err error) {
	vals := make(url.Values)

	vals.Set("chat_id", itoa(chatID))
	vals.Set("message_id", itoa(int64(messageID)))
	return res, a.client.get(ctx, a.base, "deleteMessage", vals, &res)
}

// DeleteMessages is used to delete multiple messages simultaneously.
// If some of the specified messages can't be found, they are skipped.
func (a API) DeleteMessages(ctx context.Context, chatID int64, messageIDs []int) (res APIResponseBool, err error) {
	vals := make(url.Values)

	msgIDs, err := json.Marshal(messageIDs)
	if err != nil {
		return res, err
	}

	vals.Set("chat_id", itoa(chatID))
	vals.Set("message_ids", string(msgIDs))
	return res, a.client.get(ctx, a.base, "deleteMessages", vals, &res)
}
