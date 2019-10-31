package main

import (
    "./config"
    "./dynamodbservice"
    "encoding/json"
    "fmt"
    "github.com/gorilla/mux"
    "log"
    "net/http"
    _ "sort"
)

/*
GetAllProducts - display all of the Products.
*/
func GetAllProducts(w http.ResponseWriter, r *http.Request) {
    p, err := dynamodbservice.Items.GetAll()
    w.Header().Add("Content-Type", "application/json")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(p)
}

/*
CreateProduct - create a new Product and add to the database.
*/
func CreateProduct(w http.ResponseWriter, r *http.Request) {
    var p dynamodbservice.Product
    w.Header().Add("Content-Type", "application/json")
    if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    defer r.Body.Close()

    if err := dynamodbservice.Items.AddProduct(p); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(p)
}

func main() {
    err := config.Settings.LoadConfigs("./config.json")
    if err != nil {
        log.Fatal("Error setting up config, make sure it's at specified route")
    }
    fmt.Println("Initializing database...")
    if initErr := dynamodbservice.Initialize(); initErr != nil {
        log.Fatal("Error Initializing database")
    }

    fmt.Println("DONE!")

    router := mux.NewRouter()
    router.HandleFunc("/products", GetAllProducts).Methods(http.MethodGet)
    router.HandleFunc("/product", CreateProduct).Methods(http.MethodPost)

    log.Fatal(http.ListenAndServe(config.Settings.Router.Port, router))
}