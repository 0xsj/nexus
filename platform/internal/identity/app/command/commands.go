// Package command contains the command definitions and handlers for the Identity context.
package command

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// ============================================================================
// Command Name Constants
// ============================================================================

const (
	// Registration commands
	CommandRegisterWithEmail  = "identity.RegisterWithEmail"
	CommandRegisterWithWallet = "identity.RegisterWithWallet"
	CommandRegisterWithOAuth  = "identity.RegisterWithOAuth"

	// Magic link commands
	CommandRequestMagicLink = "identity.RequestMagicLink"
	CommandVerifyMagicLink  = "identity.VerifyMagicLink"

	// Wallet authentication commands
	CommandAuthenticateWithWallet = "identity.AuthenticateWithWallet"

	// OAuth commands
	CommandInitiateOAuth         = "identity.InitiateOAuth"
	CommandAuthenticateWithOAuth = "identity.AuthenticateWithOAuth"

	// Session commands
	CommandRefreshSession        = "identity.RefreshSession"
	CommandRevokeSession         = "identity.RevokeSession"
	CommandRevokeAllUserSessions = "identity.RevokeAllUserSessions"

	// Profile commands
	CommandChangeDisplayName = "identity.ChangeDisplayName"
	CommandChangeEmail       = "identity.ChangeEmail"

	// User status commands
	CommandActivateUser   = "identity.ActivateUser"
	CommandSuspendUser    = "identity.SuspendUser"
	CommandReactivateUser = "identity.ReactivateUser"
	CommandDeleteUser     = "identity.DeleteUser"

	// DID management commands
	CommandAddDID    = "identity.AddDID"
	CommandRemoveDID = "identity.RemoveDID"

	// OAuth linking commands
	CommandLinkOAuthAccount   = "identity.LinkOAuthAccount"
	CommandUnlinkOAuthAccount = "identity.UnlinkOAuthAccount"
)

// ============================================================================
// Registration Commands
// ============================================================================

// RegisterWithEmail registers a new user via email (magic link flow).
type RegisterWithEmail struct {
	Email       types.Email `json:"email" validate:"required"`
	DisplayName string      `json:"display_name" validate:"required,min=1,max=100"`
}

// CommandName implements cqrs.Command.
func (c RegisterWithEmail) CommandName() string {
	return CommandRegisterWithEmail
}

// Validate implements cqrs.Validatable.
func (c RegisterWithEmail) Validate() error {
	if c.Email.IsEmpty() {
		return cqrs.ErrCommandValidation("RegisterWithEmail.Validate", "email is required")
	}
	if c.DisplayName == "" {
		return cqrs.ErrCommandValidation("RegisterWithEmail.Validate", "display_name is required")
	}
	if len(c.DisplayName) > 100 {
		return cqrs.ErrCommandValidation("RegisterWithEmail.Validate", "display_name must be 100 characters or less")
	}
	return nil
}

// RegisterWithEmailResult is the result data for RegisterWithEmail.
type RegisterWithEmailResult struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// RegisterWithWallet registers a new user via wallet signature (SIWE).
type RegisterWithWallet struct {
	Address     string  `json:"address" validate:"required"`
	Message     string  `json:"message" validate:"required"`
	Signature   string  `json:"signature" validate:"required"`
	ChainID     string  `json:"chain_id" validate:"required"`
	DisplayName *string `json:"display_name" validate:"omitempty,min=1,max=100"`
}

// CommandName implements cqrs.Command.
func (c RegisterWithWallet) CommandName() string {
	return CommandRegisterWithWallet
}

// Validate implements cqrs.Validatable.
func (c RegisterWithWallet) Validate() error {
	if c.Address == "" {
		return cqrs.ErrCommandValidation("RegisterWithWallet.Validate", "address is required")
	}
	if c.Message == "" {
		return cqrs.ErrCommandValidation("RegisterWithWallet.Validate", "message is required")
	}
	if c.Signature == "" {
		return cqrs.ErrCommandValidation("RegisterWithWallet.Validate", "signature is required")
	}
	if c.ChainID == "" {
		return cqrs.ErrCommandValidation("RegisterWithWallet.Validate", "chain_id is required")
	}
	if c.DisplayName != nil && len(*c.DisplayName) > 100 {
		return cqrs.ErrCommandValidation("RegisterWithWallet.Validate", "display_name must be 100 characters or less")
	}
	return nil
}

// RegisterWithWalletResult is the result data for RegisterWithWallet.
type RegisterWithWalletResult struct {
	UserID    string    `json:"user_id"`
	DID       string    `json:"did"`
	CreatedAt time.Time `json:"created_at"`
}

// RegisterWithOAuth registers a new user via OAuth provider.
type RegisterWithOAuth struct {
	Provider    string  `json:"provider" validate:"required,oneof=google github"`
	Code        string  `json:"code" validate:"required"`
	State       string  `json:"state" validate:"required"`
	DisplayName *string `json:"display_name" validate:"omitempty,min=1,max=100"`
}

// CommandName implements cqrs.Command.
func (c RegisterWithOAuth) CommandName() string {
	return CommandRegisterWithOAuth
}

// Validate implements cqrs.Validatable.
func (c RegisterWithOAuth) Validate() error {
	if c.Provider == "" {
		return cqrs.ErrCommandValidation("RegisterWithOAuth.Validate", "provider is required")
	}
	if c.Provider != "google" && c.Provider != "github" {
		return cqrs.ErrCommandValidation("RegisterWithOAuth.Validate", "provider must be 'google' or 'github'")
	}
	if c.Code == "" {
		return cqrs.ErrCommandValidation("RegisterWithOAuth.Validate", "code is required")
	}
	if c.State == "" {
		return cqrs.ErrCommandValidation("RegisterWithOAuth.Validate", "state is required")
	}
	if c.DisplayName != nil && len(*c.DisplayName) > 100 {
		return cqrs.ErrCommandValidation("RegisterWithOAuth.Validate", "display_name must be 100 characters or less")
	}
	return nil
}

// RegisterWithOAuthResult is the result data for RegisterWithOAuth.
type RegisterWithOAuthResult struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Provider  string    `json:"provider"`
	CreatedAt time.Time `json:"created_at"`
}

// ============================================================================
// Magic Link Commands
// ============================================================================

// RequestMagicLink requests a magic link to be sent to the given email.
type RequestMagicLink struct {
	Email types.Email `json:"email" validate:"required"`
}

// CommandName implements cqrs.Command.
func (c RequestMagicLink) CommandName() string {
	return CommandRequestMagicLink
}

// Validate implements cqrs.Validatable.
func (c RequestMagicLink) Validate() error {
	if c.Email.IsEmpty() {
		return cqrs.ErrCommandValidation("RequestMagicLink.Validate", "email is required")
	}
	return nil
}

// RequestMagicLinkResult is the result data for RequestMagicLink.
type RequestMagicLinkResult struct {
	Email     string    `json:"email"`
	ExpiresAt time.Time `json:"expires_at"`
	Success   bool      `json:"success"`
}

// VerifyMagicLink verifies a magic link token and creates a session.
type VerifyMagicLink struct {
	Token     string `json:"token" validate:"required"`
	IPAddress string `json:"ip_address" validate:"omitempty"`
	UserAgent string `json:"user_agent" validate:"omitempty"`
}

// CommandName implements cqrs.Command.
func (c VerifyMagicLink) CommandName() string {
	return CommandVerifyMagicLink
}

// Validate implements cqrs.Validatable.
func (c VerifyMagicLink) Validate() error {
	if c.Token == "" {
		return cqrs.ErrCommandValidation("VerifyMagicLink.Validate", "token is required")
	}
	return nil
}

// VerifyMagicLinkResult is the result data for VerifyMagicLink.
type VerifyMagicLinkResult struct {
	UserID      string    `json:"user_id"`
	SessionID   string    `json:"session_id"`
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
	IsNewUser   bool      `json:"is_new_user"`
}

// ============================================================================
// Wallet Authentication Commands
// ============================================================================

// AuthenticateWithWallet authenticates a user via wallet signature (SIWE).
type AuthenticateWithWallet struct {
	Address   string `json:"address" validate:"required"`
	Message   string `json:"message" validate:"required"`
	Signature string `json:"signature" validate:"required"`
	ChainID   string `json:"chain_id" validate:"required"`
	IPAddress string `json:"ip_address" validate:"omitempty"`
	UserAgent string `json:"user_agent" validate:"omitempty"`
}

// CommandName implements cqrs.Command.
func (c AuthenticateWithWallet) CommandName() string {
	return CommandAuthenticateWithWallet
}

// Validate implements cqrs.Validatable.
func (c AuthenticateWithWallet) Validate() error {
	if c.Address == "" {
		return cqrs.ErrCommandValidation("AuthenticateWithWallet.Validate", "address is required")
	}
	if c.Message == "" {
		return cqrs.ErrCommandValidation("AuthenticateWithWallet.Validate", "message is required")
	}
	if c.Signature == "" {
		return cqrs.ErrCommandValidation("AuthenticateWithWallet.Validate", "signature is required")
	}
	if c.ChainID == "" {
		return cqrs.ErrCommandValidation("AuthenticateWithWallet.Validate", "chain_id is required")
	}
	return nil
}

// AuthenticateWithWalletResult is the result data for AuthenticateWithWallet.
type AuthenticateWithWalletResult struct {
	UserID      string    `json:"user_id"`
	SessionID   string    `json:"session_id"`
	AccessToken string    `json:"access_token"`
	DID         string    `json:"did"`
	ExpiresAt   time.Time `json:"expires_at"`
	IsNewUser   bool      `json:"is_new_user"`
}

// ============================================================================
// OAuth Commands
// ============================================================================

// InitiateOAuth initiates an OAuth flow by generating a state and auth URL.
type InitiateOAuth struct {
	Provider    string `json:"provider" validate:"required,oneof=google github"`
	RedirectURL string `json:"redirect_url" validate:"required,url"`
}

// CommandName implements cqrs.Command.
func (c InitiateOAuth) CommandName() string {
	return CommandInitiateOAuth
}

// Validate implements cqrs.Validatable.
func (c InitiateOAuth) Validate() error {
	if c.Provider == "" {
		return cqrs.ErrCommandValidation("InitiateOAuth.Validate", "provider is required")
	}
	if c.Provider != "google" && c.Provider != "github" {
		return cqrs.ErrCommandValidation("InitiateOAuth.Validate", "provider must be 'google' or 'github'")
	}
	if c.RedirectURL == "" {
		return cqrs.ErrCommandValidation("InitiateOAuth.Validate", "redirect_url is required")
	}
	return nil
}

// InitiateOAuthResult is the result data for InitiateOAuth.
type InitiateOAuthResult struct {
	AuthURL   string    `json:"auth_url"`
	State     string    `json:"state"`
	ExpiresAt time.Time `json:"expires_at"`
}

// AuthenticateWithOAuth authenticates a user via OAuth provider callback.
type AuthenticateWithOAuth struct {
	Provider  string `json:"provider" validate:"required,oneof=google github"`
	Code      string `json:"code" validate:"required"`
	State     string `json:"state" validate:"required"`
	IPAddress string `json:"ip_address" validate:"omitempty"`
	UserAgent string `json:"user_agent" validate:"omitempty"`
}

// CommandName implements cqrs.Command.
func (c AuthenticateWithOAuth) CommandName() string {
	return CommandAuthenticateWithOAuth
}

// Validate implements cqrs.Validatable.
func (c AuthenticateWithOAuth) Validate() error {
	if c.Provider == "" {
		return cqrs.ErrCommandValidation("AuthenticateWithOAuth.Validate", "provider is required")
	}
	if c.Provider != "google" && c.Provider != "github" {
		return cqrs.ErrCommandValidation("AuthenticateWithOAuth.Validate", "provider must be 'google' or 'github'")
	}
	if c.Code == "" {
		return cqrs.ErrCommandValidation("AuthenticateWithOAuth.Validate", "code is required")
	}
	if c.State == "" {
		return cqrs.ErrCommandValidation("AuthenticateWithOAuth.Validate", "state is required")
	}
	return nil
}

// AuthenticateWithOAuthResult is the result data for AuthenticateWithOAuth.
type AuthenticateWithOAuthResult struct {
	UserID      string    `json:"user_id"`
	SessionID   string    `json:"session_id"`
	AccessToken string    `json:"access_token"`
	Email       string    `json:"email"`
	Provider    string    `json:"provider"`
	ExpiresAt   time.Time `json:"expires_at"`
	IsNewUser   bool      `json:"is_new_user"`
}

// ============================================================================
// Session Commands
// ============================================================================

// RefreshSession refreshes a session by rotating the token and extending expiration.
type RefreshSession struct {
	SessionID types.ID `json:"session_id" validate:"required"`
	Token     string   `json:"token" validate:"required"`
}

// CommandName implements cqrs.Command.
func (c RefreshSession) CommandName() string {
	return CommandRefreshSession
}

// Validate implements cqrs.Validatable.
func (c RefreshSession) Validate() error {
	if c.SessionID.IsZero() {
		return cqrs.ErrCommandValidation("RefreshSession.Validate", "session_id is required")
	}
	if c.Token == "" {
		return cqrs.ErrCommandValidation("RefreshSession.Validate", "token is required")
	}
	return nil
}

// RefreshSessionResult is the result data for RefreshSession.
type RefreshSessionResult struct {
	SessionID   string    `json:"session_id"`
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// RevokeSession revokes a specific session.
type RevokeSession struct {
	SessionID types.ID `json:"session_id" validate:"required"`
	Reason    *string  `json:"reason" validate:"omitempty,max=255"`
}

// CommandName implements cqrs.Command.
func (c RevokeSession) CommandName() string {
	return CommandRevokeSession
}

// Validate implements cqrs.Validatable.
func (c RevokeSession) Validate() error {
	if c.SessionID.IsZero() {
		return cqrs.ErrCommandValidation("RevokeSession.Validate", "session_id is required")
	}
	if c.Reason != nil && len(*c.Reason) > 255 {
		return cqrs.ErrCommandValidation("RevokeSession.Validate", "reason must be 255 characters or less")
	}
	return nil
}

// RevokeSessionResult is the result data for RevokeSession.
type RevokeSessionResult struct {
	SessionID string    `json:"session_id"`
	RevokedAt time.Time `json:"revoked_at"`
}

// RevokeAllUserSessions revokes all sessions for a user.
type RevokeAllUserSessions struct {
	UserID types.ID `json:"user_id" validate:"required"`
	Reason *string  `json:"reason" validate:"omitempty,max=255"`
}

// CommandName implements cqrs.Command.
func (c RevokeAllUserSessions) CommandName() string {
	return CommandRevokeAllUserSessions
}

// Validate implements cqrs.Validatable.
func (c RevokeAllUserSessions) Validate() error {
	if c.UserID.IsZero() {
		return cqrs.ErrCommandValidation("RevokeAllUserSessions.Validate", "user_id is required")
	}
	if c.Reason != nil && len(*c.Reason) > 255 {
		return cqrs.ErrCommandValidation("RevokeAllUserSessions.Validate", "reason must be 255 characters or less")
	}
	return nil
}

// RevokeAllUserSessionsResult is the result data for RevokeAllUserSessions.
type RevokeAllUserSessionsResult struct {
	UserID       string    `json:"user_id"`
	RevokedCount int       `json:"revoked_count"`
	RevokedAt    time.Time `json:"revoked_at"`
}

// ============================================================================
// Profile Commands
// ============================================================================

// ChangeDisplayName changes a user's display name.
type ChangeDisplayName struct {
	UserID         types.ID `json:"user_id" validate:"required"`
	NewDisplayName string   `json:"new_display_name" validate:"required,min=1,max=100"`
}

// CommandName implements cqrs.Command.
func (c ChangeDisplayName) CommandName() string {
	return CommandChangeDisplayName
}

// Validate implements cqrs.Validatable.
func (c ChangeDisplayName) Validate() error {
	if c.UserID.IsZero() {
		return cqrs.ErrCommandValidation("ChangeDisplayName.Validate", "user_id is required")
	}
	if c.NewDisplayName == "" {
		return cqrs.ErrCommandValidation("ChangeDisplayName.Validate", "new_display_name is required")
	}
	if len(c.NewDisplayName) > 100 {
		return cqrs.ErrCommandValidation("ChangeDisplayName.Validate", "new_display_name must be 100 characters or less")
	}
	return nil
}

// ChangeDisplayNameResult is the result data for ChangeDisplayName.
type ChangeDisplayNameResult struct {
	UserID      string    `json:"user_id"`
	DisplayName string    `json:"display_name"`
	ChangedAt   time.Time `json:"changed_at"`
}

// ChangeEmail changes a user's email address.
type ChangeEmail struct {
	UserID   types.ID    `json:"user_id" validate:"required"`
	NewEmail types.Email `json:"new_email" validate:"required"`
}

// CommandName implements cqrs.Command.
func (c ChangeEmail) CommandName() string {
	return CommandChangeEmail
}

// Validate implements cqrs.Validatable.
func (c ChangeEmail) Validate() error {
	if c.UserID.IsZero() {
		return cqrs.ErrCommandValidation("ChangeEmail.Validate", "user_id is required")
	}
	if c.NewEmail.IsEmpty() {
		return cqrs.ErrCommandValidation("ChangeEmail.Validate", "new_email is required")
	}
	return nil
}

// ChangeEmailResult is the result data for ChangeEmail.
type ChangeEmailResult struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	ChangedAt time.Time `json:"changed_at"`
}

// ============================================================================
// User Status Commands
// ============================================================================

// ActivateUser activates a pending user account.
type ActivateUser struct {
	UserID types.ID `json:"user_id" validate:"required"`
}

// CommandName implements cqrs.Command.
func (c ActivateUser) CommandName() string {
	return CommandActivateUser
}

// Validate implements cqrs.Validatable.
func (c ActivateUser) Validate() error {
	if c.UserID.IsZero() {
		return cqrs.ErrCommandValidation("ActivateUser.Validate", "user_id is required")
	}
	return nil
}

// ActivateUserResult is the result data for ActivateUser.
type ActivateUserResult struct {
	UserID      string    `json:"user_id"`
	ActivatedAt time.Time `json:"activated_at"`
}

// SuspendUser suspends a user account.
type SuspendUser struct {
	UserID types.ID `json:"user_id" validate:"required"`
	Reason string   `json:"reason" validate:"required,min=1,max=500"`
}

// CommandName implements cqrs.Command.
func (c SuspendUser) CommandName() string {
	return CommandSuspendUser
}

// Validate implements cqrs.Validatable.
func (c SuspendUser) Validate() error {
	if c.UserID.IsZero() {
		return cqrs.ErrCommandValidation("SuspendUser.Validate", "user_id is required")
	}
	if c.Reason == "" {
		return cqrs.ErrCommandValidation("SuspendUser.Validate", "reason is required")
	}
	if len(c.Reason) > 500 {
		return cqrs.ErrCommandValidation("SuspendUser.Validate", "reason must be 500 characters or less")
	}
	return nil
}

// SuspendUserResult is the result data for SuspendUser.
type SuspendUserResult struct {
	UserID      string    `json:"user_id"`
	SuspendedAt time.Time `json:"suspended_at"`
}

// ReactivateUser reactivates a suspended user account.
type ReactivateUser struct {
	UserID types.ID `json:"user_id" validate:"required"`
}

// CommandName implements cqrs.Command.
func (c ReactivateUser) CommandName() string {
	return CommandReactivateUser
}

// Validate implements cqrs.Validatable.
func (c ReactivateUser) Validate() error {
	if c.UserID.IsZero() {
		return cqrs.ErrCommandValidation("ReactivateUser.Validate", "user_id is required")
	}
	return nil
}

// ReactivateUserResult is the result data for ReactivateUser.
type ReactivateUserResult struct {
	UserID        string    `json:"user_id"`
	ReactivatedAt time.Time `json:"reactivated_at"`
}

// DeleteUser soft-deletes a user account.
type DeleteUser struct {
	UserID types.ID `json:"user_id" validate:"required"`
}

// CommandName implements cqrs.Command.
func (c DeleteUser) CommandName() string {
	return CommandDeleteUser
}

// Validate implements cqrs.Validatable.
func (c DeleteUser) Validate() error {
	if c.UserID.IsZero() {
		return cqrs.ErrCommandValidation("DeleteUser.Validate", "user_id is required")
	}
	return nil
}

// DeleteUserResult is the result data for DeleteUser.
type DeleteUserResult struct {
	UserID    string    `json:"user_id"`
	DeletedAt time.Time `json:"deleted_at"`
}

// ============================================================================
// DID Management Commands
// ============================================================================

// AddDID adds a DID to a user's identity.
type AddDID struct {
	UserID types.ID `json:"user_id" validate:"required"`
	DID    string   `json:"did" validate:"required"`
}

// CommandName implements cqrs.Command.
func (c AddDID) CommandName() string {
	return CommandAddDID
}

// Validate implements cqrs.Validatable.
func (c AddDID) Validate() error {
	if c.UserID.IsZero() {
		return cqrs.ErrCommandValidation("AddDID.Validate", "user_id is required")
	}
	if c.DID == "" {
		return cqrs.ErrCommandValidation("AddDID.Validate", "did is required")
	}
	return nil
}

// AddDIDResult is the result data for AddDID.
type AddDIDResult struct {
	UserID  string    `json:"user_id"`
	DID     string    `json:"did"`
	AddedAt time.Time `json:"added_at"`
}

// RemoveDID removes a DID from a user's identity.
type RemoveDID struct {
	UserID types.ID `json:"user_id" validate:"required"`
	DID    string   `json:"did" validate:"required"`
}

// CommandName implements cqrs.Command.
func (c RemoveDID) CommandName() string {
	return CommandRemoveDID
}

// Validate implements cqrs.Validatable.
func (c RemoveDID) Validate() error {
	if c.UserID.IsZero() {
		return cqrs.ErrCommandValidation("RemoveDID.Validate", "user_id is required")
	}
	if c.DID == "" {
		return cqrs.ErrCommandValidation("RemoveDID.Validate", "did is required")
	}
	return nil
}

// RemoveDIDResult is the result data for RemoveDID.
type RemoveDIDResult struct {
	UserID    string    `json:"user_id"`
	DID       string    `json:"did"`
	RemovedAt time.Time `json:"removed_at"`
}

// ============================================================================
// OAuth Linking Commands
// ============================================================================

// LinkOAuthAccount links an OAuth account to an existing user.
type LinkOAuthAccount struct {
	UserID   types.ID `json:"user_id" validate:"required"`
	Provider string   `json:"provider" validate:"required,oneof=google github"`
	Code     string   `json:"code" validate:"required"`
	State    string   `json:"state" validate:"required"`
}

// CommandName implements cqrs.Command.
func (c LinkOAuthAccount) CommandName() string {
	return CommandLinkOAuthAccount
}

// Validate implements cqrs.Validatable.
func (c LinkOAuthAccount) Validate() error {
	if c.UserID.IsZero() {
		return cqrs.ErrCommandValidation("LinkOAuthAccount.Validate", "user_id is required")
	}
	if c.Provider == "" {
		return cqrs.ErrCommandValidation("LinkOAuthAccount.Validate", "provider is required")
	}
	if c.Provider != "google" && c.Provider != "github" {
		return cqrs.ErrCommandValidation("LinkOAuthAccount.Validate", "provider must be 'google' or 'github'")
	}
	if c.Code == "" {
		return cqrs.ErrCommandValidation("LinkOAuthAccount.Validate", "code is required")
	}
	if c.State == "" {
		return cqrs.ErrCommandValidation("LinkOAuthAccount.Validate", "state is required")
	}
	return nil
}

// LinkOAuthAccountResult is the result data for LinkOAuthAccount.
type LinkOAuthAccountResult struct {
	UserID     string    `json:"user_id"`
	Provider   string    `json:"provider"`
	ExternalID string    `json:"external_id"`
	LinkedAt   time.Time `json:"linked_at"`
}

// UnlinkOAuthAccount unlinks an OAuth account from a user.
type UnlinkOAuthAccount struct {
	UserID   types.ID `json:"user_id" validate:"required"`
	Provider string   `json:"provider" validate:"required,oneof=google github"`
}

// CommandName implements cqrs.Command.
func (c UnlinkOAuthAccount) CommandName() string {
	return CommandUnlinkOAuthAccount
}

// Validate implements cqrs.Validatable.
func (c UnlinkOAuthAccount) Validate() error {
	if c.UserID.IsZero() {
		return cqrs.ErrCommandValidation("UnlinkOAuthAccount.Validate", "user_id is required")
	}
	if c.Provider == "" {
		return cqrs.ErrCommandValidation("UnlinkOAuthAccount.Validate", "provider is required")
	}
	if c.Provider != "google" && c.Provider != "github" {
		return cqrs.ErrCommandValidation("UnlinkOAuthAccount.Validate", "provider must be 'google' or 'github'")
	}
	return nil
}

// UnlinkOAuthAccountResult is the result data for UnlinkOAuthAccount.
type UnlinkOAuthAccountResult struct {
	UserID     string    `json:"user_id"`
	Provider   string    `json:"provider"`
	UnlinkedAt time.Time `json:"unlinked_at"`
}
