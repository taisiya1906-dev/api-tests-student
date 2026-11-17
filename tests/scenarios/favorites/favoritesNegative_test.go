package favorites

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	advertisementsClient "api-tests-template/internal/client/http/advertisements"
	authClient "api-tests-template/internal/client/http/auth"
	favoritesClient "api-tests-template/internal/client/http/favorites"
	"api-tests-template/internal/managers/auth"
	favoritesManager "api-tests-template/internal/managers/favorites"

	"github.com/tidwall/gjson"
)

// TestFavoritesNegative проверяет негативные сценарии
// добавления объявлений в избранное.
func (s *TestSuite) TestFavoritesNegative() {
	token := s.loginData.Token

	s.Require().NotEmpty(
		token,
		"Токен авторизации отсутствует",
	)

	s.Require().NotEmpty(
		s.loginData.User.Id,
		"ID авторизованного пользователя отсутствует",
	)

	// Добавление несуществующего объявления.
	s.Run("Несуществующее объявление", func() {
		response := favoritesClient.HttpAddToFavorites(
			s.T(),
			token,
			"00000000-0000-4000-8000-000000000001",
			http.StatusNotFound,
		)

		s.Require().NotNil(
			response,
			"Сервер не вернул ответ",
		)
		defer response.Body.Close()

		body, err := io.ReadAll(response.Body)
		s.Require().NoError(err)

		s.Require().Equal(
			"not_found",
			gjson.GetBytes(body, "error").String(),
			"Неверный тип ошибки",
		)

		s.Require().Equal(
			"Advertisement not found",
			gjson.GetBytes(body, "message").String(),
			"Неверное сообщение об отсутствии объявления",
		)

		log.Printf(
			"Несуществующее объявление: %s",
			string(body),
		)
	})

	// Попытка добавить собственное объявление.
	s.Run("Добавление собственного объявления", func() {
		createResponse := advertisementsClient.HttpCreateAdvertisement(
			s.T(),
			token,
			"../../../testdata/test_photo.png",
		)

		s.Require().NotNil(
			createResponse,
			"Сервер не вернул ответ создания объявления",
		)

		createBody, err := io.ReadAll(createResponse.Body)
		createResponse.Body.Close()
		s.Require().NoError(err)

		advertisementID := gjson.GetBytes(
			createBody,
			"id",
		).String()

		s.Require().NotEmpty(
			advertisementID,
			"Не получен ID созданного объявления",
		)

		// Очистка созданного объявления.
		s.T().Cleanup(func() {
			deleteResponse := advertisementsClient.HttpDeleteAdvertisement(
				s.T(),
				token,
				advertisementID,
			)

			if deleteResponse != nil {
				deleteResponse.Body.Close()
			}
		})

		response := favoritesClient.HttpAddToFavorites(
			s.T(),
			token,
			advertisementID,
			http.StatusForbidden,
		)

		s.Require().NotNil(response)
		defer response.Body.Close()

		body, err := io.ReadAll(response.Body)
		s.Require().NoError(err)

		s.Require().Equal(
			"forbidden",
			gjson.GetBytes(body, "error").String(),
			"Неверный тип ошибки",
		)

		s.Require().Equal(
			"Cannot add own advertisement to favorites",
			gjson.GetBytes(body, "message").String(),
			"Неверное сообщение о добавлении собственного объявления",
		)

		log.Printf(
			"Собственное объявление: %s",
			string(body),
		)
	})

	// Попытка повторно добавить объявление в избранное.
	s.Run("Повторное добавление в избранное", func() {
		// Регистрируем второго тестового пользователя.
		secondUser := auth.RegisterTestUser(
			s.T(),
			authClient.RegisterRequest{
				BirthDate: "2000-01-01T00:00:00Z",
				Email: fmt.Sprintf(
					"avito.duplicate.%d@example.com",
					time.Now().UnixNano(),
				),
				FirstName: "Test",
				LastName:  "Duplicate",
				Password:  "TestPassword123!",
			},
		)

		s.Require().NotEmpty(
			secondUser.Token,
			"Токен второго пользователя отсутствует",
		)

		// Основной пользователь создаёт объявление.
		createResponse := advertisementsClient.HttpCreateAdvertisement(
			s.T(),
			token,
			"../../../testdata/test_photo.png",
		)

		s.Require().NotNil(
			createResponse,
			"Сервер не вернул ответ создания объявления",
		)

		createBody, err := io.ReadAll(createResponse.Body)
		createResponse.Body.Close()
		s.Require().NoError(err)

		advertisementID := gjson.GetBytes(
			createBody,
			"id",
		).String()

		s.Require().NotEmpty(
			advertisementID,
			"Не получен ID созданного объявления",
		)

		// Регистрируем удаление объявления.
		s.T().Cleanup(func() {
			deleteResponse := advertisementsClient.HttpDeleteAdvertisement(
				s.T(),
				token,
				advertisementID,
			)

			if deleteResponse != nil {
				deleteResponse.Body.Close()
			}
		})

		// Первый раз добавляем объявление в избранное.
		favoritesManager.AddFavorite(
			s.T(),
			secondUser.Token,
			advertisementID,
		)

		// Регистрируем очистку избранного.
		s.T().Cleanup(func() {
			ids := favoritesManager.GetFavoritesIDs(
				s.T(),
				secondUser.Token,
			)

			for _, id := range ids {
				if id == advertisementID {
					favoritesManager.RemoveFavorite(
						s.T(),
						secondUser.Token,
						advertisementID,
					)
					break
				}
			}
		})

		// Повторное добавление должно вернуть 409.
		response := favoritesClient.HttpAddToFavorites(
			s.T(),
			secondUser.Token,
			advertisementID,
			http.StatusConflict,
		)

		s.Require().NotNil(response)
		defer response.Body.Close()

		body, err := io.ReadAll(response.Body)
		s.Require().NoError(err)

		s.Require().Equal(
			"already_exists",
			gjson.GetBytes(body, "error").String(),
			"Неверный тип ошибки повторного добавления",
		)

		s.Require().Equal(
			"Advertisement is already in favorites",
			gjson.GetBytes(body, "message").String(),
			"Неверное сообщение о повторном добавлении",
		)

		log.Printf(
			"Повторное добавление: %s",
			string(body),
		)
	})

	// Добавление в избранное без авторизации.
	s.Run("Добавление без авторизации", func() {
		response := favoritesClient.HttpAddToFavorites(
			s.T(),
			"",
			"00000000-0000-4000-8000-000000000001",
			http.StatusUnauthorized,
		)

		s.Require().NotNil(response)
		defer response.Body.Close()

		body, err := io.ReadAll(response.Body)
		s.Require().NoError(err)

		s.Require().Equal(
			"unauthorized",
			gjson.GetBytes(body, "error").String(),
			"Неверный тип ошибки авторизации",
		)

		s.Require().Equal(
			"Authorization header required",
			gjson.GetBytes(body, "message").String(),
			"Неверное сообщение об отсутствии авторизации",
		)

		log.Printf(
			"Без авторизации: %s",
			string(body),
		)
	})
}

// TestRemoveFavoritesNegative проверяет негативные сценарии
// удаления объявлений из избранного.
func (s *TestSuite) TestRemoveFavoritesNegative() {
	token := s.loginData.Token

	s.Require().NotEmpty(
		token,
		"Токен авторизации отсутствует",
	)

	// Удаление несуществующего объявления.
	s.Run("Удаляем несуществующее объявление из избранного", func() {
		response := favoritesClient.HttpRemoveFromFavorites(
			s.T(),
			token,
			"00000000-0000-4000-8000-000000000001",
			http.StatusNotFound,
		)

		s.Require().NotNil(response)
		defer response.Body.Close()

		body, err := io.ReadAll(response.Body)
		s.Require().NoError(err)

		s.Require().Equal(
			"not_found",
			gjson.GetBytes(body, "error").String(),
			"Неверный тип ошибки при удалении",
		)

		s.Require().Equal(
			"Favorite entry not found",
			gjson.GetBytes(body, "message").String(),
			"Неверное сообщение при удалении несуществующей записи",
		)

		log.Printf(
			"Удаление несуществующего объявления: %s",
			string(body),
		)
	})

	// Удаление из избранного без авторизации.
	s.Run("Удаляем из избранного без авторизации", func() {
		response := favoritesClient.HttpRemoveFromFavorites(
			s.T(),
			"",
			"00000000-0000-4000-8000-000000000001",
			http.StatusUnauthorized,
		)

		s.Require().NotNil(response)
		defer response.Body.Close()

		body, err := io.ReadAll(response.Body)
		s.Require().NoError(err)

		s.Require().Equal(
			"unauthorized",
			gjson.GetBytes(body, "error").String(),
			"Неверный тип ошибки авторизации",
		)

		s.Require().Equal(
			"Authorization header required",
			gjson.GetBytes(body, "message").String(),
			"Неверное сообщение об отсутствии авторизации",
		)

		log.Printf(
			"Удаление без авторизации: %s",
			string(body),
		)
	})
}
