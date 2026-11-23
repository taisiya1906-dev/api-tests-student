package myAdvertisement

import (
	"api-tests-template/internal/constants/testdata"
	"io"
	"log"
	"net/http"
	"os"
	"testing"

	advertisementsClient "api-tests-template/internal/client/http/advertisements"
	"api-tests-template/internal/managers/auth"
	"api-tests-template/internal/managers/auth/models"
	"api-tests-template/internal/managers/myAdvertisements"
	base "api-tests-template/tests"

	"github.com/stretchr/testify/require"
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
		"Авторизация пользователя перед проверкой собственных объявлений",
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

// TestGetMyAdvertisementsPositive проверяет,
// что созданное пользователем объявление
// присутствует в списке его собственных объявлений.
func (s *TestSuite) TestGetMyAdvertisementsPositive() {
	s.Run("Создаём и проверяем собственное объявление", func() {
		token := s.loginData.Token
		userID := s.loginData.User.Id

		require.NotEmpty(
			s.T(),
			token,
			"Токен авторизации отсутствует",
		)

		require.NotEmpty(
			s.T(),
			userID,
			"ID авторизованного пользователя отсутствует",
		)

		// Создаём собственное тестовое объявление с фотографией.
		createResponse := advertisementsClient.HttpCreateAdvertisement(
			s.T(),
			token,
			testdata.PhotoPath,
		)

		require.NotNil(
			s.T(),
			createResponse,
			"Сервер не вернул ответ на создание объявления",
		)

		createBody, err := io.ReadAll(createResponse.Body)
		createResponse.Body.Close()

		require.NoError(
			s.T(),
			err,
			"Не удалось прочитать ответ создания объявления",
		)

		advertisementID := gjson.GetBytes(
			createBody,
			"id",
		).String()

		require.NotEmpty(
			s.T(),
			advertisementID,
			"Сервер не вернул ID созданного объявления",
		)

		log.Printf(
			"Создано тестовое объявление: ID=%s",
			advertisementID,
		)

		// Регистрируем удаление тестового объявления.
		// Очистка сработает даже при ошибке проверок.
		s.T().Cleanup(func() {
			deleteResponse := advertisementsClient.HttpDeleteAdvertisement(
				s.T(),
				token,
				advertisementID,
			)

			if deleteResponse != nil {
				deleteResponse.Body.Close()
			}

			log.Printf(
				"Удалено тестовое объявление: ID=%s",
				advertisementID,
			)
		})

		// Получаем список собственных объявлений.
		advertisementsBody := myAdvertisements.GetMyAdvertisements(
			s.T(),
			token,
			http.StatusOK,
			50,
		)

		items := gjson.Get(
			advertisementsBody,
			"items",
		)

		require.True(
			s.T(),
			items.IsArray(),
			"Сервер вернул некорректный список собственных объявлений",
		)

		// Ищем только созданное нами объявление.
		var createdAdvertisement gjson.Result

		for _, item := range items.Array() {
			if item.Get("id").String() == advertisementID {
				createdAdvertisement = item
				break
			}
		}

		require.True(
			s.T(),
			createdAdvertisement.Exists(),
			"Созданное объявление %s отсутствует в списке собственных объявлений",
			advertisementID,
		)

		// Проверяем принадлежность объявления пользователю.
		require.Equal(
			s.T(),
			userID,
			createdAdvertisement.Get("user_id").String(),
			"Объявление принадлежит другому пользователю",
		)

		// Проверяем обязательные поля объявления.
		require.Equal(
			s.T(),
			advertisementID,
			createdAdvertisement.Get("id").String(),
			"ID объявления не совпадает",
		)

		require.NotEmpty(
			s.T(),
			createdAdvertisement.Get("created_at").String(),
			"Дата создания объявления отсутствует",
		)

		require.NotEmpty(
			s.T(),
			createdAdvertisement.Get("updated_at").String(),
			"Дата изменения объявления отсутствует",
		)

		require.Equal(
			s.T(),
			"Тестовое объявление",
			createdAdvertisement.Get("title").String(),
			"Название объявления не совпадает",
		)

		require.Equal(
			s.T(),
			"Описание тестового объявления",
			createdAdvertisement.Get("description").String(),
			"Описание объявления не совпадает",
		)

		require.Equal(
			s.T(),
			float64(1500),
			createdAdvertisement.Get("price").Float(),
			"Цена объявления не совпадает",
		)

		require.Equal(
			s.T(),
			int64(1),
			createdAdvertisement.Get("quantity").Int(),
			"Количество товара не совпадает",
		)

		// Проверяем наличие загруженной фотографии.
		photos := createdAdvertisement.Get("photos")

		require.True(
			s.T(),
			photos.IsArray(),
			"Поле photos должно содержать массив",
		)

		require.NotEmpty(
			s.T(),
			photos.Array(),
			"У созданного объявления отсутствует фотография",
		)

		require.NotEmpty(
			s.T(),
			photos.Array()[0].Get("url").String(),
			"У фотографии отсутствует URL",
		)

		log.Printf(
			"Собственное объявление %s успешно проверено",
			advertisementID,
		)
	})
}

// TestGetMyAdvertisementsIncorrectToken проверяет,
// что запрос с неправильным токеном возвращает 401.
func (s *TestSuite) TestGetMyAdvertisementsIncorrectToken() {
	s.Run("Получаем собственные объявления с неверным токеном", func() {
		advertisementsBody := myAdvertisements.GetMyAdvertisements(
			s.T(),
			"incorrect_token",
			http.StatusUnauthorized,
			50,
		)

		require.Equal(
			s.T(),
			"unauthorized",
			gjson.Get(advertisementsBody, "error").String(),
			"Некорректный тип ошибки авторизации",
		)

		require.Equal(
			s.T(),
			"Invalid or expired token",
			gjson.Get(advertisementsBody, "message").String(),
			"Некорректное сообщение об ошибке авторизации",
		)
	})
}
