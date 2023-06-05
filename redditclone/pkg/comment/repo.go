package comment

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"

	"redditclone/pkg/user"
)

var (
	ErrNoComment = errors.New("no comment found")
	ErrBadVote   = errors.New("bad vote number for vote")
	ErrNoDel     = errors.New("there is no comment being deleted")
)

type CommentMemoryRepository struct {
	data  map[string][]*Comment
	mutex sync.Mutex
}

func NewMemoryRepo() *CommentMemoryRepository {
	return &CommentMemoryRepository{
		data:  make(map[string][]*Comment),
		mutex: sync.Mutex{},
	}
}

func (commentRepo *CommentMemoryRepository) Get(commentID string, postID string) (*Comment, error) {
	commentRepo.mutex.Lock()
	defer commentRepo.mutex.Unlock()
	for _, comment := range commentRepo.data[postID] {
		if comment.ID == commentID {
			return comment, nil
		}
	}
	return nil, ErrNoComment
}

func (commentRepo *CommentMemoryRepository) Create(
	text string,
	author *user.User,
	postID string,
) (*Comment, error) {
	commentRepo.mutex.Lock()
	defer commentRepo.mutex.Unlock()
	comment := new(Comment)
	comment.ID = uuid.New().String()
	comment.Author = *author
	comment.Created = time.Now().Format(time.RFC3339)
	comment.Body = text
	if _, ok := commentRepo.data[postID]; ok {
		commentRepo.data[postID] = append(commentRepo.data[postID], comment)
	} else {
		commentRepo.data[postID] = make([]*Comment, 1, 10)
		commentRepo.data[postID][0] = comment
	}
	return comment, nil
}

func (commentRepo *CommentMemoryRepository) Delete(
	comments []*Comment,
	commentID string,
	postID string,
) (int, error) {
	commentRepo.mutex.Lock()
	defer commentRepo.mutex.Unlock()
	for i, comment := range comments {
		if comment.ID == commentID {
			commentRepo.data[postID] = append(commentRepo.data[postID][:i], commentRepo.data[postID][i+1:]...)
			return i, nil
		}
	}
	return -1, ErrNoDel
}

func (commentRepo *CommentMemoryRepository) DeleteAll(postID string) {
	commentRepo.mutex.Lock()
	defer commentRepo.mutex.Unlock()
	delete(commentRepo.data, postID)
}
