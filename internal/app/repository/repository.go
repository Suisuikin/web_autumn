package repository

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	minioBaseURL = "http://localhost:9000/chrono"
)

type Repository struct {
	orders []ChronoIntervals
}

type ChronoIntervals struct {
	ID          int
	Name        string
	ServiceDesc string
	MinioURL    string
	FromYear    int
	ToYear      int
}

func NewRepository() (*Repository, error) {
	chronoIntervals := []ChronoIntervals{
		{
			ID:          1,
			Name:        "Древнерусский слой",
			ServiceDesc: "Церковнославянская и летописная лексика: «вещати», «чудо», «чадо», «рать». Без заимствований.",
			MinioURL:    minioBaseURL + "/img.png",
			FromYear:    1000,
			ToYear:      1450,
		},
		{
			ID:          2,
			Name:        "Раннесреднерусский слой",
			ServiceDesc: "Смешение церковнославянской и народной речи. Первые полонизмы и кальки с латинизмов.",
			MinioURL:    minioBaseURL + "/img_1.png",
			FromYear:    1450,
			ToYear:      1600,
		},
		{
			ID:          3,
			Name:        "Позднесреднерусский слой",
			ServiceDesc: "Расширение бытовой лексики, редкие заимствования из Европы. Переходный период перед реформами Петра.",
			MinioURL:    minioBaseURL + "/img_2.png",
			FromYear:    1600,
			ToYear:      1720,
		},
		{
			ID:          4,
			Name:        "Петровский слой",
			ServiceDesc: "Активное заимствование из западных языков, формирование современного литературного языка.",
			MinioURL:    minioBaseURL + "/img_3.png",
			FromYear:    1720,
			ToYear:      1800,
		},
		{
			ID:          5,
			Name:        "Классический слой",
			ServiceDesc: "Эпоха Пушкина и Толстого. Развитие науки, формирование норм, частичная архаизация старых слов.",
			MinioURL:    minioBaseURL + "/img_4.png",
			FromYear:    1800,
			ToYear:      1917,
		},
		{
			ID:          6,
			Name:        "Революционно-советский слой",
			ServiceDesc: "Массовые неологизмы и идеологическая лексика: «колхоз», «пятилетка», «социализм».",
			MinioURL:    minioBaseURL + "/img_5.png",
			FromYear:    1917,
			ToYear:      1950,
		},
		{
			ID:          7,
			Name:        "Позднесоветский слой",
			ServiceDesc: "Техническая и бюрократическая речь: «автоматизация», «НИИ», «профком», «космодром».",
			MinioURL:    minioBaseURL + "/img_6.png",
			FromYear:    1950,
			ToYear:      1985,
		},
	}

	return &Repository{orders: chronoIntervals}, nil
}

func (r *Repository) GetOrders() ([]ChronoIntervals, error) {
	if len(r.orders) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}
	return r.orders, nil
}

func (r *Repository) GetOrderByID(id string) (*ChronoIntervals, error) {
	orderID, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("неверный формат ID: %v", err)
	}

	for _, order := range r.orders {
		if order.ID == orderID {
			return &order, nil
		}
	}

	return nil, fmt.Errorf("слой с ID %s не найден", id)
}

func (r *Repository) SearchOrders(query string) ([]ChronoIntervals, error) {
	if query == "" {
		return r.orders, nil
	}

	query = strings.ToLower(query)
	var filtered []ChronoIntervals

	for _, order := range r.orders {
		if strings.Contains(strings.ToLower(order.Name), query) {
			filtered = append(filtered, order)
		}
	}

	return filtered, nil
}
