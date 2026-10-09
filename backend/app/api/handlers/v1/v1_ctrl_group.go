package v1

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/samber/lo"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
	"github.com/sysadminsmedia/homebox/backend/internal/web/adapters"
)

type (
	GroupInvitationCreate struct {
		Uses      int       `json:"uses"      validate:"required,min=1,max=100"`
		ExpiresAt time.Time `json:"expiresAt"`
	}

	GroupInvitation struct {
		ID        uuid.UUID `json:"id"`
		Token     string    `json:"token"`
		ExpiresAt time.Time `json:"expiresAt"`
		Uses      int       `json:"uses"`
	}

	GroupAcceptInvitationResponse struct {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
	}

	CreateRequest struct {
		Name string `json:"name" validate:"required"`
	}
)

// HandleGroupGet godoc
func (ctrl *V1Controller) HandleGroupGet() errchain.HandlerFunc {
	fn := func(r *http.Request) (repo.Group, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.Groups.GroupByID(auth, auth.GID)
	}

	return adapters.Command(fn, http.StatusOK)
}

// HandleGroupUpdate godoc
func (ctrl *V1Controller) HandleGroupUpdate() errchain.HandlerFunc {
	fn := func(r *http.Request, body repo.GroupUpdate) (repo.Group, error) {
		actor := services.UseUserCtx(r.Context())
		if !actor.IsSuperuser {
			return repo.Group{}, errors.New("forbidden: only superusers can update collections")
		}

		auth := services.NewContext(r.Context())

		ok := ctrl.svc.Currencies.IsSupported(body.Currency)
		if !ok {
			return repo.Group{}, validate.NewFieldErrors(
				validate.NewFieldError("currency", "currency '"+body.Currency+"' is not supported"),
			)
		}

		return ctrl.svc.Group.UpdateGroup(auth, body)
	}

	return adapters.Action(fn, http.StatusOK)
}

// HandleGroupInvitationsCreate godoc
func (ctrl *V1Controller) HandleGroupInvitationsCreate() errchain.HandlerFunc {
	fn := func(r *http.Request, body GroupInvitationCreate) (GroupInvitation, error) {
		actor := services.UseUserCtx(r.Context())
		if !actor.IsSuperuser {
			return GroupInvitation{}, errors.New("forbidden: only superusers can create invitations")
		}

		if body.ExpiresAt.IsZero() {
			body.ExpiresAt = time.Now().Add(time.Hour * 24)
		}

		auth := services.NewContext(r.Context())

		invitation, token, err := ctrl.svc.Group.NewInvitation(auth, body.Uses, body.ExpiresAt)
		if err != nil {
			return GroupInvitation{}, err
		}

		return GroupInvitation{
			ID:        invitation.ID,
			Token:     token,
			ExpiresAt: invitation.ExpiresAt,
			Uses:      invitation.Uses,
		}, nil
	}

	return adapters.Action(fn, http.StatusCreated)
}

// HandleGroupsGetAll godoc
func (ctrl *V1Controller) HandleGroupsGetAll() errchain.HandlerFunc {
	fn := func(r *http.Request) ([]repo.Group, error) {
		// No se bloquea para que el usuario pueda ver el listado de colecciones a las que pertenece
		auth := services.NewContext(r.Context())
		return ctrl.repo.Groups.GetAllGroups(auth, auth.UID)
	}

	return adapters.Command(fn, http.StatusOK)
}

// HandleGroupCreate godoc
func (ctrl *V1Controller) HandleGroupCreate() errchain.HandlerFunc {
	fn := func(r *http.Request, body CreateRequest) (repo.Group, error) {
		actor := services.UseUserCtx(r.Context())
		if !actor.IsSuperuser {
			return repo.Group{}, errors.New("forbidden: only superusers can create collections")
		}

		auth := services.NewContext(r.Context())
		return ctrl.svc.Group.CreateGroup(auth, body.Name)
	}

	return adapters.Action(fn, http.StatusCreated)
}

// HandleGroupDelete godoc
func (ctrl *V1Controller) HandleGroupDelete() errchain.HandlerFunc {
	fn := func(r *http.Request) (any, error) {
		actor := services.UseUserCtx(r.Context())
		if !actor.IsSuperuser {
			return nil, errors.New("forbidden: only superusers can delete collections")
		}

		auth := services.NewContext(r.Context())

		currentUser, err := ctrl.repo.Users.GetOneID(auth, auth.UID)
		if err != nil {
			return nil, err
		}

		if len(currentUser.GroupIDs) <= 1 {
			return nil, validate.NewRequestError(errors.New("cannot delete the only group you are a member of"), http.StatusBadRequest)
		}

		if currentUser.DefaultGroupID == auth.GID {
			newDefaultGroupID, _ := lo.Find(currentUser.GroupIDs, func(gid uuid.UUID) bool {
				return gid != auth.GID
			})

			if err := ctrl.repo.Users.UpdateDefaultGroup(auth, auth.UID, newDefaultGroupID); err != nil {
				return nil, err
			}
		}

		err = ctrl.svc.Group.DeleteGroup(auth)
		return nil, err
	}

	return adapters.Command(fn, http.StatusNoContent)
}

// HandleGroupInvitationsGetAll godoc
func (ctrl *V1Controller) HandleGroupInvitationsGetAll() errchain.HandlerFunc {
	fn := func(r *http.Request) ([]repo.GroupInvitation, error) {
		actor := services.UseUserCtx(r.Context())
		if !actor.IsSuperuser {
			// Degradación elegante
			return []repo.GroupInvitation{}, nil
		}

		auth := services.NewContext(r.Context())
		return ctrl.repo.Groups.InvitationGetAll(auth, auth.GID)
	}

	return adapters.Command(fn, http.StatusOK)
}

// HandleGroupMembersGetAll godoc
func (ctrl *V1Controller) HandleGroupMembersGetAll() errchain.HandlerFunc {
	fn := func(r *http.Request) ([]repo.UserSummary, error) {
		actor := services.UseUserCtx(r.Context())
		if !actor.IsSuperuser {
			// Degradación elegante
			return []repo.UserSummary{}, nil
		}

		auth := services.NewContext(r.Context())
		return ctrl.repo.Users.GetUsersByGroupID(auth, auth.GID)
	}

	return adapters.Command(fn, http.StatusOK)
}

// HandleGroupMemberRemove godoc
func (ctrl *V1Controller) HandleGroupMemberRemove() errchain.HandlerFunc {
	fn := func(r *http.Request, userID uuid.UUID) (any, error) {
		actor := services.UseUserCtx(r.Context())
		if !actor.IsSuperuser {
			return nil, errors.New("forbidden: only superusers can remove members")
		}

		auth := services.NewContext(r.Context())

		if userID == auth.UID {
			return nil, validate.NewRequestError(errors.New("cannot remove yourself from the group"), http.StatusBadRequest)
		}

		members, err := ctrl.repo.Users.GetUsersByGroupID(auth, auth.GID)
		if err != nil {
			return nil, err
		}
		if len(members) <= 1 {
			return nil, validate.NewRequestError(errors.New("cannot remove the last member from the group"), http.StatusBadRequest)
		}

		err = ctrl.svc.Group.RemoveMember(auth, userID)
		return nil, err
	}

	return adapters.CommandID("user_id", fn, http.StatusNoContent)
}

// HandleGroupInvitationsDelete godoc
func (ctrl *V1Controller) HandleGroupInvitationsDelete() errchain.HandlerFunc {
	fn := func(r *http.Request, id uuid.UUID) (any, error) {
		actor := services.UseUserCtx(r.Context())
		if !actor.IsSuperuser {
			return nil, errors.New("forbidden: only superusers can delete invitations")
		}

		auth := services.NewContext(r.Context())
		err := ctrl.svc.Group.DeleteInvitation(auth, id)
		return nil, err
	}

	return adapters.CommandID("id", fn, http.StatusNoContent)
}

// HandleGroupInvitationsAccept godoc
func (ctrl *V1Controller) HandleGroupInvitationsAccept() errchain.HandlerFunc {
	fn := func(r *http.Request) (GroupAcceptInvitationResponse, error) {
		// Se permite a cualquier usuario aceptar invitaciones
		token := chi.URLParam(r, "id")
		if token == "" {
			return GroupAcceptInvitationResponse{}, validate.NewRequestError(errors.New("token is required"), http.StatusBadRequest)
		}

		auth := services.NewContext(r.Context())
		group, err := ctrl.svc.Group.AcceptInvitation(auth, token)
		if err != nil {
			if errors.Is(err, errors.New("user already a member of this group")) {
				return GroupAcceptInvitationResponse{}, validate.NewRequestError(err, http.StatusBadRequest)
			}
			return GroupAcceptInvitationResponse{}, err
		}

		return GroupAcceptInvitationResponse{ID: group.ID, Name: group.Name}, nil
	}

	return adapters.Command(fn, http.StatusOK)
}
