package v1

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/web/adapters"
)

// HandleMaintenanceGetAll godoc
func (ctrl *V1Controller) HandleMaintenanceGetAll() errchain.HandlerFunc {
	fn := func(r *http.Request, filters repo.MaintenanceFilters) ([]repo.MaintenanceEntryWithDetails, error) {
		actor := services.UseUserCtx(r.Context())
		if !actor.IsSuperuser {
			// Degradación elegante: devolvemos lista vacía
			return []repo.MaintenanceEntryWithDetails{}, nil
		}

		auth := services.NewContext(r.Context())
		return ctrl.repo.MaintEntry.GetAllMaintenance(auth, auth.GID, filters)
	}

	return adapters.Query(fn, http.StatusOK)
}

// HandleMaintenanceEntryUpdate godoc
func (ctrl *V1Controller) HandleMaintenanceEntryUpdate() errchain.HandlerFunc {
	fn := func(r *http.Request, entryID uuid.UUID, body repo.MaintenanceEntryUpdate) (repo.MaintenanceEntry, error) {
		actor := services.UseUserCtx(r.Context())
		if !actor.IsSuperuser {
			return repo.MaintenanceEntry{}, errors.New("forbidden: only superusers can update maintenance entries")
		}

		auth := services.NewContext(r.Context())
		return ctrl.repo.MaintEntry.Update(auth, auth.GID, entryID, body)
	}

	return adapters.ActionID("id", fn, http.StatusOK)
}

// HandleMaintenanceEntryDelete godoc
func (ctrl *V1Controller) HandleMaintenanceEntryDelete() errchain.HandlerFunc {
	fn := func(r *http.Request, entryID uuid.UUID) (any, error) {
		actor := services.UseUserCtx(r.Context())
		if !actor.IsSuperuser {
			return nil, errors.New("forbidden: only superusers can delete maintenance entries")
		}

		auth := services.NewContext(r.Context())
		err := ctrl.repo.MaintEntry.Delete(auth, auth.GID, entryID)
		return nil, err
	}

	return adapters.CommandID("id", fn, http.StatusNoContent)
}
