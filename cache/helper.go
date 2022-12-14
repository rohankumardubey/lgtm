package cache

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/go-gitea/lgtm/model"
	"github.com/go-gitea/lgtm/remote"
)

// GetRepos returns the list of user repositories from the cache
// associated with the current context.
func GetRepos(c context.Context, user *model.User) ([]*model.Repo, error) {
	key := fmt.Sprintf("repos:%s",
		user.Login,
	)
	// if we fetch from the cache we can return immediately
	val, err := FromContext(c).Get(key)
	if err == nil {
		return val.([]*model.Repo), nil
	}
	// else we try to grab from the remote system and
	// populate our cache.
	repos, err := remote.GetRepos(c, user)
	if err != nil {
		return nil, err
	}
	err = FromContext(c).Set(key, repos)
	if err != nil {
		return nil, err
	}
	return repos, nil
}

// GetTeams returns the list of user teams from the cache
// associated with the current context.
func GetTeams(c context.Context, user *model.User) ([]*model.Team, error) {
	key := fmt.Sprintf("teams:%s",
		user.Login,
	)
	// if we fetch from the cache we can return immediately
	val, err := FromContext(c).Get(key)
	if err == nil {
		return val.([]*model.Team), nil
	}
	// else we try to grab from the remote system and
	// populate our cache.
	teams, err := remote.GetTeams(c, user)
	if err != nil {
		return nil, err
	}

	err = FromContext(c).Set(key, teams)
	if err != nil {
		return nil, err
	}

	return teams, nil
}

// GetPerm returns the user permissions repositories from the cache
// associated with the current repository.
func GetPerm(c context.Context, user *model.User, owner, name string) (*model.Perm, error) {
	key := fmt.Sprintf("perms:%s:%s/%s",
		user.Login,
		owner,
		name,
	)
	// if we fetch from the cache we can return immediately
	val, err := FromContext(c).Get(key)
	if err == nil {
		return val.(*model.Perm), nil
	}
	// else we try to grab from the remote system and
	// populate our cache.
	perm, err := remote.GetPerm(c, user, owner, name)
	if err != nil {
		return nil, err
	}
	err = FromContext(c).Set(key, perm)
	if err != nil {
		return nil, err
	}
	return perm, nil
}

// GetMembers returns the team members from the cache.
func GetMembers(c context.Context, user *model.User, team string) ([]*model.Member, error) {
	key := fmt.Sprintf("members:%s",
		team,
	)
	// if we fetch from the cache we can return immediately
	val, err := FromContext(c).Get(key)
	if err == nil {
		return val.([]*model.Member), nil
	}
	// else we try to grab from the remote system and
	// populate our cache.
	members, err := remote.GetMembers(c, user, team)
	if err != nil {
		return nil, err
	}
	err = FromContext(c).Set(key, members)
	if err != nil {
		return nil, err
	}

	return members, nil
}
