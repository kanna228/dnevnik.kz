Dnevnik.kz is a web-based portal designed to assist students and teachers by providing a centralized platform for managing lessons, student accounts, and educational activities. This project was developed to simplify the organization of educational workflows while allowing users to log in and view relevant content based on their roles (Student or Teacher).

---

## Project Goal
The goal of Dnevnik.kz is to create an easy-to-use web application where:
- Students can access their class schedules and related materials.
- Teachers can manage lesson plans and student activities.
- Users can log in as either a **Student** or **Teacher** with personalized views for each role.

The project streamlines communication and content delivery between students and educators.

---

## Features
- **Role-Based Login**: Users can register or log in as a student or teacher.
- **Schedules**: Teachers can publish class schedules that are visible to students.
- **User-Friendly Interface**: Clean navigation for students and teachers.
- **Database Integration**: MongoDB is used for storing user information (students, teachers).
- **Backend API**: Developed using Go (Golang) for efficient handling of requests and data processing.

---

## Team Members
The project was developed by:
1. **Abdimanap Diaz**
2. **Abdyhalyk Diaz**
3. **Dildahan Zhandos**

---

## Screenshot
![image](https://github.com/user-attachments/assets/e6b8d9ba-4cec-496b-8bc1-bb5908b855ab)

---

## Tech Stack
The following tools and technologies were used to develop Dnevnik.kz:
- **Go (Golang)**: For backend server implementation.
- **MongoDB**: NoSQL database for storing users and schedules.
- **HTML/CSS/JavaScript**: Frontend for creating the user interface.
- **Fetch API**: For communication between the client and server.
- **Net/http Package**: For handling HTTP requests and responses in Go.

---

## How to Run the Project
Follow the steps below to set up and run the Dnevnik.kz project on your local machine:

### 1. Prerequisites
- Go (Golang) must be installed. [Download here](https://go.dev/dl/)
- MongoDB server must be running locally or remotely. [Download here](https://www.mongodb.com/try/download/community)

### 2. Clone the Repository
```bash
git clone https://github.com/kanna228/dnevnik.kz.git
cd dnevnik.kz
```

### 3. Set Up MongoDB
- Start your MongoDB server.
- Ensure the database name and connection details are correctly configured in the Go backend.

### 4. Start the Backend Server
Run the Go server:
```bash
go run main.go db.go
```
The server will start at `http://localhost:8080`.


### 5. Interact with the Application
- Access `http://localhost:8080` to test the backend API.
- Use the frontend UI to:
    - Register as a **Student** or **Teacher**.
    - Log in and view personalized content (e.g., schedules).

---

## Resources and References
- [Golang Documentation](https://go.dev/doc/)
- [MongoDB Documentation](https://www.mongodb.com/docs/)
- [HTML/CSS/JS Basics](https://developer.mozilla.org/)

---

## Contact
For any inquiries or issues regarding the project, please contact the development team:
tg @mirsjex

---

Thank you for using **Dnevnik.kz**!
