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
	orders []Order
}

type Order struct {
	ID          int
	Name        string
	ServiceDesc string
	MinioURL    string
}

func NewRepository() (*Repository, error) {
	orders := []Order{
		{
			ID:          1,
			Name:        "Выделение архаизмов и неологизмов",
			ServiceDesc: "Поиск слов текста, которые имеют ранние (архаизмы) или поздние (неологизмы) интервалы дат в историческом словаре",
			MinioURL:    minioBaseURL + "/img.png",
		},
		{
			ID:          2,
			Name:        "Расчет временного интервала",
			ServiceDesc: "Использование интервалов дат всех слов текста для вычисления вероятного периода написания: нижняя граница = самое раннее появление архаизмов, верхняя граница = самое позднее появление неологизмов",
			MinioURL:    minioBaseURL + "/img_1.png",
		},
		{
			ID:          3,
			Name:        "Хронологическая карта слов",
			ServiceDesc: "Составление списка всех слов с указанием их интервалов дат. Визуализация, какие слова старые, какие новые, и где их сосредоточение в тексте",
			MinioURL:    minioBaseURL + "/img_2.png",
		},
		{
			ID:          4,
			Name:        "Сравнительный анализ текстов",
			ServiceDesc: "Сравнение текстов по распределению интервалов дат слов, выявление относительной древности или новизны",
			MinioURL:    minioBaseURL + "/img_3.png",
		},
		{
			ID:          5,
			Name:        "Подбор слов для стилистики",
			ServiceDesc: "На основе интервалов дат слов формируется подборка слов, соответствующих конкретной эпохе, исключая анахронизмы",
			MinioURL:    minioBaseURL + "/img_4.png",
		},
		{
			ID:          6,
			Name:        "Автоматизированный анализ",
			ServiceDesc: "Программное решение для сканирования текста, проверки слов по историческим словарям с интервалами дат и определения вероятного периода написания",
			MinioURL:    minioBaseURL + "/img_5.png",
		},
		{
			ID:          7,
			Name:        "Визуализация структуры",
			ServiceDesc: "Графики или диаграммы распределения слов текста по их интервалам дат, отображение архаизмов и неологизмов",
			MinioURL:    minioBaseURL + "/img_6.png",
		},
	}

	return &Repository{
		orders: orders,
	}, nil
}

func (r *Repository) GetOrders() ([]Order, error) {
	if len(r.orders) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}
	return r.orders, nil
}

func (r *Repository) GetOrderByID(id string) (*Order, error) {
	orderID, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("неверный формат ID: %v", err)
	}

	for _, order := range r.orders {
		if order.ID == orderID {
			return &order, nil
		}
	}

	return nil, fmt.Errorf("заказ с ID %s не найден", id)
}

func (r *Repository) SearchOrders(query string) ([]Order, error) {
	if query == "" {
		return r.orders, nil
	}

	query = strings.ToLower(query)
	var filteredOrders []Order

	for _, order := range r.orders {
		if strings.Contains(strings.ToLower(order.Name), query) {
			filteredOrders = append(filteredOrders, order)
		}
	}

	return filteredOrders, nil
}
