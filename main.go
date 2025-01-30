package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	Code  string
	Price uint
}

var db *gorm.DB

func init() {
	database, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db = database

	if err := db.AutoMigrate(&Product{}); err != nil {
		panic(err)
	}

	var count int64
	if err := db.Model(&Product{}).Count(&count).Error; err != nil {
		panic(err)
	}

	if count != 0 {
		return
	}

	db.Create(&Product{Code: "D42", Price: 100})
	db.Create(&Product{Code: "E33", Price: 200})
	db.Create(&Product{Code: "A10", Price: 300})
}

// Run arbitrary SQL command using the following pattern:
// curl "http://localhost:8080/?code=$(printf "str'; DELETE FROM `products` --" | jq -sRr @uri)"
func handleGet(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	var product Product

	// Avoid using manual string interpolation; instead, rely on the driver's string interpolation,
	// which should sanitize the SQL string as part of the concatenation process.
	if err := db.Debug().Where(fmt.Sprintf("code = '%v'", queryParams.Get("code"))).First(&product).Error; err != nil {
		//   Problematic line: ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
		http.Error(w, "Failed to retrieve products", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

func main() {
	http.HandleFunc("GET /", handleGet)
	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
