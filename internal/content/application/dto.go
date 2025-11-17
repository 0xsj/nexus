package application

import (
	"time"

	"github.com/0xsj/nexus/internal/content/domain"
)

// PostDTO is the application-layer representation of a post.
type PostDTO struct {
	ID          string
	UserID      string
	Type        string
	Content     string
	Visibility  string
	Status      string
	Tags        []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	PublishedAt *time.Time
	ScheduledAt *time.Time
}

// PostFromDomain converts a domain Post to a PostDTO.
func PostFromDomain(post *domain.Post) PostDTO {
	return PostDTO{
		ID:          post.ID,
		UserID:      post.UserID,
		Type:        string(post.Type),
		Content:     post.Content,
		Visibility:  string(post.Visibility),
		Status:      string(post.Status),
		Tags:        post.Tags,
		CreatedAt:   post.CreatedAt,
		UpdatedAt:   post.UpdatedAt,
		PublishedAt: post.PublishedAt,
		ScheduledAt: post.ScheduledAt,
	}
}

// PostListDTO contains a list of posts with pagination info.
type PostListDTO struct {
	Posts   []PostDTO
	Total   int
	Limit   int
	Offset  int
	HasMore bool
}

// PostListFromDomain converts domain PostListResult to PostListDTO.
func PostListFromDomain(result *domain.PostListResult) PostListDTO {
	posts := make([]PostDTO, len(result.Posts))
	for i, post := range result.Posts {
		posts[i] = PostFromDomain(post)
	}

	return PostListDTO{
		Posts:   posts,
		Total:   result.Total,
		Limit:   result.Limit,
		Offset:  result.Offset,
		HasMore: result.HasMore,
	}
}
