package customerrors

import "errors"

//файл для собственных ошибок

var (
	ErrEnvNotFound     = errors.New("файл .env не найден")
	ErrParamNotFound   = errors.New("один или несколько критически важных параметров не были найдены в env, проверьте: POSTGRES_HOST, POSTGRES_PORT, POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB")
	ErrInvalidEnvValue = errors.New("некорректное значение переменной окружения")
	ErrDBCreation      = errors.New("ошибка при создании БД")
	ErrDBQuery         = errors.New("ошибка во время исполнения sql запроса")
	ErrDBScan          = errors.New("ошибка при чтении результата из БД")

	ErrNotFound      = errors.New("объект не найден")
	ErrAlreadyExists = errors.New("объект уже существует")

	ErrValidation = errors.New("ошибка валидации объекта")
	ErrForbidden  = errors.New("действие запрещено")
)
