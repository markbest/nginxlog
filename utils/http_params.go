package utils

import (
	"net/http"
	"strconv"
)

type HttpParams struct {
	request *http.Request
}

func NewHttpPrams(r *http.Request) *HttpParams {
	return &HttpParams{r}
}

//Get status param
func (h *HttpParams) GetStatus() (status string) {
	r := h.request
	r.ParseForm()
	if r.Form.Get("status") != "" {
		status = r.Form.Get("status")
	}
	return status
}

//Get method param
func (h *HttpParams) GetMethod() (method string) {
	r := h.request
	r.ParseForm()
	if r.Form.Get("method") != "" {
		method = r.Form.Get("method")
	}
	return method
}

//Get page param
func (h *HttpParams) GetPage() (page int) {
	r := h.request
	r.ParseForm()
	if r.Form.Get("page") != "" {
		page, _ = strconv.Atoi(r.Form.Get("page"))
	} else {
		page = 1
	}
	return page
}

//Get per_page param
func (h *HttpParams) GetPerPage() (perPage int) {
	r := h.request
	r.ParseForm()
	if r.Form.Get("per_page") != "" {
		perPage, _ = strconv.Atoi(r.Form.Get("per_page"))
	} else {
		perPage = 20
	}
	return perPage
}
