package services

import (
	"fmt"
	"net/http"

	"github.com/strolt/strolt/shared/apiu"
)

type getListResultItem struct {
	Temp string `json:"temp"`
}

type getListResult struct {
	Items []getListResultItem `json:"items"`
}

// getList godoc
// @Id					 getList
// @Summary      Get services list
// @Tags         services
// @Security BasicAuth
// @success 200 {object} getListResult
// @Router       /api/v1/services [get].
func (s *Services) getList(w http.ResponseWriter, r *http.Request) {
	fmt.Println(" get list ")
	apiu.RenderJSON200(w, r, getListResult{Items: []getListResultItem{}})
}
