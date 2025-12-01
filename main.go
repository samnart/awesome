package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	// "io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var anthropicClient *anthropic.Client

// Models
type SignupRequest struct {
	Name          string `json:"name"`
	CareerStage   string `json:"careerStage"`
	Goal          string `json:"goal"`
	Skills        string `json:"skills"`
	LearningStyle string `json:"learningStyle"`
	TopConcern    string `json:"topConcern"`
}

type User struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	CareerStage   string    `json:"careerStage"`
	Goal          string    `json:"goal"`
	Skills        string    `json:"skills"`
	LearningStyle string    `json:"learningStyle"`
	TopConcern    string    `json:"topConcern"`
	CreatedAt     time.Time `json:"createdAt"`
}

type Mentor struct {
	ID           string `json:"id"`
	UserID       string `json:"userId"`
	Name         string `json:"name"`
	Expertise    string `json:"expertise"`
	Personality  string `json:"personality"`
	FirstMessage string `json:"firstMessage"`
}

type MentorData struct {
	Name         string `json:"name"`
	Expertise    string `json:"expertise"`
	Personality  string `json:"personality"`
	FirstMessage string `json:"firstMessage"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	UserID   string `json:"userId"`
	MentorID string `json:"mentorId"`
	Message  string `json:"message"`
}

type ChatResponse struct {
	SessionID    string `json:"sessionId"`
	Message      string `json:"message"`
	MessageCount int    `json:"messageCount"`
}

type SessionSummaryRequest struct {
	SessionID string `json:"sessionId"`
}

type SessionSummary struct {
	Summary    string   `json:"summary"`
	NextSteps  []string `json:"nextSteps"`
	KeyTakeaway string  `json:"keyTakeaway"`
}

func init() {
	var err error
	db, err = sql.Open("sqlite3", "./mentor.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	// Create tables
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		careerStage TEXT NOT NULL,
		goal TEXT NOT NULL,
		skills TEXT NOT NULL,
		learningStyle TEXT NOT NULL,
		topConcern TEXT NOT NULL,
		createdAt DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS mentors (
		id TEXT PRIMARY KEY,
		userId TEXT NOT NULL,
		name TEXT NOT NULL,
		expertise TEXT NOT NULL,
		personality TEXT NOT NULL,
		firstMessage TEXT NOT NULL,
		createdAt DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (userId) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		userId TEXT NOT NULL,
		mentorId TEXT NOT NULL,
		messages TEXT DEFAULT '[]',
		status TEXT DEFAULT 'active',
		createdAt DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (userId) REFERENCES users(id),
		FOREIGN KEY (mentorId) REFERENCES mentors(id)
	);
	`

	if _, err := db.Exec(schema); err != nil {
		log.Fatal("Failed to create tables:", err)
	}

	// Initialize Anthropic client
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable not set")
	}

	anthropicClient = anthropic.NewClient(
		anthropic.WithAPIKey(apiKey),
	)
}

// enableCORS adds CORS headers to response
func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// handleCORS handles preflight requests
func handleCORS(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		enableCORS(w)
		w.WriteHeader(http.StatusOK)
		return
	}
}

// POST /api/signup
func handleSignup(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	handleCORS(w, r)
	if r.Method == "OPTIONS" {
		return
	}

	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	userID := uuid.New().String()

	query := `
	INSERT INTO users (id, name, careerStage, goal, skills, learningStyle, topConcern)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	if _, err := db.Exec(query, userID, req.Name, req.CareerStage, req.Goal, req.Skills, req.LearningStyle, req.TopConcern); err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"userId":  userID,
		"message": "User created successfully",
	})
}

// POST /api/get-mentor
func handleGetMentor(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	handleCORS(w, r)
	if r.Method == "OPTIONS" {
		return
	}

	var req map[string]string
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	userID := req["userId"]

	// Get user data
	var user User
	query := `SELECT id, name, careerStage, goal, skills, learningStyle, topConcern, createdAt FROM users WHERE id = ?`
	if err := db.QueryRow(query, userID).Scan(&user.ID, &user.Name, &user.CareerStage, &user.Goal, &user.Skills, &user.LearningStyle, &user.TopConcern, &user.CreatedAt); err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Generate mentor profile with Claude
	prompt := fmt.Sprintf(`You are an AI mentor matching system. Based on a student's profile, create a unique AI mentor persona.
	
Return ONLY a valid JSON object (no markdown, no extra text) with exactly these fields:
{
  "name": "mentor's name",
  "expertise": "their expertise area",
  "personality": "brief personality description",
  "firstMessage": "warm welcome message tailored to their goal"
}

Student Profile:
Name: %s
Career Stage: %s
Goal: %s
Current Skills: %s
Learning Style: %s
Top Concern: %s

Create a mentor persona that perfectly matches them.`, user.Name, user.CareerStage, user.Goal, user.Skills, user.LearningStyle, user.TopConcern)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	message, err := anthropicClient.Messages.New(ctx, &anthropic.MessageParam{
		Model:     anthropic.ModelClaude3_5Sonnet20241022,
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			{
				Role: anthropic.RoleUser,
				Content: anthropic.NewTextBlock(prompt),
			},
		},
		System: anthropic.NewTextBlock("You are a mentor matching system. Return only valid JSON."),
	})

	if err != nil {
		log.Printf("Claude API error: %v", err)
		http.Error(w, fmt.Sprintf("Claude API error: %v", err), http.StatusInternalServerError)
		return
	}

	// Extract text from response
	var responseText string
	for _, block := range message.Content {
		if textBlock, ok := block.(*anthropic.TextBlock); ok {
			responseText = textBlock.Text
			break
		}
	}

	// Parse mentor data
	var mentorData MentorData
	if err := json.Unmarshal([]byte(responseText), &mentorData); err != nil {
		log.Printf("Failed to parse mentor response: %v. Response was: %s", err, responseText)
		http.Error(w, fmt.Sprintf("Failed to parse mentor response: %v", err), http.StatusInternalServerError)
		return
	}

	// Save mentor to database
	mentorID := uuid.New().String()
	mentorQuery := `
	INSERT INTO mentors (id, userId, name, expertise, personality, firstMessage)
	VALUES (?, ?, ?, ?, ?, ?)
	`

	if _, err := db.Exec(mentorQuery, mentorID, userID, mentorData.Name, mentorData.Expertise, mentorData.Personality, mentorData.FirstMessage); err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"mentorId": mentorID,
		"mentor":   mentorData,
	})
}

// POST /api/chat
func handleChat(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	handleCORS(w, r)
	if r.Method == "OPTIONS" {
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Get mentor and user
	var mentor Mentor
	mentorQuery := `SELECT id, userId, name, expertise, personality, firstMessage FROM mentors WHERE id = ? AND userId = ?`
	if err := db.QueryRow(mentorQuery, req.MentorID, req.UserID).Scan(&mentor.ID, &mentor.UserID, &mentor.Name, &mentor.Expertise, &mentor.Personality, &mentor.FirstMessage); err != nil {
		http.Error(w, "Mentor not found", http.StatusNotFound)
		return
	}

	var user User
	userQuery := `SELECT id, name, careerStage, goal, skills, learningStyle, topConcern, createdAt FROM users WHERE id = ?`
	if err := db.QueryRow(userQuery, req.UserID).Scan(&user.ID, &user.Name, &user.CareerStage, &user.Goal, &user.Skills, &user.LearningStyle, &user.TopConcern, &user.CreatedAt); err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Get or create session
	var sessionID string
	var messagesJSON string
	sessionQuery := `SELECT id, messages FROM sessions WHERE userId = ? AND mentorId = ? AND status = 'active' LIMIT 1`
	err := db.QueryRow(sessionQuery, req.UserID, req.MentorID).Scan(&sessionID, &messagesJSON)

	var messages []ChatMessage
	if err == sql.ErrNoRows {
		// New session
		sessionID = uuid.New().String()
		messagesJSON = "[]"
	} else if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	} else {
		// Parse existing messages
		if err := json.Unmarshal([]byte(messagesJSON), &messages); err != nil {
			http.Error(w, fmt.Sprintf("Failed to parse messages: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Add user message
	messages = append(messages, ChatMessage{Role: "user", Content: req.Message})

	// Prepare messages for Claude API
	var claudeMessages []anthropic.MessageParam
	for _, msg := range messages {
		var role anthropic.MessageParamRole
		if msg.Role == "user" {
			role = anthropic.RoleUser
		} else {
			role = anthropic.RoleAssistant
		}
		claudeMessages = append(claudeMessages, anthropic.MessageParam{
			Role:    role,
			Content: anthropic.NewTextBlock(msg.Content),
		})
	}

	// Get AI response
	systemPrompt := fmt.Sprintf(`You are %s, an experienced mentor. Your expertise is in %s.
Your personality: %s

The student you're mentoring:
- Name: %s
- Career Goal: %s
- Learning Style: %s
- Current Skills: %s
- Main Concern: %s

Be warm, encouraging, and provide actionable guidance. Ask clarifying questions when needed.
Focus on helping them take the next concrete step toward their goal.`, mentor.Name, mentor.Expertise, mentor.Personality, user.Name, user.Goal, user.LearningStyle, user.Skills, user.TopConcern)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := anthropicClient.Messages.New(ctx, &anthropic.MessageParam{
		Model:     anthropic.ModelClaude3_5Sonnet20241022,
		MaxTokens: 1024,
		Messages:  claudeMessages,
		System:    anthropic.NewTextBlock(systemPrompt),
	})

	if err != nil {
		log.Printf("Claude API error: %v", err)
		http.Error(w, fmt.Sprintf("Claude API error: %v", err), http.StatusInternalServerError)
		return
	}

	// Extract response text
	var assistantMessage string
	for _, block := range response.Content {
		if textBlock, ok := block.(*anthropic.TextBlock); ok {
			assistantMessage = textBlock.Text
			break
		}
	}

	// Add assistant message
	messages = append(messages, ChatMessage{Role: "assistant", Content: assistantMessage})

	// Save or update session
	updatedMessagesJSON, _ := json.Marshal(messages)

	if err == sql.ErrNoRows {
		// Insert new session
		insertQuery := `INSERT INTO sessions (id, userId, mentorId, messages) VALUES (?, ?, ?, ?)`
		if _, err := db.Exec(insertQuery, sessionID, req.UserID, req.MentorID, string(updatedMessagesJSON)); err != nil {
			http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
			return
		}
	} else {
		// Update existing session
		updateQuery := `UPDATE sessions SET messages = ? WHERE id = ?`
		if _, err := db.Exec(updateQuery, string(updatedMessagesJSON), sessionID); err != nil {
			http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ChatResponse{
		SessionID:    sessionID,
		Message:      assistantMessage,
		MessageCount: len(messages),
	})
}

// POST /api/session-summary
func handleSessionSummary(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	handleCORS(w, r)
	if r.Method == "OPTIONS" {
		return
	}

	var req SessionSummaryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Get session
	var messagesJSON string
	query := `SELECT messages FROM sessions WHERE id = ?`
	if err := db.QueryRow(query, req.SessionID).Scan(&messagesJSON); err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	var messages []ChatMessage
	if err := json.Unmarshal([]byte(messagesJSON), &messages); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse messages: %v", err), http.StatusInternalServerError)
		return
	}

	// Build conversation text
	var conversationText strings.Builder
	for _, msg := range messages {
		conversationText.WriteString(fmt.Sprintf("%s: %s\n\n", msg.Role, msg.Content))
	}

	// Generate summary
	summaryPrompt := fmt.Sprintf(`Summarize this mentorship conversation and provide actionable next steps.

Return ONLY a valid JSON object (no markdown, no extra text) with exactly these fields:
{
  "summary": "2-3 sentence summary of what was discussed",
  "nextSteps": ["step 1", "step 2", "step 3"],
  "keyTakeaway": "one key insight from the conversation"
}

Conversation:
%s`, conversationText.String())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := anthropicClient.Messages.New(ctx, &anthropic.MessageParam{
		Model:     anthropic.ModelClaude3_5Sonnet20241022,
		MaxTokens: 512,
		Messages: []anthropic.MessageParam{
			{
				Role:    anthropic.RoleUser,
				Content: anthropic.NewTextBlock(summaryPrompt),
			},
		},
		System: anthropic.NewTextBlock("You are a learning coach. Return only valid JSON."),
	})

	if err != nil {
		log.Printf("Claude API error: %v", err)
		http.Error(w, fmt.Sprintf("Claude API error: %v", err), http.StatusInternalServerError)
		return
	}

	// Extract response text
	var responseText string
	for _, block := range response.Content {
		if textBlock, ok := block.(*anthropic.TextBlock); ok {
			responseText = textBlock.Text
			break
		}
	}

	var summary SessionSummary
	if err := json.Unmarshal([]byte(responseText), &summary); err != nil {
		log.Printf("Failed to parse summary response: %v. Response was: %s", err, responseText)
		http.Error(w, fmt.Sprintf("Failed to parse summary: %v", err), http.StatusInternalServerError)
		return
	}

	// Mark session as completed
	updateQuery := `UPDATE sessions SET status = ? WHERE id = ?`
	if _, err := db.Exec(updateQuery, "completed", req.SessionID); err != nil {
		log.Printf("Failed to update session status: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// GET /api/health
func handleHealth(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func main() {
	defer db.Close()

	// Routes
	http.HandleFunc("/api/signup", handleSignup)
	http.HandleFunc("/api/get-mentor", handleGetMentor)
	http.HandleFunc("/api/chat", handleChat)
	http.HandleFunc("/api/session-summary", handleSessionSummary)
	http.HandleFunc("/api/health", handleHealth)

	// Handle root
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Mentor MVP API"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	fmt.Printf("🚀 Mentor MVP Backend running on http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}