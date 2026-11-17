package auth

import (
	"io"
	"testing"

	authClient "api-tests-template/internal/client/http/auth"
	"api-tests-template/internal/managers/auth/models"

	"github.com/tidwall/gjson"
)

// RegisterTestUser регистрирует тестового пользователя
// и возвращает его токен и данные.
func RegisterTestUser(
	t *testing.T,
	request authClient.RegisterRequest,
) models.LoginOkResponse {
	t.Helper()

	response := authClient.HttpPostAuthRegister(t, request)

	if response == nil {
		t.Fatal("Сервер не вернул ответ на регистрацию")
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("Ошибка чтения ответа регистрации: %v", err)
	}

	token := gjson.GetBytes(body, "token").String()
	userID := gjson.GetBytes(body, "user.id").String()

	if token == "" || userID == "" {
		t.Fatalf("Регистрация не вернула токен или ID пользователя: %s", string(body))
	}

	return models.LoginOkResponse{
		Token: token,
		User: models.User{
			Id: userID,
		},
	}
}
