package rest

import (
	"net/http"

	"fizz-buzz-rest/internal/fizzbuzz"
	resp "fizz-buzz-rest/utils/resp"
)

type statsResponse struct {
	Hits    uint64            `json:"hits"`
	Request *fizzbuzz.Request `json:"request"`
}

func (a *API) Stats(w http.ResponseWriter, r *http.Request) {
	top, hits, ok := a.stats.Top()
	if !ok {
		resp.SendStruct(r.Context(), w, http.StatusOK, statsResponse{})
		return
	}
	resp.SendStruct(r.Context(), w, http.StatusOK, statsResponse{Hits: hits, Request: &top})
}
