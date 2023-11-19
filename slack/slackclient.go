package slack

import (
	"context"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

type socketClient interface {
	Ack(req socketmode.Request, payload ...interface{})
	// Debugf(format string, v ...interface{})
	// Debugln(v ...interface{})
	// Open() (info *slack.SocketModeConnection, websocketURL string, err error)
	// OpenContext(ctx context.Context) (info *slack.SocketModeConnection, websocketURL string, err error)
	//Run() error
	RunContext(ctx context.Context) error
	//Send(res socketmode.Response)

	// AddBookmark(channelID string, params slack.AddBookmarkParameters) (slack.Bookmark, error)
	// AddBookmarkContext(ctx context.Context, channelID string, params AddBookmarkParameters) (Bookmark, error)
	// AddChannelReminder(channelID string, text string, time string) (*Reminder, error)
	// AddChannelReminderContext(ctx context.Context, channelID string, text string, time string) (*Reminder, error)
	// AddPin(channel string, item ItemRef) error
	// AddPinContext(ctx context.Context, channel string, item ItemRef) error
	// AddReaction(name string, item ItemRef) error
	// AddReactionContext(ctx context.Context, name string, item ItemRef) error
	// AddRemoteFile(params RemoteFileParameters) (*RemoteFile, error)
	// AddRemoteFileContext(ctx context.Context, params RemoteFileParameters) (remotefile *RemoteFile, err error)
	// AddStar(channel string, item ItemRef) error
	// AddStarContext(ctx context.Context, channel string, item ItemRef) error
	// AddUserReminder(userID string, text string, time string) (*Reminder, error)
	// AddUserReminderContext(ctx context.Context, userID string, text string, time string) (*Reminder, error)
	// ArchiveConversation(channelID string) error
	// ArchiveConversationContext(ctx context.Context, channelID string) error
	// AuthTest() (response *AuthTestResponse, error error)
	// AuthTestContext(ctx context.Context) (response *AuthTestResponse, err error)
	// CloseConversation(channelID string) (noOp bool, alreadyClosed bool, err error)
	// CloseConversationContext(ctx context.Context, channelID string) (noOp bool, alreadyClosed bool, err error)
	// ConnectRTM() (info *Info, websocketURL string, err error)
	// ConnectRTMContext(ctx context.Context) (info *Info, websocketURL string, err error)
	// CreateConversation(params CreateConversationParams) (*Channel, error)
	// CreateConversationContext(ctx context.Context, params CreateConversationParams) (*Channel, error)
	// CreateUserGroup(userGroup UserGroup) (UserGroup, error)
	// CreateUserGroupContext(ctx context.Context, userGroup UserGroup) (UserGroup, error)
	// Debug() bool
	// Debugf(format string, v ...interface{})
	// Debugln(v ...interface{})
	// DeleteFile(fileID string) error
	// DeleteFileComment(commentID string, fileID string) error
	// DeleteFileCommentContext(ctx context.Context, fileID string, commentID string) (err error)
	// DeleteFileContext(ctx context.Context, fileID string) (err error)
	// DeleteMessage(channel string, messageTimestamp string) (string, string, error)
	// DeleteMessageContext(ctx context.Context, channel string, messageTimestamp string) (string, string, error)
	// DeleteReminder(id string) error
	// DeleteReminderContext(ctx context.Context, id string) error
	// DeleteScheduledMessage(params *DeleteScheduledMessageParameters) (bool, error)
	// DeleteScheduledMessageContext(ctx context.Context, params *DeleteScheduledMessageParameters) (bool, error)
	// DeleteUserPhoto() error
	// DeleteUserPhotoContext(ctx context.Context) (err error)
	// DisableUser(teamName string, uid string) error
	// DisableUserContext(ctx context.Context, teamName string, uid string) error
	// DisableUserGroup(userGroup string) (UserGroup, error)
	// DisableUserGroupContext(ctx context.Context, userGroup string) (UserGroup, error)
	// EditBookmark(channelID string, bookmarkID string, params EditBookmarkParameters) (Bookmark, error)
	// EditBookmarkContext(ctx context.Context, channelID string, bookmarkID string, params EditBookmarkParameters) (Bookmark, error)
	// EnableUserGroup(userGroup string) (UserGroup, error)
	// EnableUserGroupContext(ctx context.Context, userGroup string) (UserGroup, error)
	// EndDND() error
	// EndDNDContext(ctx context.Context) error
	// EndSnooze() (*DNDStatus, error)
	// EndSnoozeContext(ctx context.Context) (*DNDStatus, error)
	// GetAccessLogs(params AccessLogParameters) ([]Login, *Paging, error)
	// GetAccessLogsContext(ctx context.Context, params AccessLogParameters) ([]Login, *Paging, error)
	// GetAuditLogs(params AuditLogParameters) (entries []AuditEntry, nextCursor string, err error)
	// GetAuditLogsContext(ctx context.Context, params AuditLogParameters) (entries []AuditEntry, nextCursor string, err error)
	// GetBillableInfo(user string) (map[string]BillingActive, error)
	// GetBillableInfoContext(ctx context.Context, user string) (map[string]BillingActive, error)
	// GetBillableInfoForTeam() (map[string]BillingActive, error)
	// GetBillableInfoForTeamContext(ctx context.Context) (map[string]BillingActive, error)
	// GetBotInfo(bot string) (*Bot, error)
	// GetBotInfoContext(ctx context.Context, bot string) (*Bot, error)
	// GetConversationHistory(params *GetConversationHistoryParameters) (*GetConversationHistoryResponse, error)
	// GetConversationHistoryContext(ctx context.Context, params *GetConversationHistoryParameters) (*GetConversationHistoryResponse, error)
	//GetConversationInfo(input *slack.GetConversationInfoInput) (*slack.Channel, error)
	GetConversationInfoContext(ctx context.Context, input *slack.GetConversationInfoInput) (*slack.Channel, error)
	// GetConversationReplies(params *GetConversationRepliesParameters) (msgs []Message, hasMore bool, nextCursor string, err error)
	// GetConversationRepliesContext(ctx context.Context, params *GetConversationRepliesParameters) (msgs []Message, hasMore bool, nextCursor string, err error)
	// GetConversations(params *GetConversationsParameters) (channels []Channel, nextCursor string, err error)
	// GetConversationsContext(ctx context.Context, params *GetConversationsParameters) (channels []Channel, nextCursor string, err error)
	// GetConversationsForUser(params *GetConversationsForUserParameters) (channels []Channel, nextCursor string, err error)
	// GetConversationsForUserContext(ctx context.Context, params *GetConversationsForUserParameters) (channels []Channel, nextCursor string, err error)
	// GetDNDInfo(user *string) (*DNDStatus, error)
	// GetDNDInfoContext(ctx context.Context, user *string) (*DNDStatus, error)
	// GetDNDTeamInfo(users []string) (map[string]DNDStatus, error)
	// GetDNDTeamInfoContext(ctx context.Context, users []string) (map[string]DNDStatus, error)
	// GetEmoji() (map[string]string, error)
	// GetEmojiContext(ctx context.Context) (map[string]string, error)
	// GetFile(downloadURL string, writer io.Writer) error
	// GetFileContext(ctx context.Context, downloadURL string, writer io.Writer) error
	// GetFileInfo(fileID string, count int, page int) (*File, []Comment, *Paging, error)
	// GetFileInfoContext(ctx context.Context, fileID string, count int, page int) (*File, []Comment, *Paging, error)
	// GetFiles(params GetFilesParameters) ([]File, *Paging, error)
	// GetFilesContext(ctx context.Context, params GetFilesParameters) ([]File, *Paging, error)
	// GetOtherTeamInfo(team string) (*TeamInfo, error)
	// GetOtherTeamInfoContext(ctx context.Context, team string) (*TeamInfo, error)
	// GetPermalink(params *PermalinkParameters) (string, error)
	// GetPermalinkContext(ctx context.Context, params *PermalinkParameters) (string, error)
	// GetReactions(item ItemRef, params GetReactionsParameters) ([]ItemReaction, error)
	// GetReactionsContext(ctx context.Context, item ItemRef, params GetReactionsParameters) ([]ItemReaction, error)
	// GetRemoteFileInfo(externalID string, fileID string) (remotefile *RemoteFile, err error)
	// GetRemoteFileInfoContext(ctx context.Context, externalID string, fileID string) (remotefile *RemoteFile, err error)
	// GetScheduledMessages(params *GetScheduledMessagesParameters) (channels []ScheduledMessage, nextCursor string, err error)
	// GetScheduledMessagesContext(ctx context.Context, params *GetScheduledMessagesParameters) (channels []ScheduledMessage, nextCursor string, err error)
	// GetStarred(params StarsParameters) ([]StarredItem, *Paging, error)
	// GetStarredContext(ctx context.Context, params StarsParameters) ([]StarredItem, *Paging, error)
	// GetTeamInfo() (*TeamInfo, error)
	// GetTeamInfoContext(ctx context.Context) (*TeamInfo, error)
	// GetTeamProfile() (*TeamProfile, error)
	// GetTeamProfileContext(ctx context.Context) (*TeamProfile, error)
	// GetUserByEmail(email string) (*User, error)
	// GetUserByEmailContext(ctx context.Context, email string) (*User, error)
	// GetUserGroupMembers(userGroup string) ([]string, error)
	// GetUserGroupMembersContext(ctx context.Context, userGroup string) ([]string, error)
	// GetUserGroups(options ...GetUserGroupsOption) ([]UserGroup, error)
	// GetUserGroupsContext(ctx context.Context, options ...GetUserGroupsOption) ([]UserGroup, error)
	// GetUserIdentity() (*UserIdentityResponse, error)
	// GetUserIdentityContext(ctx context.Context) (response *UserIdentityResponse, err error)
	//GetUserInfo(user string) (*slack.User, error)
	GetUserInfoContext(ctx context.Context, user string) (*slack.User, error)
	// GetUserPrefs() (*UserPrefsCarrier, error)
	// GetUserPrefsContext(ctx context.Context) (*UserPrefsCarrier, error)
	// GetUserPresence(user string) (*UserPresence, error)
	// GetUserPresenceContext(ctx context.Context, user string) (*UserPresence, error)
	// GetUserProfile(params *GetUserProfileParameters) (*UserProfile, error)
	// GetUserProfileContext(ctx context.Context, params *GetUserProfileParameters) (*UserProfile, error)
	// GetUsers(options ...GetUsersOption) ([]User, error)
	// GetUsersContext(ctx context.Context, options ...GetUsersOption) (results []User, err error)
	// GetUsersInConversation(params *GetUsersInConversationParameters) ([]string, string, error)
	// GetUsersInConversationContext(ctx context.Context, params *GetUsersInConversationParameters) ([]string, string, error)
	// GetUsersInfo(users ...string) (*[]User, error)
	// GetUsersInfoContext(ctx context.Context, users ...string) (*[]User, error)
	// GetUsersPaginated(options ...GetUsersOption) UserPagination
	// InviteGuest(teamName string, channel string, firstName string, lastName string, emailAddress string) error
	// InviteGuestContext(ctx context.Context, teamName string, channel string, firstName string, lastName string, emailAddress string) error
	// InviteRestricted(teamName string, channel string, firstName string, lastName string, emailAddress string) error
	// InviteRestrictedContext(ctx context.Context, teamName string, channel string, firstName string, lastName string, emailAddress string) error
	// InviteToTeam(teamName string, firstName string, lastName string, emailAddress string) error
	// InviteToTeamContext(ctx context.Context, teamName string, firstName string, lastName string, emailAddress string) error
	// InviteUsersToConversation(channelID string, users ...string) (*Channel, error)
	// InviteUsersToConversationContext(ctx context.Context, channelID string, users ...string) (*Channel, error)
	//JoinConversation(channelID string) (*slack.Channel, string, []string, error)
	JoinConversationContext(ctx context.Context, channelID string) (*slack.Channel, string, []string, error)
	// KickUserFromConversation(channelID string, user string) error
	// KickUserFromConversationContext(ctx context.Context, channelID string, user string) error
	// LeaveConversation(channelID string) (bool, error)
	// LeaveConversationContext(ctx context.Context, channelID string) (bool, error)
	// ListAllStars() ([]Item, error)
	// ListAllStarsContext(ctx context.Context) (results []Item, err error)
	// ListBookmarks(channelID string) ([]Bookmark, error)
	// ListBookmarksContext(ctx context.Context, channelID string) ([]Bookmark, error)
	// ListEventAuthorizations(eventContext string) ([]EventAuthorization, error)
	// ListEventAuthorizationsContext(ctx context.Context, eventContext string) ([]EventAuthorization, error)
	// ListFiles(params ListFilesParameters) ([]File, *ListFilesParameters, error)
	// ListFilesContext(ctx context.Context, params ListFilesParameters) ([]File, *ListFilesParameters, error)
	// ListPins(channel string) ([]Item, *Paging, error)
	// ListPinsContext(ctx context.Context, channel string) ([]Item, *Paging, error)
	// ListReactions(params ListReactionsParameters) ([]ReactedItem, *Paging, error)
	// ListReactionsContext(ctx context.Context, params ListReactionsParameters) ([]ReactedItem, *Paging, error)
	// ListReminders() ([]*Reminder, error)
	// ListRemindersContext(ctx context.Context) ([]*Reminder, error)
	// ListRemoteFiles(params ListRemoteFilesParameters) ([]RemoteFile, error)
	// ListRemoteFilesContext(ctx context.Context, params ListRemoteFilesParameters) ([]RemoteFile, error)
	// ListStars(params StarsParameters) ([]Item, *Paging, error)
	// ListStarsContext(ctx context.Context, params StarsParameters) ([]Item, *Paging, error)
	// ListStarsPaginated(options ...ListStarsOption) StarredItemPagination
	// ListTeams(params ListTeamsParameters) ([]Team, string, error)
	// ListTeamsContext(ctx context.Context, params ListTeamsParameters) ([]Team, string, error)
	// MarkConversation(channel string, ts string) (err error)
	// MarkConversationContext(ctx context.Context, channel string, ts string) error
	// MuteChat(channelID string) (*UserPrefsCarrier, error)
	// NewRTM(options ...RTMOption) *RTM
	// OpenConversation(params *OpenConversationParameters) (*Channel, bool, bool, error)
	// OpenConversationContext(ctx context.Context, params *OpenConversationParameters) (*Channel, bool, bool, error)
	// OpenDialog(triggerID string, dialog Dialog) (err error)
	// OpenDialogContext(ctx context.Context, triggerID string, dialog Dialog) (err error)
	// OpenView(triggerID string, view slack.ModalViewRequest) (*slack.ViewResponse, error)
	OpenViewContext(ctx context.Context, triggerID string, view slack.ModalViewRequest) (*slack.ViewResponse, error)
	// PostEphemeral(channelID string, userID string, options ...MsgOption) (string, error)
	// PostEphemeralContext(ctx context.Context, channelID string, userID string, options ...MsgOption) (timestamp string, err error)
	// PostMessage(channelID string, options ...MsgOption) (string, string, error)
	// PostMessageContext(ctx context.Context, channelID string, options ...MsgOption) (string, string, error)
	// PublishView(userID string, view HomeTabViewRequest, hash string) (*ViewResponse, error)
	// PublishViewContext(ctx context.Context, userID string, view HomeTabViewRequest, hash string) (*ViewResponse, error)
	// PushView(triggerID string, view ModalViewRequest) (*ViewResponse, error)
	// PushViewContext(ctx context.Context, triggerID string, view ModalViewRequest) (*ViewResponse, error)
	// RemoveBookmark(channelID string, bookmarkID string) error
	// RemoveBookmarkContext(ctx context.Context, channelID string, bookmarkID string) error
	// RemovePin(channel string, item ItemRef) error
	// RemovePinContext(ctx context.Context, channel string, item ItemRef) error
	// RemoveReaction(name string, item ItemRef) error
	// RemoveReactionContext(ctx context.Context, name string, item ItemRef) error
	// RemoveRemoteFile(externalID string, fileID string) (err error)
	// RemoveRemoteFileContext(ctx context.Context, externalID string, fileID string) (err error)
	// RemoveStar(channel string, item ItemRef) error
	// RemoveStarContext(ctx context.Context, channel string, item ItemRef) error
	// RenameConversation(channelID string, channelName string) (*Channel, error)
	// RenameConversationContext(ctx context.Context, channelID string, channelName string) (*Channel, error)
	// RevokeFilePublicURL(fileID string) (*File, error)
	// RevokeFilePublicURLContext(ctx context.Context, fileID string) (*File, error)
	// SaveWorkflowStepConfiguration(workflowStepEditID string, inputs *WorkflowStepInputs, outputs *[]WorkflowStepOutput) error
	// SaveWorkflowStepConfigurationContext(ctx context.Context, workflowStepEditID string, inputs *WorkflowStepInputs, outputs *[]WorkflowStepOutput) error
	// ScheduleMessage(channelID string, postAt string, options ...MsgOption) (string, string, error)
	// ScheduleMessageContext(ctx context.Context, channelID string, postAt string, options ...MsgOption) (string, string, error)
	// Search(query string, params SearchParameters) (*SearchMessages, *SearchFiles, error)
	// SearchContext(ctx context.Context, query string, params SearchParameters) (*SearchMessages, *SearchFiles, error)
	// SearchFiles(query string, params SearchParameters) (*SearchFiles, error)
	// SearchFilesContext(ctx context.Context, query string, params SearchParameters) (*SearchFiles, error)
	// SearchMessages(query string, params SearchParameters) (*SearchMessages, error)
	// SearchMessagesContext(ctx context.Context, query string, params SearchParameters) (*SearchMessages, error)
	// SendAuthRevoke(token string) (*AuthRevokeResponse, error)
	// SendAuthRevokeContext(ctx context.Context, token string) (*AuthRevokeResponse, error)
	//SendMessage(channel string, options ...slack.MsgOption) (string, string, string, error)
	SendMessageContext(ctx context.Context, channelID string, options ...slack.MsgOption) (_channel string, _timestamp string, _text string, err error)
	// SendSSOBindingEmail(teamName string, user string) error
	// SendSSOBindingEmailContext(ctx context.Context, teamName string, user string) error
	// SetPurposeOfConversation(channelID string, purpose string) (*Channel, error)
	// SetPurposeOfConversationContext(ctx context.Context, channelID string, purpose string) (*Channel, error)
	// SetRegular(teamName string, user string) error
	// SetRegularContext(ctx context.Context, teamName string, user string) error
	// SetRestricted(teamName string, uid string, channelIds ...string) error
	// SetRestrictedContext(ctx context.Context, teamName string, uid string, channelIds ...string) error
	// SetSnooze(minutes int) (*DNDStatus, error)
	// SetSnoozeContext(ctx context.Context, minutes int) (*DNDStatus, error)
	// SetTopicOfConversation(channelID string, topic string) (*Channel, error)
	// SetTopicOfConversationContext(ctx context.Context, channelID string, topic string) (*Channel, error)
	// SetUltraRestricted(teamName string, uid string, channel string) error
	// SetUltraRestrictedContext(ctx context.Context, teamName string, uid string, channel string) error
	// SetUserAsActive() error
	// SetUserAsActiveContext(ctx context.Context) (err error)
	// SetUserCustomFields(userID string, customFields map[string]UserProfileCustomField) error
	// SetUserCustomFieldsContext(ctx context.Context, userID string, customFields map[string]UserProfileCustomField) error
	// SetUserCustomStatus(statusText string, statusEmoji string, statusExpiration int64) error
	// SetUserCustomStatusContext(ctx context.Context, statusText string, statusEmoji string, statusExpiration int64) error
	// SetUserCustomStatusContextWithUser(ctx context.Context, user string, statusText string, statusEmoji string, statusExpiration int64) error
	// SetUserCustomStatusWithUser(user string, statusText string, statusEmoji string, statusExpiration int64) error
	// SetUserPhoto(image string, params UserSetPhotoParams) error
	// SetUserPhotoContext(ctx context.Context, image string, params UserSetPhotoParams) (err error)
	// SetUserPresence(presence string) error
	// SetUserPresenceContext(ctx context.Context, presence string) error
	// SetUserRealName(realName string) error
	// SetUserRealNameContextWithUser(ctx context.Context, user string, realName string) error
	// ShareFilePublicURL(fileID string) (*File, []Comment, *Paging, error)
	// ShareFilePublicURLContext(ctx context.Context, fileID string) (*File, []Comment, *Paging, error)
	// ShareRemoteFile(channels []string, externalID string, fileID string) (file *RemoteFile, err error)
	// ShareRemoteFileContext(ctx context.Context, channels []string, externalID string, fileID string) (file *RemoteFile, err error)
	// StartRTM() (info *Info, websocketURL string, err error)
	// StartRTMContext(ctx context.Context) (info *Info, websocketURL string, err error)
	// StartSocketModeContext(ctx context.Context) (info *SocketModeConnection, websocketURL string, err error)
	// UnArchiveConversation(channelID string) error
	// UnArchiveConversationContext(ctx context.Context, channelID string) error
	// UnMuteChat(channelID string) (*UserPrefsCarrier, error)
	// UnfurlMessage(channelID string, timestamp string, unfurls map[string]Attachment, options ...MsgOption) (string, string, string, error)
	// UnfurlMessageContext(ctx context.Context, channelID string, timestamp string, unfurls map[string]Attachment, options ...MsgOption) (string, string, string, error)
	// UnfurlMessageWithAuthURL(channelID string, timestamp string, userAuthURL string, options ...MsgOption) (string, string, string, error)
	// UnfurlMessageWithAuthURLContext(ctx context.Context, channelID string, timestamp string, userAuthURL string, options ...MsgOption) (string, string, string, error)
	// UninstallApp(clientID string, clientSecret string) error
	// UninstallAppContext(ctx context.Context, clientID string, clientSecret string) error
	// UnsetUserCustomStatus() error
	// UnsetUserCustomStatusContext(ctx context.Context) error
	// UpdateMessage(channelID string, timestamp string, options ...MsgOption) (string, string, string, error)
	// UpdateMessageContext(ctx context.Context, channelID string, timestamp string, options ...MsgOption) (string, string, string, error)
	// UpdateRemoteFile(fileID string, params RemoteFileParameters) (remotefile *RemoteFile, err error)
	// UpdateRemoteFileContext(ctx context.Context, fileID string, params RemoteFileParameters) (remotefile *RemoteFile, err error)
	// UpdateUserGroup(userGroupID string, options ...UpdateUserGroupsOption) (UserGroup, error)
	// UpdateUserGroupContext(ctx context.Context, userGroupID string, options ...UpdateUserGroupsOption) (UserGroup, error)
	// UpdateUserGroupMembers(userGroup string, members string) (UserGroup, error)
	// UpdateUserGroupMembersContext(ctx context.Context, userGroup string, members string) (UserGroup, error)
	// UpdateView(view slack.ModalViewRequest, externalID string, hash string, viewID string) (*slack.ViewResponse, error)
	UpdateViewContext(ctx context.Context, view slack.ModalViewRequest, externalID string, hash string, viewID string) (*slack.ViewResponse, error)
	// UploadFile(params FileUploadParameters) (file *File, err error)
	// UploadFileContext(ctx context.Context, params FileUploadParameters) (file *File, err error)
	// UploadFileV2(params UploadFileV2Parameters) (*FileSummary, error)
	// UploadFileV2Context(ctx context.Context, params UploadFileV2Parameters) (file *FileSummary, err error)
	// WorkflowStepCompleted(workflowStepExecuteID string, options ...WorkflowStepCompletedRequestOption) error
	// WorkflowStepFailed(workflowStepExecuteID string, errorMessage string) error
}
