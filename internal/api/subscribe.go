package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	chimw "github.com/go-chi/chi/v5/middleware"
)

func (h *Handler) subscribe(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	body.Email = strings.TrimSpace(strings.ToLower(body.Email))
	if !isValidEmail(body.Email) {
		writeError(w, http.StatusBadRequest, "invalid email address")
		return
	}

	token, needsConfirm, err := h.store.CreateSubscriber(r.Context(), body.Email)
	if err != nil {
		log.Printf("[%s] store.CreateSubscriber: %v", chimw.GetReqID(r.Context()), err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if needsConfirm {
		if err := h.mailer.SendConfirmation(r.Context(), body.Email, token); err != nil {
			log.Printf("[%s] mailer.SendConfirmation: %v", chimw.GetReqID(r.Context()), err)
		}
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) confirmSubscription(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Redirect(w, r, h.siteURL+"/?subscribed=error", http.StatusSeeOther)
		return
	}

	if err := h.store.ConfirmSubscriber(r.Context(), token); err != nil {
		log.Printf("[%s] store.ConfirmSubscriber: %v", chimw.GetReqID(r.Context()), err)
		http.Redirect(w, r, h.siteURL+"/?subscribed=error", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, h.siteURL+"/?subscribed=1", http.StatusSeeOther)
}

func (h *Handler) unsubscribe(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Redirect(w, r, h.siteURL+"/?unsubscribed=error", http.StatusSeeOther)
		return
	}

	if err := h.store.UnsubscribeByToken(r.Context(), token); err != nil {
		log.Printf("[%s] store.UnsubscribeByToken: %v", chimw.GetReqID(r.Context()), err)
		http.Redirect(w, r, h.siteURL+"/?unsubscribed=error", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, h.siteURL+"/?unsubscribed=1", http.StatusSeeOther)
}

func isValidEmail(email string) bool {
	at := strings.LastIndex(email, "@")
	if at < 1 {
		return false
	}
	domain := email[at+1:]
	dot := strings.LastIndex(domain, ".")
	return dot > 0 && dot < len(domain)-1
}
