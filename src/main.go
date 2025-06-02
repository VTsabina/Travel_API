/*
=======================================================================
‖                            Travel_API                               ‖
‖                     CUP-IT-2025 Changellenge                        ‖
‖                          Кейс: Axenix                               ‖
‖                       Команда: Headliners                           ‖
‖													  		          ‖
‖																	  ‖
‖		 Авторы:			 										  ‖
‖		 - Цабина Валерия (https://github.com/VTsabina)			      ‖
‖		 - Зайцев Данила (https://github.com/ZaitsevD12)			  ‖
‖		 - Постовалова Лада (https://github.com/Lada12345)		      ‖
‖		 - Миронович Игорь (https://github.com/steelemi987)		      ‖
‖																	  ‖
‖		             © 2025 Все права защищены.						  ‖
=======================================================================
*/

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

// Константы для ключей и путей
const (
	YANDEX_API_KEY = ${ secrets.API_KEY } // API ключ для внешнего сервиса
	YANDEX_API_URL = ${ secrets.API_URL } // URL API

	CODES_FILE = "../datasource/codes.json" // Путь к файлу с кодами станций
)

// Data — тип для хранения соответствий названий станций и их кодов.
// Ключ — название станции, значение — список кодов.
type Data map[string][]string

// stationCodes — глобальная переменная для хранения загруженных данных о станциях.
var stationCodes Data

/**
 * loadStationCodes Загружает данные о станциях из файла JSON.
 *
 * Эта функция открывает указанный файл, декодирует его содержимое в переменную stationCodes.
 *
 * Args:
 *     filename (string): путь к файлу JSON с данными о станциях.
 *
 * Returns:
 *     error: ошибка при открытии или декодировании файла; nil при успешной загрузке.
 */
func loadStationCodes(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err // Возвращает ошибку, если файл не удалось открыть
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(&stationCodes) // Декодирует JSON в карту stationCodes
}

/**
 * findCode Ищет код станции по её названию.
 *
 * Производит поиск точного совпадения. Если не найдено, ищет частичное совпадение по подстроке.
 *
 * Args:
 *     name (string): название станции для поиска.
 *
 * Returns:
 *     (string, bool): найденный код станции и true, если найден; пустая строка и false, если не найдено.
 */
func findCode(name string) (string, bool) {
	name = strings.TrimSpace(name)
	if codes, found := stationCodes[name]; found && len(codes) > 0 {
		return codes[0], true // Возвращает первый код из списка
	}
	lowerName := strings.ToLower(name)
	for key, codes := range stationCodes {
		if strings.Contains(strings.ToLower(key), lowerName) && len(codes) > 0 {
			return codes[0], true // Частичное совпадение по названию
		}
	}
	return "", false // Не найдено подходящего кода
}

/**
 * getSchedule Формирует запрос к внешнему API для получения расписания или маршрута между станциями.
 *
 * Отправляет GET-запрос с параметрами API и парсит ответ в карту данных.
 *
 * Args:
 *     from (string): код начальной станции.
 *     to (string): код конечной станции.
 *     date (string): дата в формате YYYY-MM-DD.
 *
 * Returns:
 *     (map[string]interface{}, error): распарсенные данные ответа API или ошибка при запросе/парсинге.
 */
func getSchedule(from, to, date string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s?apikey=%s&format=json&from=%s&to=%s&lang=ru_RU&date=%s&transfers=true",
		YANDEX_API_URL, YANDEX_API_KEY, from, to, date)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err // Ошибка при выполнении HTTP-запроса
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err // Ошибка при чтении тела ответа
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err // Ошибка при парсинге JSON-ответа
	}
	return data, nil
}

/**
 * scheduleHandler Обрабатывает HTTP-запросы на получение расписания между двумя станциями.
 *
 * Извлекает параметры из запроса: from_station, to_station и date. Обрабатывает названия станций,
 * ищет их коды при необходимости. Вызывает getSchedule и возвращает результат клиенту в формате JSON.
 *
 * Контекст:
 *     c (*gin.Context): объект контекста Gin для обработки запроса и формирования ответа.
 */
func scheduleHandler(c *gin.Context) {
	fromStation := c.Query("from_station")
	toStation := c.Query("to_station")
	date := c.Query("date")

	if fromStation == "" || toStation == "" || date == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Необходимо указать from_station, to_station и date"})
        return
    }

    // Обработка названий станций: если не начинаются с 's', ищем их коды по названию
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

    scheduleData, err := getSchedule(fromStation, toStation, date)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return // Обработка ошибок вызова внешнего API
    }

	c.JSON(http.StatusOK, scheduleData) // Возвращает полученные данные клиенту в формате JSON
}

/**
 * routeHandler Обрабатывает запросы на поиск прямого маршрута между двумя станциями.
 *
 * Аналогично scheduleHandler: извлекает параметры из запроса,
 * ищет коды станций по названиям при необходимости,
 * вызывает getSchedule и проверяет наличие ключа 'threads' для определения наличия прямого маршрута.
 *
 * Контекст:
 *     c (*gin.Context): объект контекста Gin для обработки запроса и формирования ответа.
 */
func routeHandler(c *gin.Context) {
    origin := c.Query("origin")
    destination := c.Query("destination")
    date := c.Query("date")

    if origin == "" || destination == "" || date == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Необходимо указать origin , destination и date"})
        return
    }

    if !strings.HasPrefix(origin,"s"){
        code ,found:=findCode(origin)
        if !found{
            c.JSON(http.StatusBadRequest ,gin.H{"error":fmt.Sprintf("Не найден код для станции %s",origin)})
            return 
        }
        origin=code 
    }
    if !strings.HasPrefix(destination,"s"){
        code ,found:=findCode(destination)
        if !found{
            c.JSON(http.StatusBadRequest ,gin.H{"error":fmt.Sprintf("Не найден код для станции %s",destination)})
            return 
        }
        destination=code 
    }

    directRoute ,err:=getSchedule(origin,destination,date)
    if err!=nil{
        c.JSON(http.StatusInternalServerError ,gin.H{"error":err.Error()})
        return 
    }

    // Проверка наличия ключа 'threads' — признак наличия прямого маршрута
    if threadsData ,exists:=directRoute["threads"]; exists && threadsData!=nil{
        c.JSON(http.StatusOK ,gin.H{"direct_route":directRoute})
    }else{
        c.JSON(http.StatusOK ,
            gin.H{"message":"Прямой маршрут не найден. Функционал построения маршрутов с пересадками в разработке."})
    }
}

/**
 * complexRouteHandler Обрабатывает запросы на поиск маршрута с пересадкой через промежуточную станцию,
 * сохраняет результат в файл с уникальным именем.
 *
 * Извлекает параметры: origin , transfer , destination , date. Выполняет два вызова getSchedule —
 * от начальной до промежуточной и от промежуточной до конечной станции. Формирует итоговый результат,
 * сохраняет его в файл с уникальным именем и возвращает клиенту.
 *
 * Контекст:
 *     c (*gin.Context): объект контекста Gin для обработки запроса и формирования ответа.
 */
func complexRouteHandler(c *gin.Context) {
    origin := c.Query("origin")
    transfer := c.Query("transfer")
	destination := c.Query("destination")
	date := c.Query("date")

	if origin == "" || transfer == "" || destination == "" || date == "" {
	    c.JSON(http.StatusBadRequest,
	        gin.H{"error": "Необходимо указать origin , transfer , destination и date"})
	    return
    }

    // Получение маршрута от начальной до промежуточной станции
	leg1 ,err1:=getSchedule(origin , transfer , date)
	if err1 !=nil{
	    c.JSON(http.StatusInternalServerError,
	        gin.H{"error": fmt.Sprintf("Ошибка при получении маршрута от %s до %s: %v", origin , transfer ,err1)})
	    return 
	    }

    // Получение маршрута от промежуточной до конечной станции
	leg2 ,err2:=getSchedule(transfer , destination , date)
	if err2 !=nil{
	    c.JSON(http.StatusInternalServerError,
	        gin.H{"error": fmt.Sprintf("Ошибка при получении маршрута от %s до %s: %v", transfer , destination ,err2)})
	    return 
	    }

    // Формирование итогового результата - карты с сегментами маршрутов и временем запроса
	result:=map[string]interface{}{
	    "leg1": leg1,
	    "leg2": leg2,
	    "requested_at": time.Now().Format(time.RFC3339),
	    }

     // Маршалинг результата в формат JSON с отступами для читаемости файла
	resultJSON ,err:=json.MarshalIndent(result," ", "  ")
	if err!=nil{
	    log.Printf("Ошибка маршалинга результата: %v",err)
	    }else{
	        // Генерация уникального имени файла по текущему времени (чтобы избежать перезаписи)
	        timestamp:=time.Now().Format("20060102T150405")
	        fileName:=fmt.Sprintf("complex_route_result_%s.json", timestamp)

	        // Запись результата в файл
	        err= ioutil.WriteFile(fileName,resultJSON ,0644)
	        if err!=nil{
	            log.Printf("Ошибка записи в файл %s: %v",fileName ,err)
	        }else{
	            log.Printf("Результат сохранён в файл %s",fileName )
	        }
	        fmt.Println(string(resultJSON))
	    }

	c.JSON(http.StatusOK,result)// Отправка итогового результата клиенту в формате JSON
}

/**
 * main — точка входа программы. Загружает данные о станциях и запускает сервер Gin на порту 8080.
 *
 * Процесс:
 *   - Загружает данные о станциях из файла через loadStationCodes().
 *   - Создаёт роутеры для API, определяет маршруты и обработчики.
 *   - Запускает HTTP-сервер Gin на порту 8080, чтобы принимать входящие запросы.
 */

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
