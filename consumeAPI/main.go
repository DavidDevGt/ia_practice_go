package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Estructura de un coctel de la API
type Cocktail struct {
	Name         string `json:"strDrink"`
	Thumbnail    string `json:"strDrinkThumb"`
	Instructions string `json:"strInstructions"`
	ID           string `json:"idDrink"`
}

// Respuesta de la API para la lista de cocteles
type CocktailListResponse struct {
	Drinks []Cocktail `json:"drinks"`
}

// Respuesta de la API para el detalle
type CocktailDetailResponse struct {
	Drinks []Cocktail `json:"drinks"`
}

// Conseguir coctel por ingrediente
func FetchCocktails(ingredient string) ([]Cocktail, error) {
	url := fmt.Sprintf("https://www.thecocktaildb.com/api/json/v1/1/filter.php?i=%s", ingredient)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var cocktailListResp CocktailListResponse
	err = json.Unmarshal(body, &cocktailListResp)
	if err != nil {
		return nil, err
	}

	return cocktailListResp.Drinks, nil
}

// Conseguir coctel por ID
func FetchCocktailDetails(id string) (Cocktail, error) {
	var cocktail Cocktail
	url := fmt.Sprintf("https://www.thecocktaildb.com/api/json/v1/1/lookup.php?i=%s", id)
	resp, err := http.Get(url)
	if err != nil {
		return cocktail, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return cocktail, err
	}

	var cocktailDetailResp CocktailDetailResponse
	err = json.Unmarshal(body, &cocktailDetailResp)
	if err != nil {
		return cocktail, err
	}

	if len(cocktailDetailResp.Drinks) > 0 {
		cocktail = cocktailDetailResp.Drinks[0]
	}

	return cocktail, nil
}

func main() {
	ingredient := "Gin"

	cocktails, err := FetchCocktails(ingredient)
	if err != nil {
		fmt.Println("Error al buscar cocteles:", err)
		return
	}

	// Quiero usar goroutines para hacer las peticiones a la API
	// Canal para los cocteles
	detailedCocktailsChan := make(chan Cocktail)

	var wg sync.WaitGroup

	// Dividir el chance entre las goroutines
	batchSize := (len(cocktails) + 1) / 2

	for i := 0; i < 2; i++ {
		// Calcular el rango de cocteles que le toca a esta goroutine
		start := i * batchSize
		end := start + batchSize
		if end > len(cocktails) {
			end = len(cocktails)
		}

		wg.Add(1)
		go func(cocktailsBatch []Cocktail) {
			defer wg.Done()
			for _, cocktail := range cocktailsBatch {
				detailedCocktail, err := FetchCocktailDetails(cocktail.ID)
				if err != nil {
					fmt.Println("Error al obtener detalles del coctel:", err)
					continue
				}
				detailedCocktailsChan <- detailedCocktail

				// Agregar para que la API no de estado 429
				time.Sleep(200 * time.Millisecond)
			}
		}(cocktails[start:end])
	}

	// Cierro el canal cuando termine
	go func() {
		wg.Wait()
		close(detailedCocktailsChan)
	}()

	// Leer del canal para procesar los cocteles
	for cocktail := range detailedCocktailsChan {
		fmt.Printf("Nombre: %s\nThumbnail: %s\nInstrucciones: %s\n\n", cocktail.Name, cocktail.Thumbnail, cocktail.Instructions)
	}
}
