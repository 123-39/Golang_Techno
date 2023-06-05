package post

import (
	"errors"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"

	"redditclone/pkg/comment"
	"redditclone/pkg/user"
)

var (
	ErrNoPost    = errors.New("no post found")
	ErrNoDel     = errors.New("there is no post being deleted")
	ErrNoDelComm = errors.New("there is no comment being deleted")
)

type PostMemoryRepository struct {
	data  []Post
	mutex sync.Mutex
}

func NewMemoryRepo() *PostMemoryRepository {
	return &PostMemoryRepository{
		data:  make([]Post, 0, 10),
		mutex: sync.Mutex{},
	}
}

// ================================ GET ===============================
func (repo *PostMemoryRepository) Get(postID string) (Post, error) {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()
	for _, post := range repo.data {
		if post.ID == postID {
			return post, nil
		}
	}
	return Post{}, ErrNoPost
}

func (repo *PostMemoryRepository) GetPost(postID string) (Post, error) {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()
	for idx, post := range repo.data {
		if post.ID == postID {
			post.Views += 1
			repo.data[idx] = post
			return post, nil
		}
	}
	return Post{}, ErrNoPost
}

func (repo *PostMemoryRepository) GetCategory(category string) ([]Post, error) {
	suitablePosts := make([]Post, 0, 10)
	repo.mutex.Lock()
	defer repo.mutex.Unlock()
	for _, post := range repo.data {
		if post.Category == category {
			suitablePosts = append(suitablePosts, post)
		}
	}
	return suitablePosts, nil
}

func (repo *PostMemoryRepository) GetAllPosts() ([]Post, error) {
	return repo.data, nil
}

func (repo *PostMemoryRepository) GetUserPosts(userLogin string) ([]Post, error) {
	suitablePosts := make([]Post, 0, 10)
	repo.mutex.Lock()
	defer repo.mutex.Unlock()
	for _, post := range repo.data {
		if post.Author.Login == userLogin {
			suitablePosts = append(suitablePosts, post)
		}
	}
	return suitablePosts, nil
}

// =============================== POST ===============================
func (repo *PostMemoryRepository) Create(post Post) (Post, error) {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()
	post.Created = time.Now().Format(time.RFC3339)
	post.UpvotePercentage = 100
	post.Views = 0
	post.Score = 0
	post.Comments = make([]*comment.Comment, 0, 10)
	post.Votes = make([]*Votes, 0, 10)
	post.ID = uuid.New().String()
	repo.data = append(repo.data, post)
	return post, nil
}

func (repo *PostMemoryRepository) UpdateVote(
	vote int,
	postID string,
	author *user.User,
) (Post, error) {
	post, errGet := repo.Get(postID)
	if errGet != nil {
		return Post{}, errGet
	}
	repo.mutex.Lock()
	defer repo.mutex.Unlock()
	newVote := &Votes{
		User: author.ID,
		Vote: vote,
	}
	delIDx := -1
	isNewVote := true
	for idx, item := range post.Votes {
		if item.User == newVote.User {
			if vote == 0 {
				delIDx = idx
			} else {
				post.Votes[idx] = newVote
				isNewVote = false
			}
			break
		}
	}

	if delIDx != -1 {
		post.Votes = append(post.Votes[:delIDx], post.Votes[delIDx+1:]...)
	} else if isNewVote {
		post.Votes = append(post.Votes, newVote)
	}

	score := 0
	upvotes := 0
	numbVotes := 0
	for _, item := range post.Votes {
		score += item.Vote
		if item.Vote == 1 {
			upvotes++
		}
		numbVotes += 1
	}

	post.Score = score
	if numbVotes == 0 {
		post.UpvotePercentage = 0
	} else {
		post.UpvotePercentage = int(math.Abs(float64(upvotes) / float64(numbVotes) * 100))
	}
	// UPD repo
	for idx, item := range repo.data {
		if item.ID == postID {
			repo.data[idx] = post
		}
	}
	return post, nil
}

func (repo *PostMemoryRepository) AddComment(currpost Post, currComment *comment.Comment) (Post, error) {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()
	for idx, post := range repo.data {
		if post.ID == currpost.ID {
			currpost.Comments = append(currpost.Comments, currComment)
			repo.data[idx] = currpost
			return currpost, nil
		}
	}
	return Post{}, ErrNoDelComm
}

// ============================== DELETE ==============================
func (repo *PostMemoryRepository) Delete(postID string) (bool, error) {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()
	for i, post := range repo.data {
		if post.ID == postID {
			repo.data = append(repo.data[:i], repo.data[i+1:]...)
			return true, nil
		}
	}
	return false, ErrNoDel
}

func (repo *PostMemoryRepository) DeleteComment(delCommentID int, postID string) (Post, error) {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()
	for idx, post := range repo.data {
		if post.ID == postID {
			post.Comments = append(post.Comments[:delCommentID], post.Comments[delCommentID+1:]...)
			repo.data[idx] = post
			return post, nil
		}
	}
	return Post{}, ErrNoDelComm
}
