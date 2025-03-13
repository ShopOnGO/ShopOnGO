// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "termsOfService": "http://shopongo.com/terms/",
        "contact": {
            "name": "Support Team",
            "url": "http://shopongo.com/support",
            "email": "support@shopongo.com"
        },
        "license": {
            "name": "MIT",
            "url": "https://opensource.org/licenses/MIT"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/auth/login": {
            "post": {
                "description": "Аутентифицирует пользователя по email и паролю, возвращает JWT токен",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Вход в систему",
                "parameters": [
                    {
                        "description": "Данные для входа",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/auth.LoginRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Успешный вход, возвращает JWT токен",
                        "schema": {
                            "$ref": "#/definitions/auth.LoginResponse"
                        }
                    },
                    "401": {
                        "description": "Неверные учетные данные",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Ошибка сервера при создании токена",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/auth/refresh": {
            "post": {
                "description": "Принимает refresh-токен (из cookie), проверяет его и возвращает новый JWT токен",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Обновление токенов",
                "responses": {
                    "200": {
                        "description": "Новый JWT токен",
                        "schema": {
                            "$ref": "#/definitions/auth.LoginResponse"
                        }
                    },
                    "400": {
                        "description": "Некорректный запрос",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "401": {
                        "description": "Неверный или просроченный refresh-токен",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Ошибка сервера при создании токена",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/auth/register": {
            "post": {
                "description": "Создает учетную запись пользователя и возвращает JWT токен для аутентификации",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Регистрация нового пользователя",
                "parameters": [
                    {
                        "description": "Данные для регистрации",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/auth.RegisterRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Успешная регистрация, возвращает JWT токен",
                        "schema": {
                            "$ref": "#/definitions/auth.LoginResponse"
                        }
                    },
                    "400": {
                        "description": "Некорректные данные для регистрации",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "409": {
                        "description": "Пользователь с таким email уже существует",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Ошибка сервера при создании токена",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/home": {
            "get": {
                "description": "Получает информацию о популярных товарах, категориях потом и акциях",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "home"
                ],
                "summary": "Главная страница",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/home.HomeData"
                        }
                    }
                }
            }
        },
        "/links": {
            "get": {
                "description": "Возвращает список всех коротких ссылок с возможностью пагинации",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "links"
                ],
                "summary": "Получить все ссылки",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Количество ссылок (по умолчанию 10)",
                        "name": "limit",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "Смещение (по умолчанию 0)",
                        "name": "offset",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/link.GetAllLinksResponse"
                        }
                    },
                    "400": {
                        "description": "Некорректные параметры limit/offset",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            },
            "post": {
                "description": "Генерирует короткую ссылку по переданному URL и сохраняет ее в базе",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "links"
                ],
                "summary": "Создание короткой ссылки",
                "parameters": [
                    {
                        "description": "Данные для создания ссылки",
                        "name": "link",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/link.LinkCreateRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/link.Link"
                        }
                    },
                    "400": {
                        "description": "Некорректный запрос",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/links/{id}": {
            "put": {
                "description": "Изменяет URL или хеш существующей короткой ссылки",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "links"
                ],
                "summary": "Обновление ссылки",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "ID ссылки",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Данные для обновления ссылки",
                        "name": "link",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/link.LinkUpdateRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/link.Link"
                        }
                    },
                    "400": {
                        "description": "Некорректный запрос",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "404": {
                        "description": "Ссылка не найдена",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            },
            "delete": {
                "description": "Удаляет существующую короткую ссылку из базы данных",
                "tags": [
                    "links"
                ],
                "summary": "Удаление ссылки",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "ID ссылки",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Ссылка успешно удалена",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Некорректный ID",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "404": {
                        "description": "Ссылка не найдена",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Ошибка сервера",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/stats": {
            "get": {
                "description": "Возвращает агрегированную статистику по количеству переходов, сгруппированную по дням или месяцам",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "statistics"
                ],
                "summary": "Получить статистику переходов",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Начальная дата (формат: YYYY-MM-DD)",
                        "name": "from",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Конечная дата (формат: YYYY-MM-DD)",
                        "name": "to",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Группировка (допустимые значения: day, month)",
                        "name": "by",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Успешный ответ со статистикой",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/stat.GetStatResponse"
                            }
                        }
                    },
                    "400": {
                        "description": "Некорректные параметры запроса",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/{hash}": {
            "get": {
                "description": "Ищет короткую ссылку в базе по хешу и выполняет перенаправление",
                "tags": [
                    "links"
                ],
                "summary": "Редирект по хешу",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Хеш ссылки",
                        "name": "hash",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "307": {
                        "description": "Перенаправление",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "404": {
                        "description": "Ссылка не найдена",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "auth.LoginRequest": {
            "type": "object",
            "required": [
                "email",
                "password"
            ],
            "properties": {
                "email": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                }
            }
        },
        "auth.LoginResponse": {
            "type": "object",
            "properties": {
                "token": {
                    "type": "string"
                }
            }
        },
        "auth.RegisterRequest": {
            "type": "object",
            "required": [
                "email",
                "name",
                "password"
            ],
            "properties": {
                "email": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                }
            }
        },
        "category.Category": {
            "type": "object",
            "properties": {
                "description": {
                    "type": "string"
                },
                "image_url": {
                    "description": "Ссылка на изображение категории",
                    "type": "string"
                },
                "name": {
                    "type": "string"
                }
            }
        },
        "home.HomeData": {
            "type": "object",
            "properties": {
                "categories": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/category.Category"
                    }
                },
                "featured_products": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/product.Product"
                    }
                }
            }
        },
        "link.GetAllLinksResponse": {
            "type": "object",
            "properties": {
                "count": {
                    "type": "integer"
                },
                "links": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/link.Link"
                    }
                }
            }
        },
        "link.Link": {
            "type": "object",
            "properties": {
                "hash": {
                    "type": "string"
                },
                "stats": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/stat.Stat"
                    }
                },
                "url": {
                    "type": "string"
                }
            }
        },
        "link.LinkCreateRequest": {
            "type": "object",
            "required": [
                "url"
            ],
            "properties": {
                "url": {
                    "type": "string"
                }
            }
        },
        "link.LinkUpdateRequest": {
            "type": "object",
            "required": [
                "url"
            ],
            "properties": {
                "hash": {
                    "type": "string"
                },
                "url": {
                    "type": "string"
                }
            }
        },
        "product.Product": {
            "type": "object",
            "properties": {
                "color": {
                    "type": "string"
                },
                "description": {
                    "type": "string"
                },
                "discount": {
                    "description": "CategoryID  uint   ` + "`" + `gorm:\"not null\" json:\"category_id\"` + "`" + `//foreign key\nBrandID      uint    ` + "`" + `gorm:\"not null\"  json:\"brand_id\"` + "`" + `\nPrice        float64 ` + "`" + `gorm:\"not null\"  json:\"price\"` + "`" + `",
                    "type": "number"
                },
                "gallery": {
                    "description": "JSON хранящий ссылки на изображения",
                    "type": "string"
                },
                "gender": {
                    "type": "string"
                },
                "is_available": {
                    "description": "доступен",
                    "type": "boolean"
                },
                "material": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "rating": {
                    "description": "VendorCode   string  ` + "`" + `gorm:\"type:varchar(100);unique;not null\"json:\"vendor_code\"` + "`" + `//артикул",
                    "type": "number"
                },
                "reviews_count": {
                    "description": "количество отзывов",
                    "type": "integer"
                },
                "season": {
                    "type": "string"
                },
                "size": {
                    "type": "string"
                },
                "stock": {
                    "description": "количество в наличии",
                    "type": "integer"
                },
                "video_url": {
                    "description": "ImageURL    string  ` + "`" + `gorm:\"type:varchar(255)\" json:\"image_url\"` + "`" + `",
                    "type": "string"
                }
            }
        },
        "stat.GetStatResponse": {
            "type": "object",
            "properties": {
                "period": {
                    "type": "string"
                },
                "sum": {
                    "type": "integer"
                }
            }
        },
        "stat.Stat": {
            "type": "object",
            "properties": {
                "clicks": {
                    "type": "integer"
                },
                "date": {
                    "description": "поддерживается в postgres",
                    "type": "string",
                    "format": "date"
                },
                "link_id": {
                    "type": "integer"
                }
            }
        }
    },
    "securityDefinitions": {
        "ApiKeyAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8081",
	BasePath:         "/",
	Schemes:          []string{"http"},
	Title:            "ShopOnGO API",
	Description:      "API сервиса ShopOnGO, обеспечивающего авторизацию, управление пользователями, товарами и аналитикой.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
