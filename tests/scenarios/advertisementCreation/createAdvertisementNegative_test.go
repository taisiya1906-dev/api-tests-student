package advertisementCreation

import (
	"net/http"
	"strings"

	"api-tests-template/internal/assertions"
	advertisementsClient "api-tests-template/internal/client/http/advertisements"
	"api-tests-template/internal/constants/testdata"
)

// TestCreateAdvertisementEmptyTitle проверяет,
// что объявление нельзя создать с пустым названием.
func (s *TestSuite) TestCreateAdvertisementEmptyTitle() {
	s.Run("Создаём объявление с пустым названием", func() {
		response := advertisementsClient.HttpCreateAdvertisementWithParams(
			s.T(),
			s.loginData.Token,
			"",
			"Описание тестового объявления",
			"1500",
			"1",
			testdata.PhotoPath,
			http.StatusBadRequest,
		)

		assertions.AssertAPIError(
			s.T(),
			response,
			"invalid_request",
			"title is required",
		)
	})
}

// TestCreateAdvertisementNegativePrice проверяет,
// что объявление нельзя создать с отрицательной ценой.
func (s *TestSuite) TestCreateAdvertisementNegativePrice() {
	s.Run("Создаём объявление с отрицательной ценой", func() {
		response := advertisementsClient.HttpCreateAdvertisementWithParams(
			s.T(),
			s.loginData.Token,
			"Тестовое объявление",
			"Описание тестового объявления",
			"-1500",
			"1",
			testdata.PhotoPath,
			http.StatusBadRequest,
		)

		assertions.AssertAPIError(
			s.T(),
			response,
			"invalid_request",
			"price must be between 0 and 1000000 inclusive",
		)
	})
}

// TestCreateAdvertisementLongTitle проверяет,
// что сервер отклоняет слишком длинное название.
func (s *TestSuite) TestCreateAdvertisementLongTitle() {
	s.Run("Создаём объявление со слишком длинным названием", func() {
		longTitle := strings.Repeat("А", 1000)

		response := advertisementsClient.HttpCreateAdvertisementWithParams(
			s.T(),
			s.loginData.Token,
			longTitle,
			"Описание тестового объявления",
			"1500",
			"1",
			testdata.PhotoPath,
			http.StatusBadRequest,
		)

		assertions.AssertAPIError(
			s.T(),
			response,
			"invalid_request",
			"title must not exceed 50 characters provided 1000",
		)
	})
}

// TestCreateAdvertisementWithoutAuthorization проверяет,
// что объявление нельзя создать без авторизации.
func (s *TestSuite) TestCreateAdvertisementWithoutAuthorization() {
	s.Run("Создаём объявление без авторизации", func() {
		response := advertisementsClient.HttpCreateAdvertisementWithParams(
			s.T(),
			"",
			"Тестовое объявление",
			"Описание тестового объявления",
			"1500",
			"1",
			testdata.PhotoPath,
			http.StatusUnauthorized,
		)

		assertions.AssertAPIError(
			s.T(),
			response,
			"unauthorized",
			"Authorization header required",
		)
	})
}

// TestCreateAdvertisementInvalidToken проверяет,
// что объявление нельзя создать с неверным токеном.
func (s *TestSuite) TestCreateAdvertisementInvalidToken() {
	s.Run("Создаём объявление с неверным токеном", func() {
		response := advertisementsClient.HttpCreateAdvertisementWithParams(
			s.T(),
			"incorrect_token",
			"Тестовое объявление",
			"Описание тестового объявления",
			"1500",
			"1",
			testdata.PhotoPath,
			http.StatusUnauthorized,
		)

		assertions.AssertAPIError(
			s.T(),
			response,
			"unauthorized",
			"Invalid or expired token",
		)
	})
}

// TestCreateAdvertisementEmptyDescription проверяет,
// что объявление нельзя создать с пустым описанием.
func (s *TestSuite) TestCreateAdvertisementEmptyDescription() {
	s.Run("Создаём объявление с пустым описанием", func() {
		response := advertisementsClient.HttpCreateAdvertisementWithParams(
			s.T(),
			s.loginData.Token,
			"Тестовое объявление",
			"",
			"1500",
			"1",
			testdata.PhotoPath,
			http.StatusBadRequest,
		)

		assertions.AssertAPIError(
			s.T(),
			response,
			"invalid_request",
			"description is required",
		)
	})
}

// TestCreateAdvertisementWithoutPhoto проверяет,
// что объявление нельзя создать без фотографии.
func (s *TestSuite) TestCreateAdvertisementWithoutPhoto() {
	s.Run("Создаём объявление без фотографии", func() {
		response := advertisementsClient.HttpCreateAdvertisementWithoutPhoto(
			s.T(),
			s.loginData.Token,
			http.StatusBadRequest,
		)

		assertions.AssertAPIError(
			s.T(),
			response,
			"invalid_request",
			"at least 1 photo is required",
		)
	})
}
