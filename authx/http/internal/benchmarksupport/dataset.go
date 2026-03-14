package benchmarksupport

import (
	"fmt"
	"strings"

	"github.com/brianvoe/gofakeit/v7"
)

type Query struct {
	UserID   string
	Action   string
	Resource string
	Allowed  bool
}

type Dataset struct {
	userPermissions map[string]map[string]struct{}
	Queries         []Query
}

func NewDataset(
	seed uint64,
	userCount int,
	permissionCount int,
	permissionsPerUser int,
	queryCount int,
) Dataset {
	randSource := gofakeit.New(seed)

	permissions := make([]string, permissionCount)
	for i := 0; i < permissionCount; i++ {
		action := fmt.Sprintf("%s-%03d", normalizeFakeToken(randSource.Verb()), i/100)
		resource := fmt.Sprintf("%s-%03d", normalizeFakeToken(randSource.Noun()), i%100)
		permissions[i] = permissionKey(action, resource)
	}

	userIDs := make([]string, userCount)
	userPermissions := make(map[string]map[string]struct{}, userCount)
	for i := 0; i < userCount; i++ {
		userID := fmt.Sprintf("%s-%05d", normalizeFakeToken(randSource.Username()), i)
		userIDs[i] = userID

		assigned := make(map[string]struct{}, permissionsPerUser)
		for len(assigned) < permissionsPerUser {
			assigned[permissions[randSource.Number(0, len(permissions)-1)]] = struct{}{}
		}
		userPermissions[userID] = assigned
	}

	queries := make([]Query, queryCount)
	for i := 0; i < queryCount; i++ {
		userID := userIDs[randSource.Number(0, len(userIDs)-1)]
		assigned := userPermissions[userID]

		permission := samplePermission(randSource, assigned)
		allowed := true
		if i%2 == 1 {
			allowed = false
			for {
				candidate := permissions[randSource.Number(0, len(permissions)-1)]
				if _, exists := assigned[candidate]; !exists {
					permission = candidate
					break
				}
			}
		}

		action, resource := parsePermissionKey(permission)
		queries[i] = Query{
			UserID:   userID,
			Action:   action,
			Resource: resource,
			Allowed:  allowed,
		}
	}

	return Dataset{
		userPermissions: userPermissions,
		Queries:         queries,
	}
}

func (dataset Dataset) IsAllowed(userID string, action string, resource string) bool {
	permissions, ok := dataset.userPermissions[userID]
	if !ok {
		return false
	}
	_, allowed := permissions[permissionKey(action, resource)]
	return allowed
}

func (dataset Dataset) HasUser(userID string) bool {
	_, ok := dataset.userPermissions[userID]
	return ok
}

func permissionKey(action string, resource string) string {
	return action + "|" + resource
}

func parsePermissionKey(key string) (string, string) {
	action, resource, found := strings.Cut(key, "|")
	if !found {
		return key, ""
	}
	return action, resource
}

func samplePermission(randSource *gofakeit.Faker, assigned map[string]struct{}) string {
	target := randSource.Number(0, len(assigned)-1)
	for permission := range assigned {
		if target == 0 {
			return permission
		}
		target--
	}
	return ""
}

func normalizeFakeToken(raw string) string {
	token := strings.ToLower(strings.TrimSpace(raw))
	token = strings.ReplaceAll(token, " ", "_")
	token = strings.ReplaceAll(token, "-", "_")
	if token == "" {
		return "x"
	}
	return token
}
