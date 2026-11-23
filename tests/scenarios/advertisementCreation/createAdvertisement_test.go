package advertisementCreation

import (
	"api-tests-template/internal/constants/testdata"
	"io"
	"log"
	"os"
	"testing"

	advertisementsClient "api-tests-template/internal/client/http/advertisements"
	advertisementsManager "api-tests-template/internal/managers/advertisements"
	"api-tests-template/internal/managers/auth"
	"api-tests-template/internal/managers/auth/models"

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

	base.Precondition("Авторизация пользователя перед созданием объявления")

	s.loginData = auth.Login(
		s.T(),
		os.Getenv("TEST_LOGIN"),
		os.Getenv("TEST_PASSWORD"),
	)
}

func (s *TestSuite) TearDownSuite() {
	base.TearDownSuite()
}

// TestAuthorization проверяет получение токена авторизации.
func (s *TestSuite) TestAuthorization() {
	s.Run("Проверяем успешную авторизацию пользователя", func() {
		s.Require().NotEmpty(
			s.loginData.Token,
			"Токен авторизации не должен быть пустым",
		)
	})
}

// TestCreateAdvertisementPositive проверяет создание объявления,
// получение по ID, поиск по названию и наличие фотографии.
func (s *TestSuite) TestCreateAdvertisementPositive() {
	s.Run("Создаём объявление со всеми параметрами и фотографией", func() {
		photoPath := testdata.PhotoPath

		response := advertisementsClient.HttpCreateAdvertisement(
			s.T(),
			s.loginData.Token,
			photoPath,
		)

		s.Require().NotNil(
			response,
			"Сервер не вернул ответ при создании объявления",
		)

		defer response.Body.Close()

		body, err := io.ReadAll(response.Body)
		s.Require().NoError(
			err,
			"Не удалось прочитать ответ сервера",
		)

		advertisementID := gjson.GetBytes(body, "id").String()

		s.Require().NotEmpty(
			advertisementID,
			"Сервер не вернул ID созданного объявления",
		)

		log.Printf("Объявление создано, ID: %s", advertisementID)

		// Удаляем тестовое объявление после завершения теста.
		s.T().Cleanup(func() {
			log.Printf("Удаляем тестовое объявление с ID: %s", advertisementID)

			deleteResponse := advertisementsClient.HttpDeleteAdvertisement(
				s.T(),
				s.loginData.Token,
				advertisementID,
			)

			if deleteResponse != nil {
				defer deleteResponse.Body.Close()
			}

			log.Printf("Тестовое объявление %s удалено", advertisementID)
		})

		s.Run("Получаем созданное объявление по ID", func() {
			getResponse := advertisementsClient.HttpGetAdvertisementByID(
				s.T(),
				s.loginData.Token,
				advertisementID,
			)

			s.Require().NotNil(
				getResponse,
				"Сервер не вернул ответ при получении объявления",
			)

			defer getResponse.Body.Close()

			getBody, err := io.ReadAll(getResponse.Body)
			s.Require().NoError(
				err,
				"Не удалось прочитать полученное объявление",
			)

			s.Require().Equal(
				advertisementID,
				gjson.GetBytes(getBody, "id").String(),
				"ID полученного объявления не совпадает",
			)

			s.Require().Equal(
				"Тестовое объявление",
				gjson.GetBytes(getBody, "title").String(),
				"Название объявления не совпадает",
			)

			s.Require().Equal(
				"Описание тестового объявления",
				gjson.GetBytes(getBody, "description").String(),
				"Описание объявления не совпадает",
			)

			s.Require().Equal(
				float64(1500),
				gjson.GetBytes(getBody, "price").Float(),
				"Цена объявления не совпадает",
			)

			s.Require().Equal(
				int64(1),
				gjson.GetBytes(getBody, "quantity").Int(),
				"Количество товара не совпадает",
			)
		})

		s.Run("Проверяем поиск объявления по названию", func() {
			searchResult := advertisementsManager.GetAdvertisementsList(
				s.T(),
				s.loginData.Token,
				"Тестовое объявление",
			)

			advertisementFound := false

			for _, item := range searchResult.Items {
				if item.ID == advertisementID {
					s.Require().Equal(
						"Тестовое объявление",
						item.Title,
						"Название найденного объявления не совпадает",
					)

					advertisementFound = true
					break
				}
			}

			s.Require().True(
				advertisementFound,
				"Созданное объявление не найдено в результатах поиска",
			)
		})

		s.Run("Проверяем загруженную фотографию", func() {
			photosResponse := advertisementsClient.HttpGetAdvertisementPhotos(
				s.T(),
				s.loginData.Token,
				advertisementID,
			)

			s.Require().NotNil(
				photosResponse,
				"Сервер не вернул ответ при получении фотографий",
			)

			defer photosResponse.Body.Close()

			photosBody, err := io.ReadAll(photosResponse.Body)
			s.Require().NoError(
				err,
				"Не удалось прочитать список фотографий",
			)

			photos := gjson.ParseBytes(photosBody)

			s.Require().True(
				photos.IsArray(),
				"Сервер должен вернуть массив фотографий",
			)

			s.Require().Len(
				photos.Array(),
				1,
				"У созданного объявления должна быть одна фотография",
			)

			photo := photos.Array()[0]

			s.Require().NotEmpty(
				photo.Get("id").String(),
				"У фотографии отсутствует ID",
			)

			s.Require().NotEmpty(
				photo.Get("url").String(),
				"У фотографии отсутствует URL",
			)

			s.Require().NotEmpty(
				photo.Get("created_at").String(),
				"У фотографии отсутствует дата создания",
			)

			log.Printf(
				"Фотография объявления %s успешно найдена",
				advertisementID,
			)
		})
	})
}
