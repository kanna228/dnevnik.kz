package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"gopkg.in/gomail.v2"
)

type Transaction struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CartID        string             `bson:"cart_id" json:"cart_id"`
	Customer      Customer           `bson:"customer" json:"customer"`
	Amount        float64            `bson:"amount" json:"amount"`
	Status        string             `bson:"status" json:"status"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	PaymentMethod string             `bson:"payment_method,omitempty" json:"payment_method,omitempty"`
	CardNumber    string             `bson:"card_number,omitempty" json:"card_number,omitempty"` // ✅ Masked card number
}

type Customer struct {
	ID    primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name  string             `bson:"name" json:"name"`
	Email string             `bson:"email" json:"email"`
}

// Create a new transaction
func createTransaction(w http.ResponseWriter, r *http.Request) {
	log.Println("📥 Received request to create a transaction...")

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		log.Println("❌ Error reading request body:", err)
		return
	}

	log.Println("🔍 Raw request body:", string(body))

	// Parse JSON request
	var requestData struct {
		CartID string  `json:"cart_id"`
		Amount float64 `json:"amount"`
		UserID string  `json:"user_id"`
	}
	if err := json.Unmarshal(body, &requestData); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		log.Println("❌ JSON Decode Error:", err)
		return
	}

	// Validate User ID
	userID, err := primitive.ObjectIDFromHex(requestData.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		log.Println("❌ Invalid User ID:", err)
		return
	}

	// Fetch user details from MongoDB
	var dbUser struct {
		ID    primitive.ObjectID `bson:"_id"`
		Name  string             `bson:"name"`
		Email string             `bson:"email"`
	}

	userCollection := client.Database("your_db_name").Collection("users")
	err = userCollection.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&dbUser)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		log.Println("❌ User Not Found:", err)
		return
	}

	// Assign transaction details
	transaction := Transaction{
		ID:     primitive.NewObjectID(),
		CartID: requestData.CartID,
		Customer: Customer{
			ID:    dbUser.ID, // ✅ Properly store user ID
			Name:  dbUser.Name,
			Email: dbUser.Email,
		},
		Amount:    requestData.Amount,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	// Insert transaction into MongoDB
	collection := client.Database("your_db_name").Collection("transactions")
	result, err := collection.InsertOne(context.Background(), transaction)
	if err != nil {
		http.Error(w, "Failed to create transaction", http.StatusInternalServerError)
		log.Println("❌ MongoDB Insert Error:", err)
		return
	}

	transactionID := result.InsertedID.(primitive.ObjectID).Hex()
	log.Println("✅ Transaction created successfully with ID:", transactionID)

	// Send the transaction ID back to the frontend
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"transaction_id": transactionID})
}

// Process Payment
func processPayment(w http.ResponseWriter, r *http.Request) {

	log.Println("📥 Received payment request...")

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		log.Println("❌ Error reading request body:", err)
		return
	}

	log.Println("🔍 Raw request body:", string(body))

	// Parse JSON
	var paymentData struct {
		TransactionID string `json:"transaction_id"`
		CardNumber    string `json:"card_number"`
		Expiry        string `json:"expiry"`
		CVV           string `json:"cvv"`
	}

	if err := json.Unmarshal(body, &paymentData); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		log.Println("❌ JSON Decode Error:", err)
		return
	}

	collection := client.Database("your_db_name").Collection("transactions")
	transactionID, err := primitive.ObjectIDFromHex(paymentData.TransactionID)
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		log.Println("❌ Invalid Transaction ID:", err)
		return
	}

	// Fetch transaction details to get the logged-in user's email
	var transaction Transaction
	err = collection.FindOne(context.Background(), bson.M{"_id": transactionID}).Decode(&transaction)
	if err != nil {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		log.Println("❌ Transaction Not Found:", err)
		return
	}
	log.Println("🔍 Checking transaction status:", transaction.Status)
	if transaction.Status == "paid" {
		log.Println("❌ Transaction already paid")
		http.Error(w, "This transaction has already been paid", http.StatusBadRequest)
		return
	}
	if isTransactionPaid(transaction) {
		http.Error(w, "This transaction has already been paid", http.StatusBadRequest)
		return
	}

	maskedCard := "**** **** **** " + paymentData.CardNumber[len(paymentData.CardNumber)-4:]

	// ✅ Detect Payment Method (Simplified example)
	paymentMethod := "Visa"
	if strings.HasPrefix(paymentData.CardNumber, "4") {
		paymentMethod = "Visa"
	} else if strings.HasPrefix(paymentData.CardNumber, "5") {
		paymentMethod = "MasterCard"
	}

	// Determine payment status
	status := "paid"
	if paymentData.CardNumber == "0000 0000 0000 0000" {
		status = "declined"
	}

	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": transactionID},
		bson.M{"$set": bson.M{
			"status":         status,
			"payment_method": paymentMethod,
			"card_number":    maskedCard, // ✅ Store only masked version
		}},
	)
	if err != nil {
		http.Error(w, "Failed to update transaction", http.StatusInternalServerError)
		log.Println("❌ MongoDB Update Error:", err)
		return
	}

	// Update transaction status in MongoDB
	updateResult, err := collection.UpdateOne(
		context.Background(),
		bson.M{"_id": transactionID},
		bson.M{"$set": bson.M{"status": status}},
	)
	if err != nil {
		http.Error(w, "Failed to update transaction", http.StatusInternalServerError)
		log.Println("❌ MongoDB Update Error:", err)
		return
	}

	log.Println("✅ Transaction status updated:", updateResult.ModifiedCount)

	if status == "paid" {
		generateReceipt(paymentData.TransactionID)
		sendReceiptEmail(transaction.Customer.Email, "receipt.pdf") // ✅ Send to the correct user email
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": status})
}

// Get Payment Status for Logged-in User
func getPaymentStatus(w http.ResponseWriter, r *http.Request) {
	// Fetch user ID from the token or session
	userID := r.Header.Get("User-ID")
	log.Println("Fetching transaction for user ID:", userID)

	// Convert userID to ObjectId if it's a string
	userIDObj, err := primitive.ObjectIDFromHex(userID) // Convert userID string to ObjectId
	if err != nil {
		log.Println("Error converting userID to ObjectId:", err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Query the transactions collection to find the payment status
	collection := client.Database("your_db_name").Collection("transactions")
	var transaction Transaction
	err = collection.FindOne(context.Background(), bson.M{"customer._id": userIDObj}).Decode(&transaction)
	if err != nil {
		log.Println("Transaction not found:", err)
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	// Return payment status and other info as JSON
	response := map[string]interface{}{
		"status":         transaction.Status,
		"payment_method": transaction.PaymentMethod,
		"created_at":     transaction.CreatedAt,
		"card_number":    transaction.CardNumber,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func isTransactionPaid(transaction Transaction) bool {
	log.Println("🔍 Checking transaction status:", transaction.Status)
	if transaction.Status == "paid" {
		log.Println("❌ Transaction already paid")
		return true
	}
	return false
}

// Generate PDF Receipt
func generateReceipt(transactionID string) {
	log.Println("📄 Generating fiscal receipt for transaction:", transactionID)

	pdf := gofpdf.New("P", "mm", "A4", "")

	// ✅ Use a Unicode-compatible font for Cyrillic support
	pdf.AddUTF8Font("DejaVu", "", "DejaVuSansCondensed-Oblique.ttf")
	pdf.SetFont("DejaVu", "", 14)

	// ✅ Connect to MongoDB and retrieve transaction details
	collection := client.Database("your_db_name").Collection("transactions")
	var transaction Transaction
	objectID, err := primitive.ObjectIDFromHex(transactionID)
	if err != nil {
		log.Println("❌ Invalid transaction ID format:", err)
		return
	}

	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&transaction)
	if err != nil {
		log.Println("❌ Transaction not found in database:", err)
		return
	}

	// ✅ Set Timezone to Kazakhstan (Asia/Almaty)
	loc, err := time.LoadLocation("Asia/Almaty")
	if err != nil {
		log.Println("❌ Error loading timezone:", err)
		loc = time.FixedZone("KZ", 6*60*60) // Fallback to UTC+6
	}

	transactionDate := transaction.CreatedAt.In(loc).Format("02-01-2006 15:04:05")

	// ✅ Retrieve user details from the transaction
	customerName := transaction.Customer.Name
	customerEmail := transaction.Customer.Email
	totalAmount := transaction.Amount
	paymentMethod := transaction.PaymentMethod
	cardNumber := transaction.CardNumber

	// ✅ Mask card number (Show last 4 digits only)
	maskedCard := "**** **** **** " + cardNumber[len(cardNumber)-4:]

	// ✅ Add Header
	pdf.AddPage()
	pdf.CellFormat(0, 10, "Фискальный чек", "", 1, "C", false, 0, "")
	pdf.SetFont("DejaVu", "", 10)
	pdf.Ln(5)

	// ✅ Add Company Details
	pdf.CellFormat(0, 10, "Компания: Dnevnik.kz", "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 10, "Адрес: г. Астана, Мангилик ел ", "", 1, "L", false, 0, "")
	pdf.Ln(5)

	// ✅ Add Transaction Info
	pdf.CellFormat(0, 10, fmt.Sprintf("Номер транзакции: %s", transactionID), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 10, fmt.Sprintf("Дата (KZ): %s", transactionDate), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 10, fmt.Sprintf("ФИО клиента: %s", customerName), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 10, fmt.Sprintf("Email клиента: %s", customerEmail), "", 1, "L", false, 0, "")
	pdf.Ln(5)

	// ✅ Add Table Header
	pdf.SetFont("DejaVu", "", 10)
	pdf.CellFormat(60, 10, "Наименование", "1", 0, "C", false, 0, "")
	pdf.CellFormat(20, 10, "Кол-во", "1", 0, "C", false, 0, "")
	pdf.CellFormat(30, 10, "Цена", "1", 0, "C", false, 0, "")
	pdf.CellFormat(30, 10, "Итого", "1", 1, "C", false, 0, "")

	// ✅ Add Transaction Items (Assuming a single payment item)
	pdf.CellFormat(60, 10, "Олимпиада", "1", 0, "L", false, 0, "")
	pdf.CellFormat(20, 10, "1", "1", 0, "C", false, 0, "")
	pdf.CellFormat(30, 10, fmt.Sprintf("%.2f KZT", totalAmount), "1", 0, "C", false, 0, "")
	pdf.CellFormat(30, 10, fmt.Sprintf("%.2f KZT", totalAmount), "1", 1, "C", false, 0, "")

	// ✅ Add Total Amount
	pdf.Ln(5)
	pdf.SetFont("DejaVu", "", 12)
	pdf.CellFormat(0, 10, fmt.Sprintf("Общая сумма: %.2f KZT", totalAmount), "", 1, "R", false, 0, "")

	// ✅ Add Payment Method & Masked Card Number
	pdf.Ln(3)
	pdf.CellFormat(0, 10, fmt.Sprintf("Способ оплаты: %s", paymentMethod), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 10, fmt.Sprintf("Номер карты: %s", maskedCard), "", 1, "L", false, 0, "")

	// ✅ Save the PDF receipt
	err = pdf.OutputFileAndClose("receipt.pdf")
	if err != nil {
		log.Println("❌ Ошибка при создании чека:", err)
	} else {
		log.Println("✅ Фискальный чек успешно создан для:", customerName)
	}
}

// Send Receipt via Email
func sendReceiptEmail(to, filePath string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "dildahanz@mail.ru")
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Ваш фискальный чек")
	m.SetBody("text/plain; charset=UTF-8", "Спасибо за покупку! Ваш чек во вложении.")

	// ✅ Ensure the PDF attachment is properly encoded
	m.Attach(filePath, gomail.SetHeader(map[string][]string{
		"Content-Transfer-Encoding": {"base64"},
	}))

	d := gomail.NewDialer("smtp.mail.ru", 587, "dildahanz@mail.ru", "NmwPuFt4svU9eiDa0Bu0")
	d.LocalName = "localhost"
	err := d.DialAndSend(m)
	if err != nil {
		log.Println("❌ Email sending failed:", err)
		return err
	}
	log.Println("✅ Email sent successfully to:", to)
	return nil
}

// Serve Payment Page
func paymentPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "views/payment.html")
}

// Start Payment Service
func StartPaymentService() {

	// ✅ Serve `payment.html` when accessing `/payment`
	http.HandleFunc("/payment", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "views/payment.html")
	})

	// ✅ API Routes
	http.HandleFunc("/api/payment", corsMiddleware(processPayment))
	http.HandleFunc("/api/transaction", corsMiddleware(createTransaction))

	log.Println("🚀 Payment service running on port 8081")
	http.ListenAndServe(":8081", nil)
}
