package main

import (
	"io/ioutil"
	"net/http"

	"redditclone/pkg/comment"
	"redditclone/pkg/handlers"
	"redditclone/pkg/middleware"
	"redditclone/pkg/post"
	"redditclone/pkg/session"
	"redditclone/pkg/user"

	"fmt"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func main() {

	sm := session.NewSessionsManager()

	userRepo := user.NewMemoryRepo()
	postRepo := post.NewMemoryRepo()
	commentRepo := comment.NewMemoryRepo()

	zapLogger, err := zap.NewProduction()
	if err != nil {
		fmt.Println("zapLogger error:", err)
		return
	}
	defer func() {
		err := zapLogger.Sync()
		if err != nil {
			fmt.Println("zapLogger.Sync() error:", err)
		}
	}()
	logger := zapLogger.Sugar()

	userHandler := &handlers.UserHandler{
		UserRepo: userRepo,
		Logger:   logger,
		Sessions: sm,
	}
	postHandler := &handlers.PostHandler{
		PostRepo:    postRepo,
		CommentRepo: commentRepo,
		Logger:      logger,
		Sessions:    sm,
	}

	r := mux.NewRouter()
	// =============================== POST ===============================
	r.HandleFunc("/api/login", userHandler.Login).Methods("POST")
	r.HandleFunc("/api/register", userHandler.Register).Methods("POST")
	r.HandleFunc("/api/posts", postHandler.AddPost).Methods("POST")
	r.HandleFunc("/api/post/{POST_ID}", postHandler.AddComment).Methods("POST")

	// ================================ GET ===============================
	r.HandleFunc("/api/posts/", postHandler.GetPosts).Methods("GET")
	r.HandleFunc("/api/posts/{CATEGORY_NAME}", postHandler.GetCategoryPosts).Methods("GET")
	r.HandleFunc("/api/post/{POST_ID}", postHandler.GetPostAndComment).Methods("GET")
	r.HandleFunc("/api/post/{POST_ID}/upvote", postHandler.Rating).Methods("GET")
	r.HandleFunc("/api/post/{POST_ID}/downvote", postHandler.Rating).Methods("GET")
	r.HandleFunc("/api/post/{POST_ID}/unvote", postHandler.Rating).Methods("GET")
	r.HandleFunc("/api/user/{USER_LOGIN}", postHandler.UserPosts).Methods("GET")

	// ============================== DELETE ==============================
	r.HandleFunc("/api/post/{POST_ID}", postHandler.DelPost).Methods("DELETE")
	r.HandleFunc("/api/post/{POST_ID}/{COMMENT_ID}", postHandler.DelComment).Methods("DELETE")

	// ============================== STATIC ==============================
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	r.Handle("/", http.FileServer(http.Dir("./static/html/")))

	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file, errReadFile := ioutil.ReadFile("./static/html/index.html")
		if errReadFile != nil {
			logger.Infow("Error in Read", errReadFile)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, err := w.Write(file)
		if err != nil {
			logger.Infow("Error in Write", err)
			return
		}
	})

	mux := middleware.Auth(sm, r)
	mux = middleware.AccessLog(logger, mux)
	mux = middleware.Panic(mux)

	addr := ":8020"
	logger.Infow("starting server",
		"type", "START",
		"addr", addr,
	)
	errListen := http.ListenAndServe(addr, mux)
	if errListen != nil {
		fmt.Println("ListenAndServe error:", errListen)
		return
	}

}
