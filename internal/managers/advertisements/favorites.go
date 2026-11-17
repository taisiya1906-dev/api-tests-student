package advertisements

import (
	"io"
	"testing"

	advertisementsClient "api-tests-template/internal/client/http/advertisements"

	"github.com/tidwall/gjson"
)

// GetAnotherUserAdvertisementID ищет объявление,
// которое принадлежит другому пользователю.
func GetAnotherUserAdvertisementID(
	t *testing.T,
	token string,
	currentUserID string,
) string {
	t.Helper()

	response := advertisementsClient.HttpGetAdvertisementsForFavorites(
		t,
		token,
	)

	if response == nil {
		t.Fatal("Сервер не вернул список объявлений")
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("Не удалось прочитать список объявлений: %v", err)
	}

	items := gjson.GetBytes(body, "items")
	if !items.IsArray() {
		t.Fatal("Сервер вернул некорректный список объявлений")
	}

	if currentUserID == "" {
		t.Fatal("Не получен ID авторизованного пользователя")
	}

	for _, item := range items.Array() {
		advertisementID := item.Get("id").String()
		ownerID := item.Get("user_id").String()

		if advertisementID != "" &&
			ownerID != "" &&
			ownerID != currentUserID {
			return advertisementID
		}
	}

	t.Fatal("Среди полученных объявлений не найдено объявление другого пользователя")
	return ""
}
