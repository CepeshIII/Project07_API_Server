package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/CepeshIII/Project07_API_Server/internal/store"
	"github.com/go-chi/chi/v5"
)

func (app *application) getUser(ctx context.Context, userId int64) (*store.User, error) {
	if app.config.redisCfg.enabled {
		user, err := app.cacheStorage.Users.GetUserByID(ctx, userId)
		if err != nil {
			app.logger.Errorw("cache store error", "userId", userId, "error", err.Error())
		} else {
			if user != nil {
				fmt.Println("Has been get from cache store")
				return user, nil
			}
		}
	}

	user, err := app.store.Users.Get(ctx, userId)
	if err != nil {
		return nil, err
	}

	if app.config.redisCfg.enabled {
		if err := app.cacheStorage.Users.SetUser(ctx, user); err != nil {
			app.logger.Errorw("cache store error", "userId", userId, "error", err.Error())
		}
	}

	fmt.Println("Has been get from database")
	return user, nil
}

func (app *application) getRole(ctx context.Context, roleId int64) (*store.Role, error) {
	if app.config.redisCfg.enabled {
		role, err := app.cacheStorage.Roles.GetRoleByID(ctx, roleId)
		if err != nil {
			app.logger.Errorw("cache store error", "roleId", roleId, "error", err.Error())
		} else {
			if role != nil {
				return role, nil
			}
		}

	}

	role, err := app.store.Roles.GetRoleByID(ctx, roleId)
	if err != nil {
		return nil, err
	}

	if app.config.redisCfg.enabled {
		if err := app.cacheStorage.Roles.SetRole(ctx, role); err != nil {
			app.logger.Errorw("cache store error", "roleId", roleId, "error", err.Error())
		}
	}

	return role, nil
}

func (app *application) getRoleByName(
	ctx context.Context,
	roleName string,
) (*store.Role, error) {

	if app.config.redisCfg.enabled {
		role, err := app.cacheStorage.Roles.GetRoleByName(ctx, roleName)
		if err != nil {
			app.logger.Errorw("cache store error", "roleName", roleName, "error", err.Error())
		}

		if role != nil {
			return role, nil
		}
	}

	role, err := app.store.Roles.GetRoleByName(ctx, roleName)
	if err != nil {
		return nil, err
	}

	if app.config.redisCfg.enabled {
		if err := app.cacheStorage.Roles.SetRole(ctx, role); err != nil {
			app.logger.Errorw("cache store error", "roleName", roleName, "error", err.Error())
		}
	}

	return role, nil
}

func (app *application) getUserWithRole(ctx context.Context, userId int64) (*store.UserWithRole, error) {
	user, err := app.getUser(ctx, userId)
	if err != nil {
		return nil, err
	}

	role, err := app.getRole(ctx, user.RoleID)
	if err != nil {
		return nil, err
	}

	userWithRole := store.UserWithRole{
		User: *user,
		Role: *role,
	}

	return &userWithRole, nil
}

func parseUserID(r *http.Request) (int64, error) {
	idParam := chi.URLParam(r, "userID")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return 0, errors.New("invalid user ID")
	}
	if id <= 0 {
		return 0, errors.New("user ID must be a positive integer")
	}
	return id, nil
}

func (app *application) checkRolePrecedence(ctx context.Context, userWithRole *store.UserWithRole, roleName string) (bool, error) {
	role, err := app.getRoleByName(ctx, roleName)
	if err != nil {
		return false, err
	}

	return userWithRole.Role.Level >= role.Level, nil
}
