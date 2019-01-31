package intercom

import (
	"encoding/json"
	"errors"
	"fmt"

	"gopkg.in/bebus77/intercom-go.v2/interfaces"
)

// UserRepository defines the interface for working with Users through the API.
type UserRepository interface {
	find(UserIdentifiers) (User, error)
	list(userListParams) (UserList, error)
	scroll(scrollParam string) (UserList, error)
	save(*User) (User, error)
	delete(id string) (User, error)
}

// UserAPI implements UserRepository
type UserAPI struct {
	httpClient interfaces.HTTPClient
}

type requestScroll struct {
	ScrollParam string `json:"scroll_param,omitempty"`
}

type requestUser struct {
	ID                     string                 `json:"id,omitempty"`
	Email                  string                 `json:"email,omitempty"`
	Phone                  string                 `json:"phone,omitempty"`
	UserID                 string                 `json:"user_id,omitempty"`
	Name                   string                 `json:"name,omitempty"`
	SignedUpAt             int64                  `json:"signed_up_at,omitempty"`
	RemoteCreatedAt        int64                  `json:"remote_created_at,omitempty"`
	LastRequestAt          int64                  `json:"last_request_at,omitempty"`
	LastSeenIP             string                 `json:"last_seen_ip,omitempty"`
	UnsubscribedFromEmails *bool                  `json:"unsubscribed_from_emails,omitempty"`
	Companies              []UserCompany          `json:"companies,omitempty"`
	CustomAttributes       map[string]interface{} `json:"custom_attributes,omitempty"`
	UpdateLastRequestAt    *bool                  `json:"update_last_request_at,omitempty"`
	NewSession             *bool                  `json:"new_session,omitempty"`
	LastSeenUserAgent      string                 `json:"last_seen_user_agent,omitempty"`
}

type userSearchResponse struct {
	Users []User
}

func (api UserAPI) find(params UserIdentifiers) (User, error) {
	var user User

	switch {
	case params.ID != "":
		return unmarshalToUser(api.httpClient.Get(fmt.Sprintf("/users/%s", params.ID), nil))
	case params.UserID != "", params.Email != "":
		return unmarshalListToUser(api.httpClient.Get("/users", params))
	}

	return user, errors.New("Missing User Identifier")
}

func (api UserAPI) list(params userListParams) (UserList, error) {
	userList := UserList{}
	data, err := api.httpClient.Get("/users", params)
	if err != nil {
		return userList, err
	}
	err = json.Unmarshal(data, &userList)
	return userList, err
}

func (api UserAPI) scroll(scrollParam string) (UserList, error) {
	userList := UserList{}

	url := "/users/scroll"
	params := scrollParams{ScrollParam: scrollParam}
	data, err := api.httpClient.Get(url, params)

	if err != nil {
		return userList, err
	}
	err = json.Unmarshal(data, &userList)
	return userList, err
}

func (api UserAPI) save(user *User) (User, error) {
	return unmarshalToUser(api.httpClient.Post("/users", RequestUserMapper{}.ConvertUser(user)))
}

func unmarshalToUser(data []byte, err error) (User, error) {
	savedUser := User{}
	if err != nil {
		return savedUser, err
	}
	err = json.Unmarshal(data, &savedUser)
	return savedUser, err
}

func unmarshalListToUser(data []byte, err error) (User, error) {
	saved := User{}
	if err != nil {
		return saved, err
	}

	resp := userSearchResponse{}
	if err := json.Unmarshal(data, &resp); err != nil {
		return saved, err
	}

	if len(resp.Users) == 0 {
		return saved, interfaces.HTTPError{
			StatusCode: 404,
			Code:       "not_found",
			Message:    "User Not Found",
		}
	}

	return resp.Users[0], nil
}

func (api UserAPI) delete(id string) (User, error) {
	user := User{}
	data, err := api.httpClient.Delete(fmt.Sprintf("/users/%s", id), nil)
	if err != nil {
		return user, err
	}
	err = json.Unmarshal(data, &user)
	return user, err
}
