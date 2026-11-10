package advertisementSearch

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"api-tests-template/internal/assertions"
	"api-tests-template/internal/managers/advertisements"
	"api-tests-template/internal/managers/auth"
	"api-tests-template/internal/managers/auth/models"
	"api-tests-template/internal/utils"
	base "api-tests-template/tests"
)

const (
	existingAdTitle = "Фигурка \"Оскар\""
	searchBeginning = "Фигурка"
	searchMiddle    = "Оскар"
	searchEnd       = "Оскар\""
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
	s.Require().NotEmpty(existingAdTitle, "задайте existingAdTitle")
	s.Require().NotEmpty(searchBeginning, "задайте searchBeginning")
	s.Require().NotEmpty(searchMiddle, "задайте searchMiddle")
	s.Require().NotEmpty(searchEnd, "задайте searchEnd")
	base.Precondition("Авторизация пользователя")
	s.loginData = auth.Login(s.T(), os.Getenv("TEST_LOGIN"), os.Getenv("TEST_PASSWORD"))
}

func (s *TestSuite) TearDownSuite() {
	base.TearDownSuite()
}

func (s *TestSuite) TestSearchByTitle() {
	list := advertisements.GetAdvertisementsList(s.T(), s.loginData.Token, searchMiddle)
	s.Require().NotEmpty(list.Items, "В ответе есть хотя бы одно объявление")
	assertions.ContainsAdvertisementByTitle(s.T(), list.Items, existingAdTitle)
}

func (s *TestSuite) TestSearchByTitle_Substring() {
	cases := []struct {
		name   string
		search string
	}{
		{"начало", searchBeginning},
		{"середина", searchMiddle},
		{"конец", searchEnd},
		{"нижний регистр", strings.ToLower(searchMiddle)},
		{"верхний регистр", strings.ToUpper(searchMiddle)},
		{"смешанный регистр", utils.MixCase(searchMiddle)},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			list := advertisements.GetAdvertisementsList(s.T(), s.loginData.Token, tc.search)
			assertions.ContainsAdvertisementByTitle(s.T(), list.Items, existingAdTitle)
		})
	}
}