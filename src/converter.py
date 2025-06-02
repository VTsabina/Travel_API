import json

def load_json(filename):
    """
    Загружает JSON-данные из файла.
    
    Args:
        filename (str): Путь к файлу JSON.
        
    Returns:
        dict: Распарсенные данные из файла.
    """
    with open(filename, 'r', encoding='utf-8') as file:
        return json.load(file)

def find_code(data):
    """
    Проходит по структуре данных и собирает все уникальные коды станций по названиям.
    
    Args:
        data (dict): Структура данных, загруженная из JSON.
        
    Returns:
        dict: Словарь, где ключи — названия станций, а значения — списки кодов Yandex.
    """
    station_codes = {}
    
    # Проверка наличия ключа "countries" в данных
    if data["countries"]:
        # Перебор всех стран
        countries = data["countries"]
        for country in countries:
            # Перебор регионов внутри страны
            regions = country["regions"]
            for region in regions:
                # Перебор населённых пунктов (поселков) внутри региона
                settlements = region["settlements"]
                for settlement in settlements:
                    # Перебор станций внутри поселка
                    stations = settlement["stations"]
                    for station in stations:
                        station_name = station["title"]
                        # Инициализация списка кодов для станции, если её ещё нет в словаре
                        if station_name not in station_codes:
                            station_codes[station_name] = []
                        # Запись кода станции в список по её названию
                        station_codes[station_name].append(
                            station["codes"]["yandex_code"]
                        )
    return station_codes

def detect_code(json_filename):
    """
    Загружает данные из файла, извлекает коды станций и сохраняет их в новый JSON-файл.
    
    Args:
        json_filename (str): Имя файла с исходными данными.
    """
    # Загрузка исходных данные из файла
    json_data = load_json(json_filename)
    
    # Получение словарь станций и их кодов
    station_codes = find_code(json_data)
    
    # Запись результата в файл 'codes.json' с отступами для читаемости
    with open('codes.json', 'w', encoding='utf-8') as file:
        json.dump(station_codes, file, ensure_ascii=False, indent=4)

# Основной вызов функции для обработки файла 'rasp.json'
json_filename = 'rasp.json'
detect_code(json_filename)