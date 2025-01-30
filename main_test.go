package main

import (
	"testing"
	"time"

	"net/http"
	"net/http/httptest"

	"github.com/tebeka/selenium"
)

// Unit Test: Проверка генерации токена
func TestGenerateToken(t *testing.T) {
	token, err := generateToken()
	if err != nil {
		t.Fatalf("Token generation failed: %v", err)
	}
	if len(token) != 32 {
		t.Errorf("Incorrect token length. Expected: 32, Got: %d", len(token))
	}
}

// End-to-End Test: Авторизация пользователя
func TestUserLogin(t *testing.T) {
	// Убедись, что Selenium сервер работает на порту 4444
	caps := selenium.Capabilities{"browserName": "chrome"}
	wd, err := selenium.NewRemote(caps, "http://localhost:4444/wd/hub")
	if err != nil {
		t.Fatalf("Failed to start session: %v", err)
	}
	defer wd.Quit()

	// Открытие страницы логина
	err = wd.Get("https://04ff-85-159-27-200.ngrok-free.app/login")
	if err != nil {
		t.Fatalf("Failed to load login page: %v", err)
	}

	// Ожидание появления поля ввода email
	err = wd.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
		_, err := wd.FindElement(selenium.ByID, "email")
		return err == nil, nil
	}, 10*time.Second)
	if err != nil {
		t.Fatalf("Email input not found: %v", err)
	}

	// Заполнение поля email
	emailInput, err := wd.FindElement(selenium.ByID, "email")
	if err != nil {
		t.Fatalf("Failed to find email input: %v", err)
	}
	emailInput.SendKeys("amigo553@mail.ru")

	// Ожидание появления поля ввода пароля
	err = wd.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
		_, err := wd.FindElement(selenium.ByID, "password")
		return err == nil, nil
	}, 10*time.Second)
	if err != nil {
		t.Fatalf("Password input not found: %v", err)
	}

	// Заполнение поля пароля
	passwordInput, err := wd.FindElement(selenium.ByID, "password")
	if err != nil {
		t.Fatalf("Failed to find password input: %v", err)
	}
	passwordInput.SendKeys("hehe")

	// Ожидание появления кнопки логина
	err = wd.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
		_, err := wd.FindElement(selenium.ByID, "loginForm")
		return err == nil, nil
	}, 10*time.Second)
	if err != nil {
		t.Fatalf("Login button not found: %v", err)
	}

	// Нажатие на кнопку логина
	loginButton, err := wd.FindElement(selenium.ByID, "loginForm")
	if err != nil {
		t.Fatalf("Failed to find login button: %v", err)
	}
	loginButton.Click()

	// Проверка успешного входа
	err = wd.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
		// Проверяем, что пользователь попал на главную страницу после логина
		url, err := wd.CurrentURL()
		return err == nil && url == "https://04ff-85-159-27-200.ngrok-free.app/dashboard", nil
	}, 10*time.Second)
	if err != nil {
		t.Fatalf("Login failed or did not redirect to dashboard: %v", err)
	}

}
func TestTestEmailHandler(t *testing.T) {
	// Create a new HTTP request
	req, err := http.NewRequest("GET", "/test-email", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a ResponseRecorder to capture the response
	rr := httptest.NewRecorder()

	// Create a handler for the test
	handler := http.HandlerFunc(testEmailHandler)

	// Serve the HTTP request using the handler
	handler.ServeHTTP(rr, req)

	// Check if the response code is 200 (OK)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status 200, but got %v", status)
	}

	// Check if the response body is correct
	expected := "Test email sent successfully!"
	if rr.Body.String() != expected {
		t.Errorf("Expected response body to be %v, but got %v", expected, rr.Body.String())
	}
}
