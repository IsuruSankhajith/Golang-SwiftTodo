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
	if (len(t.todos)) == 0 {
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
func (t *TodoList) AutoSave(filename string, interval time.Duration, done chan bool, wg *sync.WaitGroup) {
	defer wg.Done()
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

	// WaitGroup to wait for auto-save to finish
	var wg sync.WaitGroup

	// Start auto-saving in a separate goroutine
	done := make(chan bool)
	wg.Add(1)
	go todoList.AutoSave(filename, 10*time.Second, done, &wg)

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Welcome to the Enhanced To-Do Application with Auto-Save!")
	fmt.Println("----------------------------------------------------------")

	for {
		fmt.Println("\n🔷 MAIN MENU")
		fmt.Println("1️⃣  ➡  Create a New To-Do")
		fmt.Println("2️⃣  ➡  View All To-Dos")
		fmt.Println("3️⃣  ➡  Update an Existing To-Do")
		fmt.Println("4️⃣  ➡  Delete a To-Do")
		fmt.Println("5️⃣  ➡  Exit")
		fmt.Print("Please enter your choice (1-5): ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			fmt.Println("\n📝 CREATE A NEW TO-DO")
			fmt.Print("Enter the title of the new to-do: ")
			title, _ := reader.ReadString('\n')
			title = strings.TrimSpace(title)
			if title == "" {
				fmt.Println("⚠️ Title cannot be empty. Please try again.")
			} else {
				todoList.CreateTodo(title)
				fmt.Println("✅ To-Do created successfully!")
			}
		case "2":
			fmt.Println("\n📋 VIEW ALL TO-DOS")
			todoList.ListTodos()
		case "3":
			fmt.Println("\n✏️ UPDATE A TO-DO")
			fmt.Print("Enter the ID of the to-do to update: ")
			idStr, _ := reader.ReadString('\n')
			idStr = strings.TrimSpace(idStr)
			var id int
			_, err := fmt.Sscan(idStr, &id)
			if err != nil {
				fmt.Println("⚠️ Invalid ID. Please enter a numeric value.")
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
			fmt.Println("✅ To-Do updated successfully!")
		case "4":
			fmt.Println("\n🗑️ DELETE A TO-DO")
			fmt.Print("Enter the ID of the to-do to delete: ")
			idStr, _ := reader.ReadString('\n')
			idStr = strings.TrimSpace(idStr)
			var id int
			_, err := fmt.Sscan(idStr, &id)
			if err != nil {
				fmt.Println("⚠️ Invalid ID. Please enter a numeric value.")
				continue
			}
			todoList.DeleteTodo(id)
			fmt.Println("✅ To-Do deleted successfully!")
		case "5":
			fmt.Println("\n👋 Exiting the application... Goodbye!")
			done <- true // Signal the goroutine to stop auto-saving
			close(done)  // Close the channel
			wg.Wait()    // Wait for the auto-save goroutine to finish
			return
		default:
			fmt.Println("⚠️ Invalid choice. Please enter a valid option (1-5).")
		}
	}
}
