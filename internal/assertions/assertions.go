package assertions

import (
	"strings"
	"testing"

	"api-tests-template/internal/managers/advertisements/models"

	"github.com/stretchr/testify/require"
)

func ContainsAdvertisementByTitle(t *testing.T, items []models.Advertisement, title string) {
	t.Helper()

	for _, item := range items {
		if strings.EqualFold(item.Title, title) {
			return
		}
	}

	require.Fail(t, "Объявление с названием не найдено", "Искали: %s", title)
}
func ContainsAdvertisementById(t *testing.T, items []models.Advertisement, id string) {
	t.Helper()
	for _, item := range items {
		if item.ID == id {
			return
		}
	}
	require.Fail(t, "Объявление с ID не найдено", "Искали ID: %s", id)
}