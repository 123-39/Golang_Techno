package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"redditclone/pkg/session"
	"redditclone/pkg/user"

	"go.uber.org/zap"
)

type UserHandler struct {
	Logger   *zap.SugaredLogger
	UserRepo user.UserRepo
	Sessions session.SessRepo
}

type LoginForm struct {
	Login    string `json:"username"`
	Password string `json:"password"`
}

type ErrForm struct {
	Location string `json:"location"`
	Param    string `json:"param"`
	Msg      string `json:"msg"`
	Value    string `json:"value,omitempty"`
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	body, errRead := io.ReadAll(r.Body)
	if errRead != nil {
		h.Logger.Infow("Error in reading req body", errRead)
		http.Error(w, `Error in reading req body`, http.StatusInternalServerError)
		return
	}
	defer func(r *http.Request) {
		errBody := r.Body.Close()
		if errBody != nil {
			h.Logger.Infow("Error in closing req body", errBody)
			return
		}
	}(r)
	logForm := &LoginForm{}
	errUnMarsh := json.Unmarshal(body, logForm)
	if errUnMarsh != nil {
		h.Logger.Infow("Error in unmarshaling LoginForm", errUnMarsh)
		h.errorResp(w, http.StatusBadRequest, "cant unpack payload")
		return
	}

	user, errAuth := h.UserRepo.Authorize(logForm.Login, logForm.Password)
	if errAuth != nil {
		h.Logger.Infow(errAuth.Error())
		h.errorResp(w, http.StatusUnauthorized, "bad login or password")
		http.Error(w, "Authorize error", http.StatusUnauthorized)
		return
	}
	sess, errSession := h.Sessions.Create(user)
	if errSession != nil {
		h.Logger.Infow("Err in session creating: ", errSession)
		http.Error(w, "Authorize error", http.StatusUnauthorized)
		return
	}
	tokenString, err := h.Sessions.CreateToken(sess)
	if err != nil {
		h.Logger.Infow("Err jwt", err.Error())
		h.errorResp(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp, errMrsh := json.Marshal(map[string]interface{}{
		"token": tokenString,
	})
	if errMrsh != nil {
		h.Logger.Infow("Error in marshal resp", errMrsh)
		return
	}
	_, errWrite := w.Write(resp)
	if errWrite != nil {
		h.Logger.Infow("Error in write resp", errWrite)
		return
	}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	body, errRead := io.ReadAll(r.Body)
	if errRead != nil {
		h.Logger.Infow("Error in reading req body", errRead)
		http.Error(w, `Error in reading req body`, http.StatusInternalServerError)
		return
	}
	defer func(r *http.Request) {
		errBody := r.Body.Close()
		if errBody != nil {
			h.Logger.Infow("Error in closing req body", errBody)
			return
		}
	}(r)

	logForm := &LoginForm{}
	errUnMarsh := json.Unmarshal(body, logForm)
	if errUnMarsh != nil {
		h.Logger.Infow("Error in unmarshaling LoginForm", errUnMarsh)
		h.errorResp(w, http.StatusBadRequest, "cant unpack payload")
		return
	}
	_, errUser := h.UserRepo.Authorize(logForm.Login, logForm.Password)
	if errUser != user.ErrNoUser {
		RespErr, errMarsh := json.Marshal(map[string][]ErrForm{
			"errors": {
				{
					Location: "body",
					Msg:      "already exists",
					Param:    "username",
					Value:    logForm.Login,
				},
			}})
		if errMarsh != nil {
			h.Logger.Infow("Error in marshal resp", errMarsh)
		}
		h.Logger.Infow("Unable to process the instructions", errUser)
		http.Error(w, "", http.StatusUnprocessableEntity)
		_, errWrite := w.Write(RespErr)
		if errWrite != nil {
			h.Logger.Infow("Error in write resp", errWrite)
		}
		return
	}
	user, errAuth := h.UserRepo.AddUser(logForm.Login, logForm.Password)
	if errAuth != nil {
		h.Logger.Infow(errAuth.Error())
		http.Error(w, "Authorize error", http.StatusUnauthorized)
		h.errorResp(w, http.StatusUnauthorized, "bad login or password")
		return
	}
	sess, errSession := h.Sessions.Create(user)
	if errSession != nil {
		h.Logger.Infow("Err in session creating", errSession)
		http.Error(w, "Authorize error", http.StatusUnauthorized)
		return
	}

	tokenString, err := h.Sessions.CreateToken(sess)
	if err != nil {
		h.Logger.Infow("Err jwt", err.Error())
		h.errorResp(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp, errMrsh := json.Marshal(map[string]interface{}{
		"token": tokenString,
	})
	if errMrsh != nil {
		h.Logger.Infow("Error in marshal resp", errMrsh)
		return
	}
	_, errWrite := w.Write(resp)
	if errWrite != nil {
		h.Logger.Infow("Error in write resp", errWrite)
		return
	}
}

func (h *UserHandler) errorResp(w http.ResponseWriter, status int, msg string) {
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
