package auth

import (
	"net/http"
	"testing"

	"api-tests-template/internal/constants/path"
	authModels "api-tests-template/internal/managers/auth/models"

	apiRunner "api-tests-template/internal/helpers/api-runner"
)

// HttpPostAuthLogin выполняет POST-запрос для авторизации пользователя
func HttpPostAuthLogin(t *testing.T, request authModels.LoginRequest) *http.Response {
	return apiRunner.GetRunner().Create().Post(path.AuthLoginPath).
		ContentType("application/json").
		JSON(request).
		Expect(t).
		Status(http.StatusOK).
		End().Response
}

// RegisterRequest содержит данные нового пользователя.
type RegisterRequest struct {
	BirthDate string `json:"birth_date"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Password  string `json:"password"`
}

// HttpPostAuthRegister регистрирует нового пользователя.
func HttpPostAuthRegister(
	t *testing.T,
	request RegisterRequest,
) *http.Response {
	return apiRunner.GetRunner().
		Create().
		Post("/auth/register").
		ContentType("application/json").
		JSON(request).
		Expect(t).
		Status(http.StatusCreated).
		End().
		Response
}
