package domain

import (
	"time"

	"github.com/0xsj/result"
)

const (
	MinPollOptions = 2
	MaxPollOptions = 10
)

// Poll represents a poll attached to a post.
type Poll struct {
	Question             string
	Options              []PollOption
	ExpiresAt            time.Time
	AllowMultipleChoices bool
	TotalVotes           int
}

// PollOption represents a single option in a poll.
type PollOption struct {
	Index      int
	Text       string
	Votes      int
	Percentage float64
}

// NewPoll creates a new Poll.
func NewPoll(question string, options []string, expiresAt time.Time, allowMultiple bool) result.Result[*Poll] {
	// Validate options count
	if len(options) < MinPollOptions {
		return result.Err[*Poll](ErrInvalidPollOptions())
	}

	if len(options) > MaxPollOptions {
		return result.Err[*Poll](ErrTooManyPollOptions(len(options), MaxPollOptions))
	}

	// Validate expiration time
	if expiresAt.Before(time.Now().UTC()) {
		return result.Err[*Poll](ErrScheduledTimeInPast())
	}

	// Create poll options
	pollOptions := make([]PollOption, len(options))
	for i, text := range options {
		pollOptions[i] = PollOption{
			Index:      i,
			Text:       text,
			Votes:      0,
			Percentage: 0,
		}
	}

	poll := &Poll{
		Question:             question,
		Options:              pollOptions,
		ExpiresAt:            expiresAt,
		AllowMultipleChoices: allowMultiple,
		TotalVotes:           0,
	}

	return result.Ok(poll)
}

// IsExpired checks if the poll has expired.
func (p *Poll) IsExpired() bool {
	return time.Now().UTC().After(p.ExpiresAt)
}

// CanVote checks if voting is still allowed.
func (p *Poll) CanVote() bool {
	return !p.IsExpired()
}

// Vote records a vote for an option.
func (p *Poll) Vote(optionIndex int) error {
	if p.IsExpired() {
		return ErrPollExpired()
	}

	if optionIndex < 0 || optionIndex >= len(p.Options) {
		return ErrInvalidPollOptions()
	}

	p.Options[optionIndex].Votes++
	p.TotalVotes++
	p.recalculatePercentages()

	return nil
}

// recalculatePercentages updates vote percentages for all options.
func (p *Poll) recalculatePercentages() {
	if p.TotalVotes == 0 {
		return
	}

	for i := range p.Options {
		p.Options[i].Percentage = float64(p.Options[i].Votes) / float64(p.TotalVotes) * 100
	}
}
