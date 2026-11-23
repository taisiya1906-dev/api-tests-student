package favorites

import (
	"api-tests-template/internal/constants/testdata"
	"fmt"
	"io"
	"net/http"
	"time"

	"api-tests-template/internal/assertions"
	advertisementsClient "api-tests-template/internal/client/http/advertisements"
	authClient "api-tests-template/internal/client/http/auth"
	favoritesClient "api-tests-template/internal/client/http/favorites"
	"api-tests-template/internal/managers/auth"
	favoritesManager "api-tests-template/internal/managers/favorites"

	"github.com/tidwall/gjson"
)

const nonexistentAdvertisementID = "00000000-0000-4000-8000-000000000001"

// TestFavoritesNegative проверяет ошибки добавления в избранное.
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

	s.Run("Несуществующее объявление", func() {
		response := favoritesClient.HttpAddToFavorites(
			s.T(),
			token,
			nonexistentAdvertisementID,
			http.StatusNotFound,
		)

		assertions.AssertAPIError(
			s.T(),
			response,
			"not_found",
			"Advertisement not found",
		)
	})

	s.Run("Добавление собственного объявления", func() {
		createResponse := advertisementsClient.HttpCreateAdvertisement(
			s.T(),
			token,
			testdata.PhotoPath,
		)

		s.Require().NotNil(
			createResponse,
			"Сервер не вернул ответ на создание собственного объявления",
		)

		createBody, err := io.ReadAll(createResponse.Body)
		createResponse.Body.Close()

		s.Require().NoError(
			err,
			"Не удалось прочитать ответ на создание собственного объявления",
		)

		advertisementID := gjson.GetBytes(createBody, "id").String()

		s.Require().NotEmpty(
			advertisementID,
			"В ответе на создание собственного объявления отсутствует ID. Ответ: %s",
			string(createBody),
		)

		// Удаляем тестовое объявление после проверки.
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

		assertions.AssertAPIError(
			s.T(),
			response,
			"forbidden",
			"Cannot add own advertisement to favorites",
		)
	})

	s.Run("Повторное добавление в избранное", func() {
		// Регистрируем второго пользователя.
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
			"После регистрации второго пользователя отсутствует токен",
		)

		// Основной пользователь создаёт объявление.
		createResponse := advertisementsClient.HttpCreateAdvertisement(
			s.T(),
			token,
			testdata.PhotoPath,
		)

		s.Require().NotNil(
			createResponse,
			"Сервер не вернул ответ на создание объявления для повторного добавления",
		)

		createBody, err := io.ReadAll(createResponse.Body)
		createResponse.Body.Close()

		s.Require().NoError(
			err,
			"Не удалось прочитать ответ на создание объявления",
		)

		advertisementID := gjson.GetBytes(createBody, "id").String()

		s.Require().NotEmpty(
			advertisementID,
			"В ответе на создание объявления для повторного добавления отсутствует ID. Ответ: %s",
			string(createBody),
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

		// Первое добавление должно быть успешным.
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

		assertions.AssertAPIError(
			s.T(),
			response,
			"already_exists",
			"Advertisement is already in favorites",
		)
	})

	s.Run("Добавление без авторизации", func() {
		response := favoritesClient.HttpAddToFavorites(
			s.T(),
			"",
			nonexistentAdvertisementID,
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

// TestRemoveFavoritesNegative проверяет ошибки удаления из избранного.
func (s *TestSuite) TestRemoveFavoritesNegative() {
	token := s.loginData.Token

	s.Require().NotEmpty(
		token,
		"Токен авторизации отсутствует",
	)

	s.Run("Удаляем несуществующее объявление из избранного", func() {
		response := favoritesClient.HttpRemoveFromFavorites(
			s.T(),
			token,
			nonexistentAdvertisementID,
			http.StatusNotFound,
		)

		assertions.AssertAPIError(
			s.T(),
			response,
			"not_found",
			"Favorite entry not found",
		)
	})

	s.Run("Удаляем из избранного без авторизации", func() {
		response := favoritesClient.HttpRemoveFromFavorites(
			s.T(),
			"",
			nonexistentAdvertisementID,
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

// TestGetFavoritesWithoutAuthorization проверяет,
// что GET /favorites без токена возвращает 401.
func (s *TestSuite) TestGetFavoritesWithoutAuthorization() {
	s.Run("Получаем избранное без авторизации", func() {
		response := favoritesClient.HttpGetFavorites(
			s.T(),
			"",
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
