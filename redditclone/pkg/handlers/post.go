package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"redditclone/pkg/comment"
	"redditclone/pkg/post"
	"redditclone/pkg/session"
	"redditclone/pkg/user"
)

type PostHandler struct {
	Logger      *zap.SugaredLogger
	PostRepo    post.PostRepo
	CommentRepo comment.CommentRepo
	Sessions    session.SessRepo
}

type PostForm struct {
	Category string `json:"category"`
	Text     string `json:"text"`
	Title    string `json:"title"`
	Type     string `json:"type,omitempty"`
	URL      string `json:"url,omitempty"`
}

type CommentForm struct {
	Comment string `json:"comment"`
}

type PostSort []post.Post

func (a PostSort) Len() int           { return len(a) }
func (a PostSort) Less(i, j int) bool { return a[i].Score > a[j].Score }
func (a PostSort) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// ================================ GET ===============================
func (h *PostHandler) GetPosts(w http.ResponseWriter, r *http.Request) {
	posts, errGetData := h.PostRepo.GetAllPosts()
	if errGetData != nil {
		h.Logger.Infow("Error in getting posts", errGetData)
		http.Error(w, `Error in getting posts`, http.StatusInternalServerError)
		return
	}
	sort.Sort(PostSort(posts))
	resp, errMarsh := json.Marshal(posts)
	if errMarsh != nil {
		h.Logger.Infow("Error in marshaling response", errMarsh)
		h.errorResp(w, http.StatusBadRequest, "cant unpack payload")
		return
	}
	_, errWrite := w.Write(resp)
	if errWrite != nil {
		h.Logger.Infow("Error in writing", errWrite)
		return
	}
}

func (h *PostHandler) GetCategoryPosts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	category, errVars := vars["CATEGORY_NAME"]
	if !errVars {
		h.Logger.Infow("Error in getting category", errVars)
		http.Error(w, `Bad category`, http.StatusBadGateway)
		return
	}
	posts, errGet := h.PostRepo.GetCategory(category)
	if errGet != nil {
		h.Logger.Infow("Error in getting posts", errGet)
		http.Error(w, `Error in getting posts`, http.StatusInternalServerError)
		return
	}
	sort.Sort(PostSort(posts))
	resp, errMarsh := json.Marshal(posts)
	if errMarsh != nil {
		h.Logger.Infow("Error in marshaling response", errMarsh)
		h.errorResp(w, http.StatusBadRequest, "cant unpack payload")
		return
	}
	_, errWrite := w.Write(resp)
	if errWrite != nil {
		h.Logger.Infow("Error in writing", errWrite)
		return
	}
}

func (h *PostHandler) GetPostAndComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID, errVars := vars["POST_ID"]
	if !errVars {
		h.Logger.Infow("Error in getting ID", errVars)
		http.Error(w, `Bad id`, http.StatusBadGateway)
		return
	}
	post, errGet := h.PostRepo.GetPost(postID)
	if errGet != nil {
		h.Logger.Infow("Error in getting posts", errGet)
		http.Error(w, `Error in getting posts`, http.StatusInternalServerError)
		return
	}

	resp, errMarsh := json.Marshal(post)
	if errMarsh != nil {
		h.Logger.Infow("Error in marshaling response", errMarsh)
		h.errorResp(w, http.StatusBadRequest, "cant unpack payload")
		return
	}
	_, errWrite := w.Write(resp)
	if errWrite != nil {
		h.Logger.Infow("Error in writing", errWrite)
		return
	}
}

func (h *PostHandler) Rating(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID, errVars := vars["POST_ID"]
	if !errVars {
		h.Logger.Infow("Error in getting ID", errVars)
		http.Error(w, `Bad id`, http.StatusBadGateway)
		return
	}
	currSession, errSession := session.SessionFromContext(r.Context())
	if errSession != nil {
		h.Logger.Infow("Unauthorized", errSession.Error())
		http.Error(w, "Authorize error", http.StatusUnauthorized)
		h.errorResp(w, http.StatusUnauthorized, "bad token")
		return
	}
	currUser := &user.User{}
	currUser.ID = currSession.UserID
	currUser.Login = currSession.UserLogin

	path := strings.Split(r.URL.Path, "/")
	voteType := path[len(path)-1]
	var vote int
	switch voteType {
	case "upvote":
		vote = 1
	case "downvote":
		vote = -1
	default: // case "downvote"
		vote = 0
	}
	elem, errVote := h.PostRepo.UpdateVote(vote, postID, currUser)
	if errVote != nil {
		h.Logger.Infow("Error in UpdateVote", errVote)
		http.Error(w, `Error in updating vote`, http.StatusInternalServerError)
		return
	}

	resp, errMarshal := json.Marshal(elem)
	if errMarshal != nil {
		h.Logger.Infow("Error in Marshaling response", errMarshal)
		h.errorResp(w, http.StatusBadRequest, "cant unpack payload")
		return
	}
	_, errWrite := w.Write(resp)
	if errWrite != nil {
		h.Logger.Infow("Error in writing", errWrite)
		return
	}
}

func (h *PostHandler) UserPosts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userLogin, errVars := vars["USER_LOGIN"]
	if !errVars {
		h.Logger.Infow("Error in getting login", errVars)
		http.Error(w, `Bad login`, http.StatusBadGateway)
		return
	}

	posts, errGet := h.PostRepo.GetUserPosts(userLogin)
	if errGet != nil {
		h.Logger.Infow("Error in getting posts", errGet)
		http.Error(w, `Error in getting posts`, http.StatusInternalServerError)
		return
	}
	sort.Sort(PostSort(posts))
	resp, errMarsh := json.Marshal(posts)
	if errMarsh != nil {
		h.Logger.Infow("Error in marshaling response", errMarsh)
		h.errorResp(w, http.StatusBadRequest, "cant unpack payload")
	}
	_, errWrite := w.Write(resp)
	if errWrite != nil {
		h.Logger.Infow("Error in writing", errWrite)
		return
	}
}

// =============================== POST ===============================
func (h *PostHandler) AddPost(w http.ResponseWriter, r *http.Request) {

	body, errBodyRead := io.ReadAll(r.Body)
	if errBodyRead != nil {
		h.Logger.Infow("Error in reading req body", errBodyRead)
		http.Error(w, `Error in reading req body`, http.StatusInternalServerError)
		return
	}
	defer func(r *http.Request) {
		errBodyClose := r.Body.Close()
		if errBodyClose != nil {
			h.Logger.Infow("Error in closing req body", errBodyClose)
			return
		}
	}(r)

	currSession, errSession := session.SessionFromContext(r.Context())
	if errSession != nil {
		h.Logger.Infow("Unauthorized", errSession)
		http.Error(w, "Authorize error", http.StatusUnauthorized)
		h.errorResp(w, http.StatusUnauthorized, "bad token")
		return
	}

	post := post.Post{}
	errUnmarsh := json.Unmarshal(body, &post)
	if errUnmarsh != nil {
		h.Logger.Infow("Error in unmarshaling", errUnmarsh)
		http.Error(w, `Error in unmarshaling`, http.StatusInternalServerError)
		return
	}
	currUser := &user.User{}
	currUser.ID = currSession.UserID
	currUser.Login = currSession.UserLogin

	post.Author = *currUser
	post, errCreate := h.PostRepo.Create(post)
	if errCreate != nil {
		h.Logger.Infow("Error in creating post", errCreate)
		http.Error(w, `Error in creating post`, http.StatusInternalServerError)
		return
	}
	post, errUpd := h.PostRepo.UpdateVote(1, post.ID, currUser)
	if errUpd != nil {
		h.Logger.Infow("Error in UpdateVote", errUpd)
		http.Error(w, `Error in updating vote`, http.StatusInternalServerError)
		return
	}

	resp, errMarsh := json.Marshal(post)
	if errMarsh != nil {
		h.Logger.Infow("Error in marshaling", errMarsh)
		http.Error(w, `Error in marshaling`, http.StatusInternalServerError)
		return
	}
	_, errWrite := w.Write(resp)
	if errWrite != nil {
		h.Logger.Infow("Error in writing responce", errWrite)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *PostHandler) AddComment(w http.ResponseWriter, r *http.Request) {

	body, errBodyRead := io.ReadAll(r.Body)
	if errBodyRead != nil {
		h.Logger.Infow("Error in reading req body", errBodyRead)
		http.Error(w, `Error in reading req body`, http.StatusInternalServerError)
		return
	}
	defer func(r *http.Request) {
		errBodyClose := r.Body.Close()
		if errBodyClose != nil {
			h.Logger.Infow("Error in closing req body", errBodyClose)
			return
		}
	}(r)

	currSession, errSession := session.SessionFromContext(r.Context())
	if errSession != nil {
		h.Logger.Infow("Unauthorized", errSession)
		http.Error(w, "Authorize error", http.StatusUnauthorized)
		h.errorResp(w, http.StatusUnauthorized, "bad token")
		return
	}
	currUser := &user.User{}
	currUser.ID = currSession.UserID
	currUser.Login = currSession.UserLogin

	vars := mux.Vars(r)
	id, errID := vars["POST_ID"]
	if !errID {
		h.Logger.Infow("Error in getting id", errID)
		http.Error(w, `Bad id`, http.StatusBadGateway)
		return
	}

	commentForm := &CommentForm{}
	errUnmarsh := json.Unmarshal(body, commentForm)
	if errUnmarsh != nil {
		h.Logger.Infow("Error in unmarshaling", errUnmarsh)
		http.Error(w, `Error in unmarshaling`, http.StatusInternalServerError)
		return
	}
	if commentForm.Comment == "" {
		RespErr, errMarh := json.Marshal(map[string][]ErrForm{
			"errors": {
				{
					Location: "body",
					Param:    "comment",
					Msg:      "is required",
				},
			}})
		if errMarh != nil {
			h.Logger.Infow("Error in Marshaling response", errMarh)
		}
		http.Error(w, "Unable to process the instructions", http.StatusUnprocessableEntity)
		_, errWrite := w.Write(RespErr)
		if errWrite != nil {
			h.Logger.Infow("Error in writing", errWrite)
		}
		return
	}
	post, errGetPost := h.PostRepo.Get(id)
	if errGetPost != nil {
		h.Logger.Infow("Error in getting post", errGetPost)
		http.Error(w, `Error in getting post`, http.StatusInternalServerError)
		return
	}
	currComment, errComment := h.CommentRepo.Create(commentForm.Comment, currUser, post.ID)
	if errComment != nil {
		h.Logger.Infow("Error in creating comment", errComment)
		http.Error(w, `Error in creating comment`, http.StatusInternalServerError)
		return
	}
	post, errAddComment := h.PostRepo.AddComment(post, currComment)
	if errAddComment != nil {
		h.Logger.Infow("Error in adding comment", errAddComment)
		http.Error(w, `Error in adding comment`, http.StatusInternalServerError)
		return
	}
	resp, errMarsh := json.Marshal(post)
	if errMarsh != nil {
		h.Logger.Infow("Error in marshaling response", errMarsh)
		h.errorResp(w, http.StatusBadRequest, "cant unpack payload")
		return
	}
	_, errWrite := w.Write(resp)
	if errWrite != nil {
		h.Logger.Infow("Error in writing", errWrite)
		return
	}
}

// ============================== DELETE ==============================
func (h *PostHandler) DelPost(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	postID, errID := vars["POST_ID"]
	if !errID {
		h.Logger.Infow("Error in getting id", errID)
		http.Error(w, `Bad id`, http.StatusBadGateway)
		return
	}

	currSession, errSession := session.SessionFromContext(r.Context())
	if errSession != nil {
		h.Logger.Infow("Unauthorized", errSession)
		http.Error(w, "Authorize error", http.StatusUnauthorized)
		h.errorResp(w, http.StatusUnauthorized, "bad token")
		return
	}
	currUser := &user.User{}
	currUser.ID = currSession.UserID
	currUser.Login = currSession.UserLogin

	post, errGet := h.PostRepo.GetPost(postID)
	if errGet != nil {
		h.Logger.Infow("Error in getting posts", errGet)
		http.Error(w, `Error in getting posts`, http.StatusInternalServerError)
		return
	}
	if currUser.ID != post.Author.ID {
		h.Logger.Infow("Unauthorized", errSession)
		http.Error(w, "Authorize error", http.StatusUnauthorized)
		h.errorResp(w, http.StatusUnauthorized, "bad token")
		return
	}

	ok, errDel := h.PostRepo.Delete(postID)
	if errDel != nil {
		h.Logger.Infow("Error in deleting post", errID)
		http.Error(w, `Error in deleting post`, http.StatusInternalServerError)
		return
	}
	if ok {
		// also del comments repo
		h.CommentRepo.DeleteAll(postID)
		resp, errMarsh := json.Marshal(map[string]interface{}{
			"message": "success",
		})
		if errMarsh != nil {
			h.Logger.Infow("Error of Marshal", errMarsh)
		}
		_, errWrite := w.Write(resp)
		if errWrite != nil {
			h.Logger.Infow("Error of write", errWrite)
		}
	} else {
		h.errorResp(w, http.StatusInternalServerError, "error of delete")
	}
}

func (h *PostHandler) DelComment(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	postID, errID := vars["POST_ID"]
	if !errID {
		h.Logger.Infow("Error in getting id", errID)
		http.Error(w, `Bad id`, http.StatusBadGateway)
		return
	}

	currSession, errSession := session.SessionFromContext(r.Context())
	if errSession != nil {
		h.Logger.Infow("Unauthorized", errSession)
		http.Error(w, "Authorize error", http.StatusUnauthorized)
		h.errorResp(w, http.StatusUnauthorized, "bad token")
		return
	}
	currUser := &user.User{}
	currUser.ID = currSession.UserID
	currUser.Login = currSession.UserLogin

	commentID, errCommentID := vars["COMMENT_ID"]
	if !errCommentID {
		h.Logger.Infow("Error in getting comment id", errCommentID)
		http.Error(w, `Bad id`, http.StatusBadGateway)
		return
	}
	post, errGetPost := h.PostRepo.Get(postID)
	if errGetPost != nil {
		h.Logger.Infow("Error in getting post", errGetPost)
		http.Error(w, `Error in getting post`, http.StatusInternalServerError)
		return
	}

	comment, errGet := h.CommentRepo.Get(commentID, post.ID)
	if errGet != nil {
		h.Logger.Infow("Error in getting comment", errGet)
		http.Error(w, `Error in getting comment`, http.StatusInternalServerError)
		return
	}
	if currUser.ID != comment.Author.ID {
		h.Logger.Infow("Unauthorized", errSession)
		http.Error(w, "Authorize error", http.StatusUnauthorized)
		h.errorResp(w, http.StatusUnauthorized, "bad token")
		return
	}

	delIDx, errDel := h.CommentRepo.Delete(post.Comments, commentID, post.ID)
	if errDel != nil {
		h.Logger.Infow("Error in deleting comment", errID)
		http.Error(w, `Error in deleting comment`, http.StatusInternalServerError)
		return
	}
	post, errDelComment := h.PostRepo.DeleteComment(delIDx, postID)
	if errDelComment != nil {
		h.Logger.Infow("Error in deleting comment in post", errDelComment)
		http.Error(w, `Error in deleting comment in post`, http.StatusInternalServerError)
		return
	}
	resp, errMarsh := json.Marshal(post)
	if errMarsh != nil {
		h.Logger.Infow("Error in marshaling response", errMarsh)
		h.errorResp(w, http.StatusBadRequest, "cant unpack payload")
		return
	}
	_, errWrite := w.Write(resp)
	if errWrite != nil {
		h.Logger.Infow("Error in writing", errWrite)
		return
	}
}

// ============================== HELP FUNC ==============================
func (h *PostHandler) errorResp(w http.ResponseWriter, status int, msg string) {
	resp, errMarsh := json.Marshal(map[string]interface{}{
		"status": status,
		"error":  msg,
	})
	if errMarsh != nil {
		h.Logger.Infow("Error in marshal resp", errMarsh)
	}
	_, errWrite := w.Write(resp)
	if errWrite != nil {
		h.Logger.Infow("Error in write resp", errMarsh)
	}
}
