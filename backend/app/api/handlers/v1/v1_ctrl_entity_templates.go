package v1

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/samber/lo"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/web/adapters"
)

// HandleEntityTemplatesGetAll godoc
func (ctrl *V1Controller) HandleEntityTemplatesGetAll() errchain.HandlerFunc {
	fn := func(r *http.Request) ([]repo.EntityTemplateSummary, error) {
		actor := services.UseUserCtx(r.Context())
		if !actor.IsSuperuser {
			// Degradación elegante: devolvemos lista vacía
			return []repo.EntityTemplateSummary{}, nil
		}

		auth := services.NewContext(r.Context())
		return ctrl.repo.EntityTemplates.GetAll(r.Context(), auth.GID)
	}

	return adapters.Command(fn, http.StatusOK)
}

// HandleEntityTemplatesGet godoc
func (ctrl *V1Controller) HandleEntityTemplatesGet() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID) (repo.EntityTemplateOut, error) {
		actor := services.UseUserCtx(r.Context())
		if !actor.IsSuperuser {
			return repo.EntityTemplateOut{}, errors.New("forbidden: only superusers can view templates")
		}

		auth := services.NewContext(r.Context())
		return ctrl.repo.EntityTemplates.GetOne(r.Context(), auth.GID, ID)
	}

	return adapters.CommandID("id", fn, http.StatusOK)
}

// HandleEntityTemplatesCreate godoc
func (ctrl *V1Controller) HandleEntityTemplatesCreate() errchain.HandlerFunc {
	fn := func(r *http.Request, body repo.EntityTemplateCreate) (repo.EntityTemplateOut, error) {
		actor := services.UseUserCtx(r.Context())
		if !actor.IsSuperuser {
			return repo.EntityTemplateOut{}, errors.New("forbidden: only superusers can create templates")
		}

		auth := services.NewContext(r.Context())
		return ctrl.repo.EntityTemplates.Create(r.Context(), auth.GID, body)
	}

	return adapters.Action(fn, http.StatusCreated)
}

// HandleEntityTemplatesUpdate godoc
func (ctrl *V1Controller) HandleEntityTemplatesUpdate() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID, body repo.EntityTemplateUpdate) (repo.EntityTemplateOut, error) {
		actor := services.UseUserCtx(r.Context())
		if !actor.IsSuperuser {
			return repo.EntityTemplateOut{}, errors.New("forbidden: only superusers can update templates")
		}

		auth := services.NewContext(r.Context())
		body.ID = ID
		return ctrl.repo.EntityTemplates.Update(r.Context(), auth.GID, body)
	}

	return adapters.ActionID("id", fn, http.StatusOK)
}

// HandleEntityTemplatesDelete godoc
func (ctrl *V1Controller) HandleEntityTemplatesDelete() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID) (any, error) {
		actor := services.UseUserCtx(r.Context())
		if !actor.IsSuperuser {
			return nil, errors.New("forbidden: only superusers can delete templates")
		}

		auth := services.NewContext(r.Context())
		err := ctrl.repo.EntityTemplates.Delete(r.Context(), auth.GID, ID)
		return nil, err
	}

	return adapters.CommandID("id", fn, http.StatusNoContent)
}

type EntityTemplateCreateItemRequest struct {
	Name         string      `json:"name"        validate:"required,min=1,max=255"`
	Description  string      `json:"description" validate:"max=1000"`
	ParentID     uuid.UUID   `json:"parentId"    validate:"required"`
	EntityTypeID uuid.UUID   `json:"entityTypeId"`
	TagIDs       []uuid.UUID `json:"tagIds"`
	Quantity     *float64    `json:"quantity"`
}

// HandleEntityTemplatesCreateItem godoc
func (ctrl *V1Controller) HandleEntityTemplatesCreateItem() errchain.HandlerFunc {
	fn := func(r *http.Request, templateID uuid.UUID, body EntityTemplateCreateItemRequest) (repo.EntityOut, error) {
		actor := services.UseUserCtx(r.Context())
		if !actor.IsSuperuser {
			return repo.EntityOut{}, errors.New("forbidden: only superusers can create items from templates")
		}

		auth := services.NewContext(r.Context())

		template, err := ctrl.repo.EntityTemplates.GetOne(r.Context(), auth.GID, templateID)
		if err != nil {
			return repo.EntityOut{}, err
		}

		quantity := template.DefaultQuantity
		if body.Quantity != nil {
			quantity = *body.Quantity
		}

		fields := lo.Map(template.Fields, func(f repo.TemplateField, _ int) repo.EntityFieldData {
			return repo.EntityFieldData{
				Type:         f.Type,
				Name:         f.Name,
				TextValue:    f.TextValue,
				NumberValue:  f.NumberValue,
				BooleanValue: f.BooleanValue,
			}
		})

		return ctrl.repo.Entities.CreateFromTemplate(r.Context(), auth.GID, repo.EntityCreateFromTemplate{
			Name:             body.Name,
			Description:      body.Description,
			Quantity:         quantity,
			ParentID:         body.ParentID,
			EntityTypeID:     body.EntityTypeID,
			TagIDs:           body.TagIDs,
			Insured:          template.DefaultInsured,
			Manufacturer:     template.DefaultManufacturer,
			ModelNumber:      template.DefaultModelNumber,
			LifetimeWarranty: template.DefaultLifetimeWarranty,
			WarrantyDetails:  template.DefaultWarrantyDetails,
			Fields:           fields,
		})
	}

	return adapters.ActionID("id", fn, http.StatusCreated)
}
