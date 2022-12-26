package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/strolt/strolt/apps/stroltm/internal/config"
	"github.com/strolt/strolt/shared/apiu"
)

type authValidateBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// authValidate godoc
// @Id					 validate
// @Summary      Validate user creds
// @Tags         auth
// @Param request body authValidateBody true "body"
// @success 200 {object} apiu.ResultSuccess
// @success 500 {object} apiu.ResultError
// @Router       /api/v1/auth/validate [post].
func (api *API) authValidate(w http.ResponseWriter, r *http.Request) {
	var body authValidateBody

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apiu.RenderJSON400(w, r, fmt.Errorf("invalid body"))
		return
	}

	user, ok := config.Get().API.Users[body.Username]
	if !ok {
		apiu.RenderJSON401(w, r)
		return
	}

	if user.Password != body.Password { //pragma: allowlist secret
		apiu.RenderJSON401(w, r)
		return
	}

	apiu.RenderJSON200(w, r, "")
}
