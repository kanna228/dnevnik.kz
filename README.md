# Dnevnik.kz

Dnevnik.kz is a comprehensive web portal that simplifies the educational process by uniting students and teachers on a single platform. The application allows for secure user registration, role-based access, real-time communication, and several integrated services designed to enhance both teaching and learning experiences.

---

## Project Goals

- **For Students:**
  - Access schedules and educational materials.
  - Communicate in real time with teachers using WebSocket.

- **For Teachers:**
  - Manage lesson plans and perform CRUD operations on user data.
  - Access an exclusive admin dashboard to send emails (with file attachments) and handle advanced tasks.

- **General Features:**
  - User registration requires email confirmation.
  - Secure login via JSON Web Token (JWT) authentication.
  - Purchase Olympiad tickets through an integrated microservice available on the main screen.

---

## Features

- **User Management (CRUD):**  
  Create, read, update, and delete operations for both student and teacher profiles. All user data is stored in the database, which connects automatically on the server.

- **Email Confirmation:**  
  New users must confirm their email address during registration.

- **JWT Authentication:**  
  Secure login system using JSON Web Tokens to protect user sessions.

- **Role-Based Access Control:**  
  Users register as either a **Student** or **Teacher**. Teachers gain access to an exclusive admin dashboard.

- **Admin Dashboard (Teachers Only):**  
  Teachers can send emails (with file attachments) and manage additional administrative tasks.

- **Olympiad Ticket Microservice:**  
  Purchase Olympiad tickets directly from the main screen.

- **Real-Time Communication:**  
  Communicate instantly via WebSocket, enabling direct messaging between students and teachers.

---

## Tech Stack

- **Go (Golang):** Backend server and API implementation.
- **MongoDB:** NoSQL database (automatically connected on the server).
- **HTML/CSS/JavaScript:** Frontend development for a user-friendly interface.
- **JWT:** Secure user authentication.
- **WebSocket:** Real-time communication between users.
- **Fetch API:** Client-server interaction.
- **Email Service:** For email confirmations and sending messages with attachments.
- **Microservice Architecture:** Manages the purchase of Olympiad tickets.

---

## Installation and Running

1. **Clone the Repository:**
   ```bash
   git clone https://github.com/kanna228/dnevnik.kz.git
   cd dnevnik.kz
1. **Run the Application: Execute the following command to start the server:**
   go run .

**Team and Contact**
    Project Team:

    Abdimanap Diaz
    Abdyhalyk Diaz
    Dildahan Zhandos

**For any inquiries or support, please contact us via Telegram: @mirsjex**

**Thank you for choosing Dnevnik.kz!**