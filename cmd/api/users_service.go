package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/CepeshIII/Project07_API_Server/internal/store"
	"github.com/CepeshIII/Project07_API_Server/internal/store/cache"
	"github.com/go-chi/chi/v5"
)

func (app *application) getUser(ctx context.Context, userID int64) (*store.User, error) {
	if app.config.Redis.Enabled {
		user, err := app.cacheStorage.Users.GetUserByID(ctx, userID)
		if err == nil {
			return user, nil // Cache hit!
		}

		// If it's anything other than a normal cache miss, log the Redis failure
		if !errors.Is(err, cache.ErrCacheMissed) {
			app.logger.Errorw("redis cache get error", "userID", userID, "error", err)
		}
		// If it was a cache miss, we silently fall through to the database below
	}

	// Fallback to database store
	user, err := app.store.Users.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Populate cache asynchronously or synchronously on a cache miss
	if app.config.Redis.Enabled {
		err := app.cacheStorage.Users.SetUser(ctx, user)
		if err != nil {
			app.logger.Errorw("cache store error", "userID", userID, "error", err.Error())
		}
	}

	return user, nil
}

func (app *application) getRole(ctx context.Context, roleID int64) (*store.Role, error) {
	if app.config.Redis.Enabled {
		role, err := app.cacheStorage.Roles.GetRoleByID(ctx, roleID)
		if err == nil {
			return role, nil
		}

		if !errors.Is(err, cache.ErrCacheMissed) {
			app.logger.Errorw("cache store error", "roleId", roleID, "error", err.Error())
		}
	}

	role, err := app.store.Roles.GetRoleByID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	if app.config.Redis.Enabled {
		err := app.cacheStorage.Roles.SetRole(ctx, role)
		if err != nil {
			app.logger.Errorw("cache store error", "roleId", roleID, "error", err.Error())
		}
	}

	return role, nil
}

func (app *application) getRoleByName(
	ctx context.Context,
	roleName string,
) (*store.Role, error) {
	if app.config.Redis.Enabled {
		role, err := app.cacheStorage.Roles.GetRoleByName(ctx, roleName)
		if err == nil {
			return role, nil
		}

		if !errors.Is(err, cache.ErrCacheMissed) {
			app.logger.Errorw("cache store error", "roleName", roleName, "error", err.Error())
		}
	}

	role, err := app.store.Roles.GetRoleByName(ctx, roleName)
	if err != nil {
		return nil, err
	}

	if app.config.Redis.Enabled {
		err := app.cacheStorage.Roles.SetRole(ctx, role)
		if err != nil {
			app.logger.Errorw("cache store error", "roleName", roleName, "error", err.Error())
		}
	}

	return role, nil
}

func (app *application) getUserWithRole(ctx context.Context, userID int64) (*store.UserWithRole, error) {
	user, err := app.getUser(ctx, userID)
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
		return 0, errInvalidUserID
	}

	if id <= 0 {
		return 0, errInvalidUserID
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
