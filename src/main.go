package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	YANDEX_API_KEY = ${{ secrets.API_KEY }}
	YANDEX_API_URL = ${{ secrets.API_URL }}

	CODES_FILE = "../datasource/codes.json"
)

type Data map[string][]string

var stationCodes Data

func loadStationCodes(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewDecoder(file).Decode(&stationCodes)
}

func findCode(name string) (string, bool) {
	name = strings.TrimSpace(name)
	if codes, found := stationCodes[name]; found && len(codes) > 0 {
		return codes[0], true
	}
	lowerName := strings.ToLower(name)
	for key, codes := range stationCodes {
		if strings.Contains(strings.ToLower(key), lowerName) && len(codes) > 0 {
			return codes[0], true
		}
	}
	return "", false
}

func getSchedule(from, to, date string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s?apikey=%s&format=json&from=%s&to=%s&lang=ru_RU&date=%s&transfers=true",
		YANDEX_API_URL, YANDEX_API_KEY, from, to, date)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func scheduleHandler(c *gin.Context) {
	fromStation := c.Query("from_station")
	toStation := c.Query("to_station")
	date := c.Query("date")

	if fromStation == "" || toStation == "" || date == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Необходимо указать from_station, to_station и date"})
		return
	}

	if !strings.HasPrefix(fromStation, "s") {
		code, found := findCode(fromStation)
		if !found {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Не найден код для станции %s", fromStation)})
			return
		}
		fromStation = code
	}
	if !strings.HasPrefix(toStation, "s") {
		code, found := findCode(toStation)
		if !found {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Не найден код для станции %s", toStation)})
			return
		}
		toStation = code
	}

	schedule, err := getSchedule(fromStation, toStation, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, schedule)
}

func routeHandler(c *gin.Context) {
	origin := c.Query("origin")
	destination := c.Query("destination")
	date := c.Query("date")

	if origin == "" || destination == "" || date == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Необходимо указать origin, destination и date"})
		return
	}

	if !strings.HasPrefix(origin, "s") {
		code, found := findCode(origin)
		if !found {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Не найден код для станции %s", origin)})
			return
		}
		origin = code
	}
	if !strings.HasPrefix(destination, "s") {
		code, found := findCode(destination)
		if !found {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Не найден код для станции %s", destination)})
			return
		}
		destination = code
	}

	directRoute, err := getSchedule(origin, destination, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if threads, exists := directRoute["threads"]; exists && threads != nil {
		c.JSON(http.StatusOK, gin.H{"direct_route": directRoute})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Прямой маршрут не найден. Функционал построения маршрутов с пересадками в разработке."})
	}
}

func complexRouteHandler(c *gin.Context) {
	origin := c.Query("origin")
	transfer := c.Query("transfer")
	destination := c.Query("destination")
	date := c.Query("date")

	if origin == "" || transfer == "" || destination == "" || date == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Необходимо указать origin, transfer, destination и date"})
		return
	}

	leg1, err1 := getSchedule(origin, transfer, date)
	if err1 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Ошибка при получении маршрута от %s до %s: %v", origin, transfer, err1)})
		return
	}

	leg2, err2 := getSchedule(transfer, destination, date)
	if err2 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Ошибка при получении маршрута от %s до %s: %v", transfer, destination, err2)})
		return
	}

	result := map[string]interface{}{
		"leg1":         leg1,
		"leg2":         leg2,
		"requested_at": time.Now().Format(time.RFC3339),
	}

	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Printf("Ошибка маршалинга результата: %v", err)
	} else {

		timestamp := time.Now().Format("20060102T150405")
		fileName := fmt.Sprintf("complex_route_result_%s.json", timestamp)
		err = ioutil.WriteFile(fileName, resultJSON, 0644)
		if err != nil {
			log.Printf("Ошибка записи в файл %s: %v", fileName, err)
		} else {
			log.Printf("Результат сохранён в файл %s", fileName)
		}

		fmt.Println(string(resultJSON))
	}

	c.JSON(http.StatusOK, result)
}

func main() {

	if err := loadStationCodes(CODES_FILE); err != nil {
		log.Fatalf("Ошибка загрузки кодов станций: %v", err)
	}

	router := gin.Default()
	router.GET("/api/schedule", scheduleHandler)
	router.GET("/api/routes", routeHandler)
	router.GET("/api/complex_route", complexRouteHandler)

	log.Fatal(router.Run(":8080"))
}
