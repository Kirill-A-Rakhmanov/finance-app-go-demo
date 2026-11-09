package auth

import (
	"finance-app-go-demo/internal/domain/enum"
	"finance-app-go-demo/internal/domain/model"

	"github.com/google/uuid"
)

type seedCategory struct {
	Name          string
	Slug          string
	Icon          string
	Color         string
	Type          enum.TransactionCategoryType
	Subcategories []seedCategory
}

var defaultCategories = []seedCategory{
	{
		Name:  "Автомобиль",
		Slug:  "car",
		Icon:  "car",
		Color: "#22C55E",
		Type:  enum.Expense,
		Subcategories: []seedCategory{
			{
				Name:  "Страховка",
				Slug:  "insurance",
				Icon:  "insurance",
				Color: "#22C55E",
				Type:  enum.Expense,
			},
			{
				Name:  "Бензин",
				Slug:  "gas",
				Icon:  "gas",
				Color: "#A16207",
				Type:  enum.Expense,
			},
			{
				Name:  "Обслуживание",
				Slug:  "service",
				Icon:  "service",
				Color: "#EAB308",
				Type:  enum.Expense,
			},
			{
				Name:  "Парковка",
				Slug:  "parking",
				Icon:  "parking",
				Color: "#EAB308",
				Type:  enum.Expense,
			},
		},
	},

	{
		Name:  "Продукты",
		Slug:  "food",
		Icon:  "shopping-cart",
		Color: "#22C55E",
		Type:  enum.Expense,
	},

	{
		Name:  "Подарки",
		Slug:  "gifts-out",
		Icon:  "gift",
		Color: "#22C55E",
		Type:  enum.Expense,
	},

	{
		Name:  "Транспорт",
		Slug:  "transport",
		Icon:  "car",
		Color: "#3B82F6",
		Type:  enum.Expense,
		Subcategories: []seedCategory{
			{
				Name:  "Такси",
				Slug:  "taxi",
				Icon:  "car-front",
				Color: "#FACC15",
				Type:  enum.Expense,
			},
			{
				Name:  "Общественный транспорт",
				Slug:  "public-transport",
				Icon:  "bus",
				Color: "#06B6D4",
				Type:  enum.Expense,
			},
			{
				Name:  "Каршеринг",
				Slug:  "carsharing",
				Icon:  "car",
				Color: "#06B6D4",
				Type:  enum.Expense,
			},
		},
	},

	{
		Name:  "Счета",
		Slug:  "bills",
		Icon:  "home",
		Color: "#8B5CF6",
		Type:  enum.Expense,
		Subcategories: []seedCategory{
			{
				Name:  "Аренда",
				Slug:  "rent",
				Icon:  "building",
				Color: "#8B5CF6",
				Type:  enum.Expense,
			},
			{
				Name:  "Коммунальные услуги",
				Slug:  "utilities",
				Icon:  "zap",
				Color: "#F59E0B",
				Type:  enum.Expense,
			},
			{
				Name:  "Интернет и телефон",
				Slug:  "internet-phone",
				Icon:  "wifi",
				Color: "#0EA5E9",
				Type:  enum.Expense,
			},
		},
	},

	{
		Name:  "Развлечения",
		Slug:  "entertainment",
		Icon:  "film",
		Color: "#EC4899",
		Type:  enum.Expense,
	},

	{
		Name:  "Подписки",
		Slug:  "subscriptions",
		Icon:  "repeat",
		Color: "#14B8A6",
		Type:  enum.Expense,
	},

	{
		Name:  "Покупки",
		Slug:  "shopping",
		Icon:  "shopping-bag",
		Color: "#F43F5E",
		Type:  enum.Expense,
	},

	{
		Name:  "Красота и здоровье",
		Slug:  "health",
		Icon:  "heart-pulse",
		Color: "#EF4444",
		Type:  enum.Expense,
	},

	{
		Name:  "Путешествия",
		Slug:  "travel",
		Icon:  "plane",
		Color: "#06B6D4",
		Type:  enum.Expense,
	},

	{
		Name:  "Образование",
		Slug:  "education",
		Icon:  "graduation-cap",
		Color: "#6366F1",
		Type:  enum.Expense,
	},

	{
		Name:  "Налоги",
		Slug:  "taxes",
		Icon:  "receipt",
		Color: "#EF4444",
		Type:  enum.Expense,
	},

	{
		Name:  "Переводы",
		Slug:  "transfers-out",
		Icon:  "arrow-down",
		Color: "#EC4899",
		Type:  enum.Expense,
	},

	{
		Name:  "Зарплата",
		Slug:  "salary",
		Icon:  "banknote",
		Color: "#22C55E",
		Type:  enum.Income,
	},

	{
		Name:  "Фриланс",
		Slug:  "freelance",
		Icon:  "briefcase",
		Color: "#3B82F6",
		Type:  enum.Income,
	},

	{
		Name:  "Инвестиции",
		Slug:  "investments",
		Icon:  "chart-column",
		Color: "#10B981",
		Type:  enum.Income,
	},

	{
		Name:  "Кэшбек",
		Slug:  "cashback",
		Icon:  "coins",
		Color: "#EAB308",
		Type:  enum.Income,
	},

	{
		Name:  "Подарки",
		Slug:  "gifts-in",
		Icon:  "gift",
		Color: "#EC4899",
		Type:  enum.Income,
	},

	{
		Name:  "Переводы",
		Slug:  "transfers-in",
		Icon:  "arrow-down",
		Color: "#EC4899",
		Type:  enum.Income,
	},
}

func getDefaultCategories(userID uuid.UUID) []model.TransactionCategory {
	return buildCategories(
		userID,
		nil,
		defaultCategories,
	)
}

func buildCategories(
	userID uuid.UUID,
	parentID *uuid.UUID,
	seeds []seedCategory,
) []model.TransactionCategory {
	var result []model.TransactionCategory

	for _, seed := range seeds {
		category := buildCategory(
			userID,
			parentID,
			seed,
		)

		result = append(result, category)

		if len(seed.Subcategories) > 0 {
			children := buildCategories(
				userID,
				&category.ID,
				seed.Subcategories,
			)

			result = append(result, children...)
		}
	}

	return result
}

func buildCategory(
	userID uuid.UUID,
	parentID *uuid.UUID,
	seed seedCategory,
) model.TransactionCategory {
	id := uuid.New()

	return model.TransactionCategory{
		ID:        id,
		UserID:    userID,
		ParentID:  parentID,
		Name:      seed.Name,
		Slug:      &seed.Slug,
		Icon:      seed.Icon,
		Color:     seed.Color,
		IsDefault: true,
		Type:      seed.Type,
	}
}
