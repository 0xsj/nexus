package domain

import (
	"time"
)

// ============================================================================
// Provider Data Models
// ============================================================================

// ProviderData represents normalized data fetched from an external provider.
// Each provider adapter returns data in this standard format.
type ProviderData struct {
	// Provider metadata
	ProviderType     ProviderType
	ProviderUserID   string
	ProviderUsername string
	Email            string

	// Profile data
	Profile ProfileData

	// Provider-specific data
	GitHub   *GitHubData
	LinkedIn *LinkedInData
	Coursera *CourseraData
	Twitter  *TwitterData
	Google   *GoogleData
	AWS      *AWSData

	// Metadata
	FetchedAt time.Time
	RawData   map[string]any // Original API response
}

// ProfileData represents common profile information across providers.
type ProfileData struct {
	Name        string
	Bio         string
	AvatarURL   string
	Location    string
	Website     string
	CompanyName string
	JobTitle    string
}

// ============================================================================
// GitHub Provider Data
// ============================================================================

// GitHubData represents data fetched from GitHub.
type GitHubData struct {
	// Profile
	Login     string
	Name      string
	Bio       string
	AvatarURL string
	Location  string
	Company   string
	Blog      string
	Email     string

	// Stats
	PublicRepos      int
	PublicGists      int
	Followers        int
	Following        int
	TotalStarred     int
	TotalCommits     int

	// Activity
	Repositories     []GitHubRepository
	Contributions    []GitHubContribution
	Organizations    []GitHubOrganization

	// Verification flags
	IsVerified       bool
	IsSiteAdmin      bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// GitHubRepository represents a GitHub repository.
type GitHubRepository struct {
	Name        string
	FullName    string
	Description string
	URL         string
	Homepage    string
	Language    string
	Stars       int
	Forks       int
	IsPrivate   bool
	IsFork      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	PushedAt    time.Time
}

// GitHubContribution represents contribution activity.
type GitHubContribution struct {
	Repository string
	Commits    int
	PRs        int
	Issues     int
	Reviews    int
}

// GitHubOrganization represents a GitHub organization membership.
type GitHubOrganization struct {
	Login       string
	Name        string
	Description string
	AvatarURL   string
	Role        string
}

// ============================================================================
// LinkedIn Provider Data
// ============================================================================

// LinkedInData represents data fetched from LinkedIn.
type LinkedInData struct {
	// Profile
	ID            string
	FirstName     string
	LastName      string
	Headline      string
	ProfileURL    string
	PictureURL    string
	Location      string
	Industry      string
	Summary       string

	// Experience
	Positions     []LinkedInPosition
	Education     []LinkedInEducation
	Certifications []LinkedInCertification
	Skills        []LinkedInSkill

	// Stats
	ConnectionCount int
	FollowerCount   int

	// Metadata
	IsVerified    bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// LinkedInPosition represents a work position.
type LinkedInPosition struct {
	Title       string
	Company     string
	Location    string
	Description string
	StartDate   time.Time
	EndDate     *time.Time // nil if current
	IsCurrent   bool
}

// LinkedInEducation represents educational background.
type LinkedInEducation struct {
	School      string
	Degree      string
	FieldOfStudy string
	Grade       string
	StartDate   time.Time
	EndDate     *time.Time
	Activities  string
	Description string
}

// LinkedInCertification represents a professional certification.
type LinkedInCertification struct {
	Name        string
	Authority   string
	LicenseNumber string
	URL         string
	StartDate   time.Time
	EndDate     *time.Time
}

// LinkedInSkill represents a professional skill with endorsements.
type LinkedInSkill struct {
	Name           string
	Endorsements   int
	Proficiency    string
}

// ============================================================================
// Coursera Provider Data
// ============================================================================

// CourseraData represents data fetched from Coursera.
type CourseraData struct {
	// Profile
	UserID      string
	FullName    string
	Email       string
	ProfileURL  string

	// Courses and Certifications
	Enrollments    []CourseraEnrollment
	Certifications []CourseraCertification
	Specializations []CourseraSpecialization

	// Stats
	TotalCourses      int
	CompletedCourses  int
	TotalCertificates int

	// Metadata
	JoinedAt    time.Time
	UpdatedAt   time.Time
}

// CourseraEnrollment represents a course enrollment.
type CourseraEnrollment struct {
	CourseID     string
	CourseName   string
	CourseURL    string
	InstructorName string
	Institution  string
	EnrolledAt   time.Time
	CompletedAt  *time.Time
	Progress     int // 0-100
	Grade        *float64
	IsCompleted  bool
}

// CourseraCertification represents a verified certificate.
type CourseraCertification struct {
	CertificateID   string
	CertificateURL  string
	CourseID        string
	CourseName      string
	Institution     string
	CompletedAt     time.Time
	Grade           *float64
	VerificationURL string
	IsVerified      bool
}

// CourseraSpecialization represents a specialization program.
type CourseraSpecialization struct {
	SpecializationID string
	Name             string
	Institution      string
	Courses          []string
	CompletedCourses int
	TotalCourses     int
	IsCompleted      bool
	CompletedAt      *time.Time
}

// ============================================================================
// Twitter Provider Data
// ============================================================================

// TwitterData represents data fetched from Twitter/X.
type TwitterData struct {
	// Profile
	ID            string
	Username      string
	DisplayName   string
	Bio           string
	ProfileImage  string
	BannerImage   string
	Location      string
	Website       string

	// Stats
	TweetCount    int
	Followers     int
	Following     int
	Listed        int

	// Verification
	IsVerified    bool
	IsBlueVerified bool
	IsGoldVerified bool
	IsGrayVerified bool

	// Activity
	RecentTweets  []TwitterTweet

	// Metadata
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// TwitterTweet represents a tweet.
type TwitterTweet struct {
	ID            string
	Text          string
	CreatedAt     time.Time
	RetweetCount  int
	LikeCount     int
	ReplyCount    int
	QuoteCount    int
	ViewCount     int
	IsRetweet     bool
	IsReply       bool
}

// ============================================================================
// Google Provider Data
// ============================================================================

// GoogleData represents data fetched from Google services.
type GoogleData struct {
	// Profile (from Google People API)
	ID            string
	Email         string
	Name          string
	GivenName     string
	FamilyName    string
	Picture       string
	Locale        string

	// Verification
	EmailVerified bool

	// Metadata
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// ============================================================================
// AWS Provider Data
// ============================================================================

// AWSData represents data fetched from AWS.
type AWSData struct {
	// IAM User
	UserID        string
	UserName      string
	ARN           string

	// Certifications (via AWS Training API)
	Certifications []AWSCertification

	// Account metadata
	AccountID     string
	Region        string

	// Metadata
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// AWSCertification represents an AWS certification.
type AWSCertification struct {
	CertificationID   string
	Name              string
	Level             string // Foundational, Associate, Professional, Specialty
	EarnedDate        time.Time
	ExpirationDate    *time.Time
	VerificationCode  string
	IsActive          bool
}
