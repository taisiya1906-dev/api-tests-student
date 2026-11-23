package favorites

import (
	"net/http"
	"testing"

	api_runner "api-tests-template/internal/helpers/api-runner"
)

const favoritesPath = "/favorites"

// HttpAddToFavorites добавляет объявление в избранное.
func HttpAddToFavorites(
	t *testing.T,
	token string,
	advertisementID string,
	expectedStatus int,
) *http.Response {
	return api_runner.GetRunner().
		Auth(token).
		Create().
		Post(favoritesPath).
		JSON(map[string]string{
			"advertisement_id": advertisementID,
		}).
		Expect(t).
		Status(expectedStatus).
		End().
		Response
}

// HttpGetFavorites получает список избранного
// и проверяет ожидаемый HTTP-статус.
func HttpGetFavorites(
	t *testing.T,
	token string,
	expectedStatus int,
) *http.Response {
	return api_runner.GetRunner().
		Auth(token).
		Create().
		Get(favoritesPath).
		Expect(t).
		Status(expectedStatus).
		End().
		Response
}

// HttpRemoveFromFavorites удаляет объявление из избранного.
func HttpRemoveFromFavorites(
	t *testing.T,
	token string,
	advertisementID string,
	expectedStatus int,
) *http.Response {
	return api_runner.GetRunner().
		Auth(token).
		Create().
		Delete(favoritesPath + "/" + advertisementID).
		Expect(t).
		Status(expectedStatus).
		End().
		Response
}
