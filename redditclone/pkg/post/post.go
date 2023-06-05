package post

import (
	"redditclone/pkg/comment"
	"redditclone/pkg/user"
)

type Votes struct {
	User string `json:"user"`
	Vote int    `json:"vote"`
}

type Post struct {
	Author           user.User          `json:"author"`
	Category         string             `json:"category"`
	Comments         []*comment.Comment `json:"comments"`
	Created          string             `json:"created"`
	ID               string             `json:"id"`
	Score            int                `json:"score"`
	Text             string             `json:"text,omitempty"`
	URL              string             `json:"url,omitempty"`
	Title            string             `json:"title"`
	Type             string             `json:"type"`
	UpvotePercentage int                `json:"upvotePercentage"`
	Views            int                `json:"views"`
	Votes            []*Votes           `json:"votes"`
}

type PostRepo interface {
	Get(postID string) (Post, error)
	GetPost(postID string) (Post, error)
	GetCategory(category string) ([]Post, error)
	GetAllPosts() ([]Post, error)
	GetUserPosts(userLogin string) ([]Post, error)
	UpdateVote(vote int, postID string, author *user.User) (Post, error)
	Create(post Post) (Post, error)
	AddComment(currpost Post, currComment *comment.Comment) (Post, error)
	Delete(postID string) (bool, error)
	DeleteComment(delCommentID int, postID string) (Post, error)
}
