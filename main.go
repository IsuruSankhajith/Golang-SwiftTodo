package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// Todo represents a single task with a title, completion status, and creation time.
type Todo struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
}

// TodoList is a struct that manages a list of todos and a mutex for thread-safe operations.
type TodoList struct {
	todos     []Todo
	idCounter int
	mu        sync.Mutex
	changed   bool // Flag to track if any changes have been made
}

// CreateTodo adds a new todo to the list.
func (t *TodoList) CreateTodo(title string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.idCounter++
	newTodo := Todo{
		ID:        t.idCounter,
		Title:     title,
		Completed: false,
		CreatedAt: time.Now(),
	}
	t.todos = append(t.todos, newTodo)
	t.changed = true
	fmt.Println("To-Do added successfully.")
}

// ListTodos prints all todos in the list.
func (t *TodoList) ListTodos() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.todos) == 0 {
		fmt.Println("No To-Dos found.")
		return
	}
	fmt.Println("\nTo-Do List:")
	for _, todo := range t.todos {
		status := "Incomplete"
		if todo.Completed {
			status = "Completed"
		}
		fmt.Printf("ID: %d | Title: %s | Status: %s | Created At: %s\n", todo.ID, todo.Title, status, todo.CreatedAt.Format(time.RFC822))
	}
}

// UpdateTodo allows updating a todo's title and completion status.
func (t *TodoList) UpdateTodo(id int, newTitle string, completed bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for i, todo := range t.todos {
		if todo.ID == id {
			if newTitle != "" {
				t.todos[i].Title = newTitle
			}
			t.todos[i].Completed = completed
			t.changed = true
			fmt.Println("To-Do updated successfully.")
			return
		}
	}
	fmt.Println("To-Do not found.")
}

// DeleteTodo removes a todo from the list by ID.
func (t *TodoList) DeleteTodo(id int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for i, todo := range t.todos {
		if todo.ID == id {
			t.todos = append(t.todos[:i], t.todos[i+1:]...)
			t.changed = true
			fmt.Println("To-Do deleted successfully.")
			return
		}
	}
	fmt.Println("To-Do not found.")
}

// SaveToFile saves the todos to a file in JSON format.
func (t *TodoList) SaveToFile(filename string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(t.todos)
	if err != nil {
		return err
	}
	fmt.Println("To-Do list saved to file.")
	t.changed = false // Reset the changed flag after saving
	return nil
}

// LoadFromFile loads todos from a file.
func (t *TodoList) LoadFromFile(filename string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&t.todos)
	if err != nil {
		return err
	}
	fmt.Println("To-Do list loaded from file.")
	return nil
}

// AutoSave periodically saves the todos to a file if there are changes.
func (t *TodoList) AutoSave(filename string, interval time.Duration, done chan bool) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Only save if there are changes
			t.mu.Lock()
			shouldSave := t.changed
			t.mu.Unlock()

			if shouldSave {
				err := t.SaveToFile(filename)
				if err != nil {
					fmt.Println("Error saving file:", err)
				}
			}
		case <-done:
			fmt.Println("Auto-save stopped.")
			return
		}
	}
}

func main() {
	todoList := &TodoList{}
	filename := "todos.json"

	// Load from file at the start
	err := todoList.LoadFromFile(filename)
	if err != nil && !os.IsNotExist(err) {
		fmt.Println("Error loading file:", err)
	}

	// Start auto-saving in a separate goroutine
	done := make(chan bool)
	go todoList.AutoSave(filename, 10*time.Second, done)

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Enhanced To-Do Application with Auto-Save")
	fmt.Println("----------------------------")

	for {
		// Display menu
		fmt.Println("\nMenu:")
		fmt.Println("1. Create To-Do")
		fmt.Println("2. List To-Dos")
		fmt.Println("3. Update To-Do")
		fmt.Println("4. Delete To-Do")
		fmt.Println("5. Exit")
		fmt.Print("Enter your choice: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			fmt.Print("Enter the title of the new to-do: ")
			title, _ := reader.ReadString('\n')
			title = strings.TrimSpace(title)
			if title == "" {
				fmt.Println("Title cannot be empty.")
			} else {
				todoList.CreateTodo(title)
			}
		case "2":
			todoList.ListTodos()
		case "3":
			fmt.Print("Enter the ID of the to-do to update: ")
			idStr, _ := reader.ReadString('\n')
			idStr = strings.TrimSpace(idStr)
			var id int
			_, err := fmt.Sscan(idStr, &id)
			if err != nil {
				fmt.Println("Invalid ID. Please enter a numeric value.")
				continue
			}

			fmt.Print("Enter new title (leave empty to keep the current title): ")
			newTitle, _ := reader.ReadString('\n')
			newTitle = strings.TrimSpace(newTitle)

			fmt.Print("Mark as completed? (yes/no): ")
			completedStr, _ := reader.ReadString('\n')
			completedStr = strings.TrimSpace(completedStr)
			completed := strings.ToLower(completedStr) == "yes"

			todoList.UpdateTodo(id, newTitle, completed)
		case "4":
			fmt.Print("Enter the ID of the to-do to delete: ")
			idStr, _ := reader.ReadString('\n')
			idStr = strings.TrimSpace(idStr)
			var id int
			_, err := fmt.Sscan(idStr, &id)
			if err != nil {
				fmt.Println("Invalid ID. Please enter a numeric value.")
				continue
			}
			todoList.DeleteTodo(id)
		case "5":
			fmt.Println("Exiting...")
			done <- true // Signal the goroutine to stop auto-saving
			return
		default:
			fmt.Println("Invalid choice. Please try again.")
		}
	}
}
