package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
)

func main_page(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "main_page.html") // Подключаем файл HTML
}
func login_page(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "login.html") // Подключаем файл HTML
}
func teacher_login_page(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "teacher_login.html") // Подключаем файл HTML
}
func register_page(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "register.html") // Подключаем файл HTML
}

// Главный обработчик
func handler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "contact.html")
}

// Структура для обработки JSON-запроса
type RequestData struct {
	Message string `json:"message"` // Ожидаемое поле "message"
}

// Структура для JSON-ответа
type ResponseData struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// Обработчик POST-запроса
func postHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Читаем и декодируем тело запроса
	var requestData struct {
		Message string `json:"message"`
		Email   string `json:"email"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil || requestData.Message == "" || requestData.Email == "" {
		// Если JSON некорректный или поля "message"/"email" отсутствуют или пусты
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ResponseData{
			Status:  "fail",
			Message: "Некорректный JSON: обязательные поля message и email",
		})
		return
	}

	// Проверка формата email
	emailRegex := `^[^\s@]+@[^\s@]+\.[^\s@]+$`
	matched, _ := regexp.MatchString(emailRegex, requestData.Email)
	if !matched {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ResponseData{
			Status:  "fail",
			Message: "Неверный формат email",
		})
		return
	}

	// Если JSON корректный, выводим сообщение и email в консоль
	fmt.Printf("Получено сообщение: %s\nEmail: %s\n", requestData.Message, requestData.Email)

	// Отправляем успешный ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResponseData{
		Status:  "success",
		Message: "Данные успешно приняты",
	})
}

func main() {
	// Регистрируем маршруты
	fs := http.FileServer(http.Dir("./static")) // Указываем папку "static"
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", main_page)
	http.HandleFunc("/login", login_page)
	http.HandleFunc("/teacher_login", teacher_login_page)
	http.HandleFunc("/register", register_page)
	http.HandleFunc("/contact", handler)
	http.HandleFunc("/api", postHandler)

	// Запускаем сервер
	fmt.Println("Сервер запущен на http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Ошибка запуска сервера:", err)
	}
}
