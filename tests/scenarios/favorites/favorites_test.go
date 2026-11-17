package favorites

import (
	"fmt"
	"io"
	"log"
	"os"
	"testing"
	"time"

	advertisementsClient "api-tests-template/internal/client/http/advertisements"
	authClient "api-tests-template/internal/client/http/auth"
	advertisementsManager "api-tests-template/internal/managers/advertisements"
	"api-tests-template/internal/managers/auth"
	"api-tests-template/internal/managers/auth/models"
	favoritesManager "api-tests-template/internal/managers/favorites"
	base "api-tests-template/tests"

	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
)

type TestSuite struct {
	suite.Suite
	loginData models.LoginOkResponse
}

func TestSuiteRun(t *testing.T) {
	suite.Run(t, &TestSuite{})
}

func (s *TestSuite) SetupSuite() {
	base.SetupSuite()

	base.Precondition(
		"Авторизация пользователя перед тестированием избранного",
	)

	s.loginData = auth.Login(
		s.T(),
		os.Getenv("TEST_LOGIN"),
		os.Getenv("TEST_PASSWORD"),
	)
}

func (s *TestSuite) TearDownSuite() {
	base.TearDownSuite()
}

// TestFindAnotherUserAdvertisement проверяет поиск
// объявления другого пользователя.
func (s *TestSuite) TestFindAnotherUserAdvertisement() {
	s.Run("Находим объявление другого пользователя", func() {
		s.Require().NotEmpty(
			s.loginData.Token,
			"Токен авторизации отсутствует",
		)

		s.Require().NotEmpty(
			s.loginData.User.Id,
			"ID авторизованного пользователя отсутствует",
		)

		advertisementID := advertisementsManager.GetAnotherUserAdvertisementID(
			s.T(),
			s.loginData.Token,
			s.loginData.User.Id,
		)

		s.Require().NotEmpty(
			advertisementID,
			"Не удалось найти объявление другого пользователя",
		)

		log.Printf(
			"Найдено чужое объявление: ID=%s",
			advertisementID,
		)
	})
}

// TestFavoritesPositive проверяет полный цикл:
// добавление в избранное, получение списка и удаление.

func (s *TestSuite) TestFavoritesPositive() {
	s.Run("Добавляем, проверяем и удаляем избранное", func() {
		ownerToken := s.loginData.Token
		s.Require().NotEmpty(ownerToken, "Токен владельца отсутствует")

		// Создаём второго пользователя.
		secondUser := auth.RegisterTestUser(
			s.T(),
			authClient.RegisterRequest{
				BirthDate: "2000-01-01T00:00:00Z",
				Email: fmt.Sprintf(
					"avito.favorite.%d@example.com",
					time.Now().UnixNano(),
				),
				FirstName: "Test",
				LastName:  "Favorite",
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
			ownerToken,
			"../../../testdata/test_photo.png",
		)

		s.Require().NotNil(createResponse)

		createBody, err := io.ReadAll(createResponse.Body)
		createResponse.Body.Close()
		s.Require().NoError(err)

		advertisementID := gjson.GetBytes(createBody, "id").String()
		s.Require().NotEmpty(advertisementID)

		// Объявление удаляем после завершения теста.
		s.T().Cleanup(func() {
			response := advertisementsClient.HttpDeleteAdvertisement(
				s.T(),
				ownerToken,
				advertisementID,
			)
			if response != nil {
				response.Body.Close()
			}
		})

		// Второй пользователь добавляет чужое объявление.
		favoritesManager.AddFavorite(
			s.T(),
			secondUser.Token,
			advertisementID,
		)

		// Регистрируем очистку избранного на случай ошибки.
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

		// Проверяем добавление.
		ids := favoritesManager.GetFavoritesIDs(
			s.T(),
			secondUser.Token,
		)

		s.Require().Contains(
			ids,
			advertisementID,
			"Объявление не появилось в избранном",
		)

		// Удаляем из избранного.
		favoritesManager.RemoveFavorite(
			s.T(),
			secondUser.Token,
			advertisementID,
		)

		// Проверяем удаление.
		idsAfterDelete := favoritesManager.GetFavoritesIDs(
			s.T(),
			secondUser.Token,
		)

		s.Require().NotContains(
			idsAfterDelete,
			advertisementID,
			"Объявление осталось в избранном",
		)

		log.Printf(
			"Полный цикл избранного выполнен: %s",
			advertisementID,
		)
	})
}

// TestDeletedAdvertisementRemovedFromFavorites проверяет,
// что удалённое владельцем объявление исчезает
// из избранного другого пользователя.
func (s *TestSuite) TestDeletedAdvertisementRemovedFromFavorites() {
	s.Run("Удаление объявления очищает избранное", func() {
		ownerToken := s.loginData.Token

		s.Require().NotEmpty(
			ownerToken,
			"Токен владельца отсутствует",
		)

		// Регистрируем второго пользователя с уникальной почтой.
		email := fmt.Sprintf(
			"avito.qa.%d@example.com",
			time.Now().UnixNano(),
		)

		secondUser := auth.RegisterTestUser(
			s.T(),
			authClient.RegisterRequest{
				BirthDate: "2000-01-01T00:00:00Z",
				Email:     email,
				FirstName: "Test",
				LastName:  "Favorite",
				Password:  "TestPassword123!",
			},
		)

		s.Require().NotEmpty(
			secondUser.Token,
			"Токен второго пользователя отсутствует",
		)

		s.Require().NotEmpty(
			secondUser.User.Id,
			"ID второго пользователя отсутствует",
		)

		log.Printf(
			"Второй тестовый пользователь зарегистрирован",
		)

		// Основной пользователь создаёт объявление.
		createResponse := advertisementsClient.HttpCreateAdvertisement(
			s.T(),
			ownerToken,
			"../../../testdata/test_photo.png",
		)

		s.Require().NotNil(
			createResponse,
			"Сервер не вернул ответ на создание объявления",
		)

		createBody, err := io.ReadAll(createResponse.Body)
		createResponse.Body.Close()

		s.Require().NoError(
			err,
			"Не удалось прочитать ответ создания объявления",
		)

		advertisementID := gjson.GetBytes(
			createBody,
			"id",
		).String()

		s.Require().NotEmpty(
			advertisementID,
			"В ответе отсутствует ID объявления",
		)

		deleted := false

		// Если тест прервётся, удаляем созданное объявление.
		s.T().Cleanup(func() {
			if deleted {
				return
			}

			response := advertisementsClient.HttpDeleteAdvertisement(
				s.T(),
				ownerToken,
				advertisementID,
			)

			if response != nil {
				response.Body.Close()
			}
		})

		log.Printf(
			"Владелец создал объявление: ID=%s",
			advertisementID,
		)

		// Второй пользователь добавляет объявление в избранное.
		favoritesManager.AddFavorite(
			s.T(),
			secondUser.Token,
			advertisementID,
		)

		// Проверяем, что объявление появилось в избранном.
		ids := favoritesManager.GetFavoritesIDs(
			s.T(),
			secondUser.Token,
		)

		s.Require().Contains(
			ids,
			advertisementID,
			"Объявление отсутствует в избранном второго пользователя",
		)

		log.Printf(
			"Объявление %s добавлено в избранное второго пользователя",
			advertisementID,
		)

		// Владелец удаляет своё объявление.
		deleteResponse := advertisementsClient.HttpDeleteAdvertisement(
			s.T(),
			ownerToken,
			advertisementID,
		)

		s.Require().NotNil(
			deleteResponse,
			"Сервер не вернул ответ на удаление объявления",
		)

		deleteResponse.Body.Close()
		deleted = true

		log.Printf(
			"Владелец удалил объявление: ID=%s",
			advertisementID,
		)

		// Проверяем, что объявление исчезло из избранного.
		idsAfterDelete := favoritesManager.GetFavoritesIDs(
			s.T(),
			secondUser.Token,
		)

		s.Require().NotContains(
			idsAfterDelete,
			advertisementID,
			"Удалённое объявление осталось в избранном",
		)

		log.Printf(
			"Удалённое объявление %s отсутствует в избранном",
			advertisementID,
		)
	})
}
