package models

// Country represents a country in the system
type Country struct {
	Code   string `gorm:"primaryKey;size:2" json:"code"`    // ISO 3166-1 alpha-2 код
	Name   string `gorm:"not null;size:100" json:"name"`    // Название страны
	NameRu string `gorm:"not null;size:100" json:"name_ru"` // Название на русском
	Flag   string `gorm:"not null;size:10" json:"flag"`     // Emoji флага
	Region string `gorm:"not null;size:50" json:"region"`   // Регион (CIS, Europe, Americas)
}

// GetCountriesSeed returns initial countries data for seeding
func GetCountriesSeed() []Country {
	return []Country{
		// СНГ
		{"RU", "Russia", "Россия", "🇷🇺", "CIS"},
		{"UA", "Ukraine", "Україна", "🇺🇦", "CIS"},
		{"BY", "Belarus", "Беларусь", "🇧🇾", "CIS"},
		{"KZ", "Kazakhstan", "Қазақстан", "🇰🇿", "CIS"},
		{"UZ", "Uzbekistan", "Ўзбекистон", "🇺🇿", "CIS"},
		{"AM", "Armenia", "Հայաստան", "🇦🇲", "CIS"},
		{"AZ", "Azerbaijan", "Azərbaycan", "🇦🇿", "CIS"},
		{"GE", "Georgia", "საქართველო", "🇬🇪", "CIS"},
		{"KG", "Kyrgyzstan", "Кыргызстан", "🇰🇬", "CIS"},
		{"MD", "Moldova", "Moldova", "🇲🇩", "CIS"},
		{"TJ", "Tajikistan", "Тоҷикистон", "🇹🇯", "CIS"},
		{"TM", "Turkmenistan", "Türkmenistan", "🇹🇲", "CIS"},

		// Европа
		{"DE", "Germany", "Германия", "🇩🇪", "Europe"},
		{"FR", "France", "Франция", "🇫🇷", "Europe"},
		{"GB", "United Kingdom", "Великобритания", "🇬🇧", "Europe"},
		{"IT", "Italy", "Италия", "🇮🇹", "Europe"},
		{"ES", "Spain", "Испания", "🇪🇸", "Europe"},
		{"PL", "Poland", "Польша", "🇵🇱", "Europe"},
		{"NL", "Netherlands", "Нидерланды", "🇳🇱", "Europe"},
		{"SE", "Sweden", "Швеция", "🇸🇪", "Europe"},
		{"NO", "Norway", "Норвегия", "🇳🇴", "Europe"},
		{"FI", "Finland", "Финляндия", "🇫🇮", "Europe"},
		{"DK", "Denmark", "Дания", "🇩🇰", "Europe"},
		{"CH", "Switzerland", "Швейцария", "🇨🇭", "Europe"},
		{"AT", "Austria", "Австрия", "🇦🇹", "Europe"},
		{"BE", "Belgium", "Бельгия", "🇧🇪", "Europe"},
		{"PT", "Portugal", "Португалия", "🇵🇹", "Europe"},
		{"CZ", "Czech Republic", "Чехия", "🇨🇿", "Europe"},
		{"GR", "Greece", "Греция", "🇬🇷", "Europe"},
		{"HU", "Hungary", "Венгрия", "🇭🇺", "Europe"},
		{"RO", "Romania", "Румыния", "🇷🇴", "Europe"},
		{"BG", "Bulgaria", "Болгария", "🇧🇬", "Europe"},
		{"SK", "Slovakia", "Словакия", "🇸🇰", "Europe"},
		{"HR", "Croatia", "Хорватия", "🇭🇷", "Europe"},
		{"SI", "Slovenia", "Словения", "🇸🇮", "Europe"},
		{"LT", "Lithuania", "Литва", "🇱🇹", "Europe"},
		{"LV", "Latvia", "Латвия", "🇱🇻", "Europe"},
		{"EE", "Estonia", "Эстония", "🇪🇪", "Europe"},
		{"IE", "Ireland", "Ирландия", "🇮🇪", "Europe"},

		// Америка
		{"US", "United States", "США", "🇺🇸", "Americas"},
		{"CA", "Canada", "Канада", "🇨🇦", "Americas"},
		{"MX", "Mexico", "Мексика", "🇲🇽", "Americas"},
		{"BR", "Brazil", "Бразилия", "🇧🇷", "Americas"},
		{"AR", "Argentina", "Аргентина", "🇦🇷", "Americas"},

		// Азия
		{"CN", "China", "Китай", "🇨🇳", "Asia"},
		{"JP", "Japan", "Япония", "🇯🇵", "Asia"},
		{"KR", "South Korea", "Южная Корея", "🇰🇷", "Asia"},
		{"IN", "India", "Индия", "🇮🇳", "Asia"},
		{"TR", "Turkey", "Турция", "🇹🇷", "Asia"},
		{"IL", "Israel", "Израиль", "🇮🇱", "Asia"},

		// Океания
		{"AU", "Australia", "Австралия", "🇦🇺", "Oceania"},
		{"NZ", "New Zealand", "Новая Зеландия", "🇳🇿", "Oceania"},
	}
}
