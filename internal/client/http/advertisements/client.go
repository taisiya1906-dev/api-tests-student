package advertisements

import (
	"net/http"
	"testing"

	"api-tests-template/internal/constants/path"
	api_runner "api-tests-template/internal/helpers/api-runner"
)

func HttpGetAdvertisements(t *testing.T, token string, textForSearch string) *http.Response {
	return api_runner.GetRunner().Auth(token).Create().
		Get(path.AdvertisementsPath).
		Query("search", textForSearch).
		Expect(t).Status(http.StatusOK).End().Response
}

func HttpCreateAdvertisement(
	t *testing.T,
	token string,
	photoPath string,
) *http.Response {
	return api_runner.GetRunner().
		Auth(token).
		Create().
		Post(path.AdvertisementPath).
		MultipartFormData("title", "Тестовое объявление").
		MultipartFormData("description", "Описание тестового объявления").
		MultipartFormData("price", "1500").
		MultipartFormData("quantity", "1").
		MultipartFile("photos", photoPath).
		Expect(t).
		Status(http.StatusCreated).
		End().
		Response
}

// HttpGetAdvertisementByID получает объявление по его ID.
func HttpGetAdvertisementByID(
	t *testing.T,
	token string,
	advertisementID string,
) *http.Response {
	return api_runner.GetRunner().
		Auth(token).
		Create().
		Get(path.AdvertisementPath).
		Query("id", advertisementID).
		Expect(t).
		Status(http.StatusOK).
		End().
		Response
}

// HttpDeleteAdvertisement удаляет объявление по его ID.
func HttpDeleteAdvertisement(
	t *testing.T,
	token string,
	advertisementID string,
) *http.Response {
	return api_runner.GetRunner().
		Auth(token).
		Create().
		Delete(path.AdvertisementPath).
		Query("id", advertisementID).
		Expect(t).
		Status(http.StatusNoContent).
		End().
		Response
}

// HttpGetAdvertisementPhotos получает фотографии объявления по ID.
func HttpGetAdvertisementPhotos(
	t *testing.T,
	token string,
	advertisementID string,
) *http.Response {
	return api_runner.GetRunner().
		Auth(token).
		Create().
		Get(path.AdvertisementsPath + "/" + advertisementID + "/photos").
		Expect(t).
		Status(http.StatusOK).
		End().
		Response
}

// HttpCreateAdvertisementWithParams создаёт объявление
// с переданными параметрами и проверяет ожидаемый HTTP-статус.
func HttpCreateAdvertisementWithParams(
	t *testing.T,
	token string,
	title string,
	description string,
	price string,
	quantity string,
	photoPath string,
	expectedStatus int,
) *http.Response {
	return api_runner.GetRunner().
		Auth(token).
		Create().
		Post(path.AdvertisementPath).
		MultipartFormData("title", title).
		MultipartFormData("description", description).
		MultipartFormData("price", price).
		MultipartFormData("quantity", quantity).
		MultipartFile("photos", photoPath).
		Expect(t).
		Status(expectedStatus).
		End().
		Response
}

// HttpCreateAdvertisementWithoutPhoto отправляет запрос

func HttpGetAdvertisementsForFavorites(
	t *testing.T,
	token string,
) *http.Response {
	return api_runner.GetRunner().
		Auth(token).
		Create().
		Get(path.AdvertisementsPath).
		Expect(t).
		Status(http.StatusOK).
		End().
		Response
}

// HttpCreateAdvertisementWithoutPhoto отправляет запрос
// на создание объявления без фотографии.
func HttpCreateAdvertisementWithoutPhoto(
	t *testing.T,
	token string,
	expectedStatus int,
) *http.Response {
	return api_runner.GetRunner().
		Auth(token).
		Create().
		Post(path.AdvertisementPath).
		MultipartFormData("title", "Тестовое объявление").
		MultipartFormData("description", "Описание тестового объявления").
		MultipartFormData("price", "1500").
		MultipartFormData("quantity", "1").
		Expect(t).
		Status(expectedStatus).
		End().
		Response
}
