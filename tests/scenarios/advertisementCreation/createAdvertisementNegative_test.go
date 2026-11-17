package advertisementCreation

import (
	"io"
	"log"
	"net/http"
	"strings"

	advertisementsClient "api-tests-template/internal/client/http/advertisements"

	"github.com/tidwall/gjson"
)

// TestCreateAdvertisementEmptyTitle проверяет,
// что объявление нельзя создать с пустым названием.
func (s *TestSuite) TestCreateAdvertisementEmptyTitle() {
	s.Run("Создаём объявление с пустым названием", func() {
		photoPath := "../../../testdata/test_photo.png"

		response := advertisementsClient.HttpCreateAdvertisementWithParams(
			s.T(),
			s.loginData.Token,
			"",
			"Описание тестового объявления",
			"1500",
			"1",
			photoPath,
			http.StatusBadRequest,
		)

		s.Require().NotNil(
			response,
			"Сервер не вернул ответ",
		)

		defer response.Body.Close()

		body, err := io.ReadAll(response.Body)
		s.Require().NoError(
			err,
			"Не удалось прочитать ответ сервера",
		)

		errorValue := gjson.GetBytes(body, "error").String()
		message := gjson.GetBytes(body, "message").String()

		s.Require().Equal(
			"invalid_request",
			errorValue,
			"Сервер вернул неправильный тип ошибки",
		)

		s.Require().Equal(
			"title is required",
			message,
			"Сервер вернул неправильное сообщение об ошибке",
		)

		log.Printf(
			"Ошибка при пустом названии: error=%s, message=%s",
			errorValue,
			message,
		)
	})
}

// TestCreateAdvertisementNegativePrice проверяет,
// что объявление нельзя создать с отрицательной ценой.
func (s *TestSuite) TestCreateAdvertisementNegativePrice() {
	s.Run("Создаём объявление с отрицательной ценой", func() {
		photoPath := "../../../testdata/test_photo.png"

		response := advertisementsClient.HttpCreateAdvertisementWithParams(
			s.T(),
			s.loginData.Token,
			"Тестовое объявление",
			"Описание тестового объявления",
			"-1500",
			"1",
			photoPath,
			http.StatusBadRequest,
		)

		s.Require().NotNil(
			response,
			"Сервер не вернул ответ",
		)
		defer response.Body.Close()

		body, err := io.ReadAll(response.Body)
		s.Require().NoError(
			err,
			"Не удалось прочитать ответ сервера",
		)

		errorValue := gjson.GetBytes(body, "error").String()
		message := gjson.GetBytes(body, "message").String()

		s.Require().Equal(
			"invalid_request",
			errorValue,
			"Сервер вернул неправильный тип ошибки",
		)

		s.Require().Equal(
			"price must be between 0 and 1000000 inclusive",
			message,
			"Сервер вернул неправильное сообщение об ошибке цены",
		)

		log.Printf(
			"Ошибка при отрицательной цене: error=%s, message=%s",
			errorValue,
			message,
		)
	})
}

// TestCreateAdvertisementLongTitle проверяет,
// что сервер отклоняет слишком длинное название объявления.
func (s *TestSuite) TestCreateAdvertisementLongTitle() {
	s.Run("Создаём объявление со слишком длинным названием", func() {
		longTitle := strings.Repeat("А", 1000)
		photoPath := "../../../testdata/test_photo.png"

		response := advertisementsClient.HttpCreateAdvertisementWithParams(
			s.T(),
			s.loginData.Token,
			longTitle,
			"Описание тестового объявления",
			"1500",
			"1",
			photoPath,
			http.StatusBadRequest,
		)

		s.Require().NotNil(
			response,
			"Сервер не вернул ответ",
		)
		defer response.Body.Close()

		body, err := io.ReadAll(response.Body)
		s.Require().NoError(
			err,
			"Не удалось прочитать ответ сервера",
		)

		errorValue := gjson.GetBytes(body, "error").String()
		message := gjson.GetBytes(body, "message").String()

		s.Require().Equal(
			"invalid_request",
			errorValue,
			"Сервер вернул неправильный тип ошибки",
		)

		s.Require().Equal(
			"title must not exceed 50 characters provided 1000",
			message,
			"Сервер вернул неправильное сообщение об ограничении длины названия",
		)

		log.Printf(
			"Ошибка при длинном названии: error=%s, message=%s",
			errorValue,
			message,
		)
	})
}

// TestCreateAdvertisementWithoutAuthorization проверяет,
// что объявление нельзя создать без токена авторизации.
func (s *TestSuite) TestCreateAdvertisementWithoutAuthorization() {
	s.Run("Создаём объявление без авторизации", func() {
		photoPath := "../../../testdata/test_photo.png"

		response := advertisementsClient.HttpCreateAdvertisementWithParams(
			s.T(),
			"",
			"Тестовое объявление",
			"Описание тестового объявления",
			"1500",
			"1",
			photoPath,
			http.StatusUnauthorized,
		)

		s.Require().NotNil(
			response,
			"Сервер не вернул ответ",
		)
		defer response.Body.Close()

		body, err := io.ReadAll(response.Body)
		s.Require().NoError(
			err,
			"Не удалось прочитать ответ сервера",
		)

		errorValue := gjson.GetBytes(body, "error").String()
		message := gjson.GetBytes(body, "message").String()

		s.Require().Equal(
			"unauthorized",
			errorValue,
			"Сервер вернул неправильный тип ошибки авторизации",
		)

		s.Require().Equal(
			"Authorization header required",
			message,
			"Сервер вернул неправильное сообщение об отсутствии токена",
		)

		log.Printf(
			"Ошибка без авторизации: error=%s, message=%s",
			errorValue,
			message,
		)
	})
}

// TestCreateAdvertisementInvalidToken проверяет,
// что создание объявления с неверным токеном запрещено.
func (s *TestSuite) TestCreateAdvertisementInvalidToken() {
	s.Run("Создаём объявление с неверным токеном", func() {
		photoPath := "../../../testdata/test_photo.png"

		response := advertisementsClient.HttpCreateAdvertisementWithParams(
			s.T(),
			"incorrect_token",
			"Тестовое объявление",
			"Описание тестового объявления",
			"1500",
			"1",
			photoPath,
			http.StatusUnauthorized,
		)

		s.Require().NotNil(
			response,
			"Сервер не вернул ответ",
		)
		defer response.Body.Close()

		body, err := io.ReadAll(response.Body)
		s.Require().NoError(
			err,
			"Не удалось прочитать ответ сервера",
		)

		errorValue := gjson.GetBytes(body, "error").String()
		message := gjson.GetBytes(body, "message").String()

		s.Require().Equal(
			"unauthorized",
			errorValue,
			"Неверный тип ошибки авторизации",
		)

		s.Require().Equal(
			"Invalid or expired token",
			message,
			"Сервер вернул неправильное сообщение о неверном токене",
		)

		log.Printf(
			"Ошибка при неверном токене: error=%s, message=%s",
			errorValue,
			message,
		)
	})
}

// TestCreateAdvertisementEmptyDescription проверяет,
// что объявление нельзя создать с пустым описанием.
func (s *TestSuite) TestCreateAdvertisementEmptyDescription() {
	s.Run("Создаём объявление с пустым описанием", func() {
		photoPath := "../../../testdata/test_photo.png"

		response := advertisementsClient.HttpCreateAdvertisementWithParams(
			s.T(),
			s.loginData.Token,
			"Тестовое объявление",
			"",
			"1500",
			"1",
			photoPath,
			http.StatusBadRequest,
		)

		s.Require().NotNil(
			response,
			"Сервер не вернул ответ",
		)
		defer response.Body.Close()

		body, err := io.ReadAll(response.Body)
		s.Require().NoError(
			err,
			"Не удалось прочитать ответ сервера",
		)

		errorValue := gjson.GetBytes(body, "error").String()
		message := gjson.GetBytes(body, "message").String()

		s.Require().Equal(
			"invalid_request",
			errorValue,
			"Сервер вернул неправильный тип ошибки",
		)

		s.Require().Equal(
			"description is required",
			message,
			"Сервер вернул неправильное сообщение об отсутствии описания",
		)

		log.Printf(
			"Ошибка при пустом описании: error=%s, message=%s",
			errorValue,
			message,
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

		s.Require().NotNil(
			response,
			"Сервер не вернул ответ",
		)
		defer response.Body.Close()

		body, err := io.ReadAll(response.Body)
		s.Require().NoError(
			err,
			"Не удалось прочитать ответ сервера",
		)

		errorValue := gjson.GetBytes(body, "error").String()
		message := gjson.GetBytes(body, "message").String()

		s.Require().Equal(
			"invalid_request",
			errorValue,
			"Сервер вернул неправильный тип ошибки",
		)

		s.Require().Equal(
			"at least 1 photo is required",
			message,
			"Сервер вернул неправильное сообщение об отсутствии фотографии",
		)

		log.Printf(
			"Ошибка при отсутствии фотографии: error=%s, message=%s",
			errorValue,
			message,
		)
	})
}
