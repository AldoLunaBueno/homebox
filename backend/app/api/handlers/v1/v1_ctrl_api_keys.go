package v1

import (
	"errors" // AGREGADA PARA MANEJAR EL RECHAZO
	"net/http"

	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/web/adapters"
)

// HandleUserAPIKeysList godoc
//
//	@Summary    List API Keys
//	@Tags       User
//	@Produce    json
//	@Success    200 {object}    []repo.APIKeyOut
//	@Router     /v1/users/self/api-keys [GET]
//	@Security   Bearer
func (ctrl *V1Controller) HandleUserAPIKeysList() errchain.HandlerFunc {
	fn := func(r *http.Request, _ struct{}) ([]repo.APIKeyOut, error) {
		actor := services.UseUserCtx(r.Context())

		// Le decimos pacíficamente que tiene 0 llaves.
		if !actor.IsSuperuser {
			return []repo.APIKeyOut{}, nil
		}

		return ctrl.svc.User.ListAPIKeys(r.Context(), actor.ID)
	}
	return adapters.Query(fn, http.StatusOK)
}

// HandleUserAPIKeyCreate godoc
//
//	@Summary    Create API Key
//	@Tags       User
//	@Produce    json
//	@Param      payload body        repo.APIKeyCreate   true    "API Key Data"
//	@Success    201     {object}    repo.APIKeyCreatedOut
//	@Router     /v1/users/self/api-keys [POST]
//	@Security   Bearer
func (ctrl *V1Controller) HandleUserAPIKeyCreate() errchain.HandlerFunc {
	fn := func(r *http.Request, in repo.APIKeyCreate) (repo.APIKeyCreatedOut, error) {
		actor := services.UseUserCtx(r.Context())

		// BLOQUEO DE SEGURIDAD BACKEND
		if !actor.IsSuperuser {
			return repo.APIKeyCreatedOut{}, errors.New("forbidden: only superusers can create API keys")
		}

		return ctrl.svc.User.CreateAPIKey(r.Context(), actor.ID, in)
	}
	return adapters.Action(fn, http.StatusCreated)
}

// HandleUserAPIKeyDelete godoc
//
//	@Summary    Delete API Key
//	@Tags       User
//	@Param      id  path    string  true    "API Key ID"
//	@Success    204
//	@Router     /v1/users/self/api-keys/{id} [DELETE]
//	@Security   Bearer
func (ctrl *V1Controller) HandleUserAPIKeyDelete() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID) (any, error) {
		actor := services.UseUserCtx(r.Context())

		// BLOQUEO DE SEGURIDAD BACKEND
		if !actor.IsSuperuser {
			return nil, errors.New("forbidden: only superusers can delete API keys")
		}

		return nil, ctrl.svc.User.DeleteAPIKey(r.Context(), actor.ID, ID)
	}
	return adapters.CommandID("id", fn, http.StatusNoContent)
}
