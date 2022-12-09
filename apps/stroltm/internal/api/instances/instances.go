package instances

import (
	"net/http"

	"github.com/strolt/strolt/apps/stroltm/internal/sdk/strolt/generated/models"
	"github.com/strolt/strolt/shared/apiu"

	"github.com/go-chi/chi/v5"
)

func Router(r chi.Router) {
	r.Route("/api/v1/instances", func(r chi.Router) {
		r.Get("/", getList)
	})
}

type getListResult struct {
	Data []getListResultItem `json:"data"`
}

type getListResultItem struct {
	URL        string           `json:"url"`
	Operations models.APIConfig `json:"operations"`
}

// getStatus godoc
// @Summary      Get task statuses
// @Tags         instances
// @Accept       json
// @Produce      json
// @success 200 {object} getListResult
// @Router       /api/instances [get].
func getList(w http.ResponseWriter, r *http.Request) {
	data := []getListResultItem{}

	// for _, instance := range config.Get().Instances {
	// 	instanceData, err := strolt.New(instance.URL).GetConfig()
	// 	if err == nil {
	// 		data = append(data, getListResultItem{
	// 			URL:        instance.URL,
	// 			Operations: *instanceData.Payload,
	// 		})
	// 	}
	// }

	apiu.RenderJSON200(w, r, getListResult{Data: data})
}
