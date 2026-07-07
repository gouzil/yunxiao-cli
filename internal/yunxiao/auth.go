package yunxiao

import (
	"context"

	"github.com/gouzil/yunxiao-cli/internal/api"
	"github.com/gouzil/yunxiao-cli/internal/auth"
)

type AuthService interface {
	GetUserByToken(ctx context.Context, request GetUserByTokenRequest) (GetUserByTokenResult, error)
}

type GetUserByTokenRequest struct {
	Endpoint string `json:"endpoint"`
	Token    string `json:"token"`
}

type GetUserByTokenResult struct {
	User   User     `json:"user"`
	Scopes []string `json:"scopes"`
}

type getUserByTokenResponse struct {
	RequestID string   `json:"requestId"`
	Success   bool     `json:"success"`
	ID        string   `json:"id"`
	UserID    string   `json:"userId"`
	AccountID string   `json:"accountId"`
	Name      string   `json:"name"`
	NickName  string   `json:"nickName"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	LastOrg   string   `json:"lastOrganization"`
	OrgID     string   `json:"organizationId"`
	OrgName   string   `json:"organizationName"`
	Scopes    []string `json:"scopes"`
	Data      userData `json:"data"`
}

type userData struct {
	UserID    string   `json:"userId"`
	AccountID string   `json:"accountId"`
	Name      string   `json:"name"`
	NickName  string   `json:"nickName"`
	Email     string   `json:"email"`
	OrgID     string   `json:"organizationId"`
	OrgName   string   `json:"organizationName"`
	Scopes    []string `json:"scopes"`
}

func (s ClientServices) GetUserByToken(ctx context.Context, request GetUserByTokenRequest) (GetUserByTokenResult, error) {
	client := s.client.WithEndpoint(request.Endpoint).WithToken(request.Token)
	response := getUserByTokenResponse{}
	_, err := client.Do(ctx, api.Request{Method: "GET", Path: "/oapi/v1/platform/user"}, &response)
	if err != nil {
		return GetUserByTokenResult{}, err
	}
	data := response.Data
	user := User{
		ID:           first(data.UserID, response.ID, response.UserID, data.AccountID, response.AccountID),
		DisplayName:  first(data.NickName, data.Name, response.NickName, response.Name),
		Username:     first(response.Username, data.Name, response.Name),
		Email:        first(data.Email, response.Email),
		Organization: first(data.OrgName, response.OrgName, data.OrgID, response.OrgID, response.LastOrg),
	}
	return GetUserByTokenResult{User: user, Scopes: firstStrings(data.Scopes, response.Scopes)}, nil
}

type TokenVerifier struct {
	service AuthService
}

func NewTokenVerifier(service AuthService) TokenVerifier {
	return TokenVerifier{service: service}
}

func (v TokenVerifier) VerifyToken(ctx context.Context, endpoint string, token string) (auth.AuthenticatedUser, []string, error) {
	result, err := v.service.GetUserByToken(ctx, GetUserByTokenRequest{Endpoint: endpoint, Token: token})
	if err != nil {
		return auth.AuthenticatedUser{}, nil, err
	}
	return auth.AuthenticatedUser{
		ID:           result.User.ID,
		DisplayName:  result.User.DisplayName,
		Username:     result.User.Username,
		Email:        result.User.Email,
		Organization: result.User.Organization,
	}, result.Scopes, nil
}

func first(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func firstStrings(values ...[]string) []string {
	for _, value := range values {
		if len(value) > 0 {
			return value
		}
	}
	return nil
}
