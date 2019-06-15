package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/camptocamp/terradb/internal/storage"
	"github.com/gorilla/mux"
)

func (s *server) ListWorkspaces(w http.ResponseWriter, r *http.Request) {
	page, pageSize, err := s.parsePagination(r)
	if err != nil {
		err500(err, "", w)
		return
	}

	coll, err := s.st.ListWorkspaces(page, pageSize)
	if err != nil {
		err500(err, "failed to retrieve workspaces", w)
		return
	}

	data, err := json.Marshal(coll)
	if err != nil {
		err500(err, "failed to marshal workspaces", w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return
}

func (s *server) LockWorkspace(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	var currentLock, remoteLock storage.LockInfo

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err500(err, "failed to read body", w)
		return
	}

	err = json.Unmarshal(body, &currentLock)
	if err != nil {
		err500(err, "failed to unmarshal lock", w)
		return
	}

	remoteLock, err = s.st.GetLockStatus(params["name"])
	if err == storage.ErrNoDocuments {
		err = s.st.LockWorkspace(params["name"], currentLock)
		if err != nil {
			err500(err, "failed to lock state", w)
			return
		}

		w.WriteHeader(http.StatusOK)
		return
	} else if err != nil {
		err500(err, "failed to get lock status", w)
		return
	}

	if currentLock.ID == remoteLock.ID {
		d, _ := json.Marshal(remoteLock)
		w.WriteHeader(http.StatusLocked)
		w.Write(d)
		return
	}

	d, _ := json.Marshal(remoteLock)
	w.WriteHeader(http.StatusConflict)
	w.Write(d)
	return

}

func (s *server) UnlockWorkspace(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	var lockData storage.LockInfo

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err500(err, "failed to read body", w)
		return
	}

	err = json.Unmarshal(body, &lockData)
	if err != nil {
		err500(err, "failed to unmarshal lock", w)
		return
	}

	err = s.st.UnlockWorkspace(params["name"], lockData)
	if err != nil {
		err500(err, "failed to unlock state", w)
		return
	}

	w.WriteHeader(http.StatusOK)
	return
}
