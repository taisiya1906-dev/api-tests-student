package assertions

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"api-tests-template/internal/managers/advertisements/models"

	"github.com/stretchr/testify/require"
)

// ContainsAdvertisementByTitle проверяет наличие объявления по названию.
func ContainsAdvertisementByTitle(
	t *testing.T,
	items []models.Advertisement,
	title string,
) {
	t.Helper()

	for _, item := range items {
		if strings.EqualFold(item.Title, title) {
			return
		}
	}

	require.Fail(
		t,
		"Объявление с названием не найдено",
		"Искали: %s",
		title,
	)
}

// ContainsAdvertisementById проверяет наличие объявления по ID.
func ContainsAdvertisementById(
	t *testing.T,
	items []models.Advertisement,
	id string,
) {
	t.Helper()

	for _, item := range items {
		if item.ID == id {
			return
		}
	}

	require.Fail(
		t,
		"Объявление с ID не найдено",
		"Искали ID: %s",
		id,
	)
}

// AssertAPIError проверяет тип и сообщение ошибки в HTTP-ответе.
// Также закрывает тело ответа после чтения.
func AssertAPIError(
	t *testing.T,
	response *http.Response,
	expectedError string,
	expectedMessage string,
) {
	t.Helper()

	require.NotNil(
		t,
		response,
		"Сервер не вернул HTTP-ответ при проверке ошибки API",
	)

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)

	require.NoError(
		t,
		err,
		"Не удалось прочитать тело ответа API",
	)

	require.JSONEq(
		t,
		`{"error":"`+expectedError+`","message":"`+expectedMessage+`"}`,
		string(body),
		"Ответ API не соответствует ожидаемой ошибке",
	)
}
