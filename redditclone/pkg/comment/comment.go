package comment

import "redditclone/pkg/user"

type Comment struct {
	Author  user.User `json:"author"`
	Body    string    `json:"body"`
	Created string    `json:"created"`
	ID      string    `json:"id"`
}

type CommentRepo interface {
	Get(commentID string, postID string) (*Comment, error)
	Create(text string, author *user.User, postID string) (*Comment, error)
	Delete(comments []*Comment, commentID string, postID string) (int, error)
	DeleteAll(postID string)
}
