package favorites

import (
	"io"
	"net/http"
	"testing"

	advertisementsClient "api-tests-template/internal/client/http/advertisements"
	favoritesClient "api-tests-template/internal/client/http/favorites"

	"github.com/tidwall/gjson"
)

// readBody читает ответ HTTP и закрывает его тело.
func readBody(t *testing.T, response *http.Response) []byte {
	t.Helper()

	if response == nil {
		t.Fatal("Сервер не вернул HTTP-ответ")
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("Не удалось прочитать ответ сервера: %v", err)
	}

	return body
}

// GetFavoritesIDs возвращает ID объявлений из избранного.
func GetFavoritesIDs(t *testing.T, token string) []string {
	t.Helper()

	response := favoritesClient.HttpGetFavorites(t, token, http.StatusOK)
	body := readBody(t, response)

	result := gjson.ParseBytes(body)

	// Поддерживаем массив и ответ с полем items.
	items := result
	if !result.IsArray() {
		items = result.Get("items")
	}

	if !items.IsArray() {
		t.Fatalf("Некорректный список избранного: %s", string(body))
	}

	ids := make([]string, 0)

	for _, item := range items.Array() {
		id := item.Get("advertisement.id").String()

		if id == "" {
			id = item.Get("advertisement_id").String()
		}

		if id == "" {
			t.Fatalf("В записи избранного отсутствует ID объявления: %s", item.Raw)
		}

		ids = append(ids, id)
	}

	return ids
}

// FindAvailableAdvertisement ищет чужое объявление,
// которого ещё нет в избранном.
func FindAvailableAdvertisement(
	t *testing.T,
	token string,
	currentUserID string,
) string {
	t.Helper()

	if currentUserID == "" {
		t.Fatal("ID авторизованного пользователя пуст")
	}

	favoriteIDs := GetFavoritesIDs(t, token)

	excluded := make(map[string]bool)
	for _, id := range favoriteIDs {
		excluded[id] = true
	}

	response := advertisementsClient.HttpGetAdvertisementsForFavorites(
		t,
		token,
	)
	body := readBody(t, response)

	items := gjson.GetBytes(body, "items")
	if !items.IsArray() {
		t.Fatalf("Некорректный список объявлений: %s", string(body))
	}

	for _, item := range items.Array() {
		id := item.Get("id").String()
		ownerID := item.Get("user_id").String()

		if id != "" &&
			ownerID != "" &&
			ownerID != currentUserID &&
			!excluded[id] {
			return id
		}
	}

	t.Fatal("Не найдено доступное чужое объявление вне избранного")
	return ""
}

// AddFavorite добавляет объявление и проверяет ответ.
func AddFavorite(
	t *testing.T,
	token string,
	advertisementID string,
) {
	t.Helper()

	response := favoritesClient.HttpAddToFavorites(
		t,
		token,
		advertisementID,
		http.StatusCreated,
	)

	body := readBody(t, response)

	resultID := gjson.GetBytes(body, "advertisement.id").String()
	addedAt := gjson.GetBytes(body, "added_at").String()

	if resultID != advertisementID {
		t.Fatalf(
			"Неверный ID добавленного объявления: ожидали %s, получили %s",
			advertisementID,
			resultID,
		)
	}

	if addedAt == "" {
		t.Fatal("В ответе отсутствует added_at")
	}
}

// RemoveFavorite удаляет объявление из избранного.
func RemoveFavorite(
	t *testing.T,
	token string,
	advertisementID string,
) {
	t.Helper()

	response := favoritesClient.HttpRemoveFromFavorites(
		t,
		token,
		advertisementID,
		http.StatusNoContent,
	)

	_ = readBody(t, response)
}
