'''
=======================================================================
                             Travel_API                               
                      CUP-IT-2025 Changellenge                        
                           Кейс: Axenix                               
                        Команда: Headliners                           
 													  		          
 																	  
 	Авторы:			 										  
 	 - Цабина Валерия (https://github.com/VTsabina)			      
 	 - Зайцев Данила (https://github.com/ZaitsevD12)			  
 	 - Постовалова Лада (https://github.com/Lada12345)		      
 	 - Миронович Игорь (https://github.com/steelemi987)		      
 																	  
 	© 2025 Все права защищены.						  
=======================================================================
'''

import json
from datetime import datetime, timedelta
import requests
import os

class SimpleRoute:
    """
    Класс для представления простого маршрута между двумя точками.
    Хранит информацию о начальной и конечной точке, времени отправления и прибытия,
    количестве пересадок и длительности маршрута.
    """
    def __init__(self, s, f, st, ft, tr):
        """
        Инициализация объекта SimpleRoute.
        
        Args:
            s (str): название начальной точки.
            f (str): название конечной точки.
            st (list): список строк с датой и временем отправления в формате [ГГГГ-ММ-ДД, ЧЧ:ММ:СС].
            ft (list): список строк с датой и временем прибытия в формате [ГГГГ-ММ-ДД, ЧЧ:ММ:СС].
            tr (list): список пересадок (названий станций).
        """
        self.start = s
        self.finish = f
        # Преобразование строк в объекты datetime
        self.start_time = datetime.fromisoformat(st[0] + ' ' + st[1])
        self.finish_time = datetime.fromisoformat(ft[0] + ' ' + ft[1])
        self.transfers = tr
        # Вычисление длительности маршрута
        self.duration = self.finish_time - self.start_time

    def __str__(self):
        """
        Строковое представление маршрута для вывода.
        """
        return (f'Откуда: {self.start}, Куда: {self.finish}, '
                f'Отправление: {self.start_time}, Прибытие: {self.finish_time}, '
                f'Пересадки: {self.transfers}, Длительность: {self.duration}')

class ComplexRoute:
    """
    Класс для представления сложного маршрута, состоящего из нескольких сегментов.
    Хранит путь (список SimpleRoute) и общую длительность.
    """
    def __init__(self, way, time):
        """
        Инициализация объекта ComplexRoute.
        
        Args:
            way (list): список объектов SimpleRoute — сегментов маршрута.
            time (timedelta): общая длительность маршрута.
        """
        self.path = way
        self.whole_duration = time

    def __str__(self):
        """
        Строковое представление сложного маршрута.
        Выводит все сегменты и общую длительность.
        """
        for elem in self.path:
            print(elem)
        return f'Общая длительность: {self.whole_duration}'

def calculate_time_difference(datetime1, datetime2):
    """
    Вычисляет разницу во времени между двумя объектами datetime.
    
    Args:
        datetime1 (datetime): первая дата/время.
        datetime2 (datetime): вторая дата/время.
        
    Returns:
        timedelta: разница времени. Если первая позже второй — возвращает 0.
    """
    if datetime1 > datetime2:
        return timedelta(minutes=0)
    else:
        return datetime2 - datetime1

def get_cities():
    """
    Запрашивает у пользователя последовательность городов и дат отправления для маршрутов.
    
    Returns:
        dict: словарь городов и соответствующих дат отправления.
    """
    cities = {}
    while True:
        if len(cities) == 0:
            # Ввод первого города и даты
            city = input("Введите город: ")
            date = input("Введите дату отправления в формате ГГГГ-ММ-ДД: ")
            cities[city] = date
        else:
            # Ввод следующего города или завершение
            city = input("Введите следующую точку маршрута или пустую строку для выхода: ")  
            if city == '':
                break
            else:
                date = input("Введите дату отправления в формате ГГГГ-ММ-ДД: ")
                cities[city] = date 
                if date == '':
                    break  
    return cities

def get_route_list(sources):
    """
    Загружает данные маршрутов из указанных файлов и создает список объектов SimpleRoute для каждого сегмента.
    
    Args:
        sources (list): список имён файлов с данными маршрутов.
        
    Returns:
        dict: словарь сегментов ('seg1', 'seg2', ...) с списками объектов SimpleRoute.
    """
    route_list = {}
    i = 1
    for source in sources:
        with open(source, 'r', encoding='utf-8') as file:
            data = json.load(file)

        route_list[f'seg{i}'] = []

        # Проверка наличия ключа "segments" в данных
        if "segments" in data.keys():
            routes = data["segments"]
            for route in routes:
                # Определение начальной точки
                if "from" in route.keys():
                    start_point = route["from"]["title"]
                else:
                    start_point = route["departure_from"]["title"]
                # Определение конечной точки
                if "to" in route.keys():
                    finish_point = route["to"]["title"]
                else:
                    finish_point = route["arrival_to"]["title"]
                # Время отправления и прибытия
                start_time = route["departure"]
                finish_time = route["arrival"]
                # Обработка пересадок, если есть
                transfers = []
                if "details" in route.keys():
                    for transfer in route["details"]:
                        if "is_transfer" in transfer.keys():
                            transfers.append(transfer["transfer_point"]["title"])
                # Создание объекта SimpleRoute и добавление его в список сегмента
                new_route = SimpleRoute(
                    start_point,
                    finish_point,
                    start_time.split(sep='T'),
                    finish_time.split(sep='T'),
                    transfers
                )
                route_list[f'seg{i}'].append(new_route)
            i += 1  # Переход к следующему сегменту файла
    return route_list

def main():
    """
    Основная функция программы. Выполняет сбор данных о городах,
    загрузку маршрутов, построение возможных путей и вывод результатов.
    Также удаляет временные файлы после обработки.
    """
    
    # Получение списка городов и дат от пользователя
    cities = get_cities()

    sources = []
    city_list = list(cities.keys())

    # Запрос данных у API для последовательных пар городов
    for i, city in enumerate(city_list):
        if i < len(cities) - 1:
            response = requests.get(
                f'http://localhost:8080/api/schedule?from_station={city}&to_station={city_list[i+1]}&date={cities[city]}&transfers=true'
            )
            filename = f'{i}.json'
            with open(filename, 'w', encoding='utf-8') as json_file:
                json.dump(response.json(), json_file, ensure_ascii=False, indent=4)
            sources.append(filename)

    # Получение списка маршрутов из загруженных файлов
    route_list = get_route_list(sources)

    # Формирование начальных путей — только первый сегмент каждого варианта маршрута
    complex_routes = []
    for item in route_list["seg1"]:
        complex_routes.append([item])

    # Удаление первого сегмента из общего списка для дальнейшего расширения путей
    # (чтобы не повторять его при объединении)
    
    route_list.pop("seg1")

    # Построение всех возможных путей с учетом времени пересадки (>15 минут)
    for route in complex_routes:
        for key in route_list.keys():
            for new_route in route_list[key]:
                if calculate_time_difference(route[-1].finish_time, new_route.start_time) > timedelta(minutes=15):
                    complex_routes.append([*route, new_route])

    # Отбор путей длиной равной количеству исходных точек — полные маршруты по всем городам
    complex_routes_res = []
    for item in complex_routes:
        if len(item) == len(sources):
             complex_routes_res.append(item)

    # Создание объектов ComplexRoute для каждого полного маршрута и вычисление общей длительности
    complex_routes_final = []
    for route in complex_routes_res:
        route_duration = route[-1].finish_time - route[0].start_time 
        complex_routes_final.append(ComplexRoute(route, route_duration))

    # Сортировка маршрутов по времени в пути (от меньшего к большему)
    sorted_routes = sorted(complex_routes_final, key=lambda x: x.whole_duration)

    # Вывод отсортированных маршрутов по порядковому номеру
    n = 1
    for item in sorted_routes:
        print(f"Маршрут {n}:")
        n += 1
        print(item)

    # Удаление временных файлов с данными о сегментах после обработки
    for filename in sources:
        if os.path.exists(filename):
            os.remove(filename)


if __name__ == '__main__':
    main()
