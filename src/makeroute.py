import json
from datetime import datetime, timedelta
import requests
import os


class SimpleRoute:
    def __init__(self, s, f, st, ft, tr):
        self.start = s
        self.finish = f
        self.start_time = datetime.fromisoformat(st[0] + ' ' + st[1])
        self.finish_time = datetime.fromisoformat(ft[0] + ' ' + ft[1])
        self.transfers = tr
        self.duration = self.finish_time - self.start_time

    def __str__(self):
        return f'Откуда: {self.start}, Куда: {self.finish}, Отправление: {self.start_time}, Прибытие:{self.finish_time}, Пересадки: {self.transfers}, Длительность: {self.duration}'

class ComplexRoute:
    def __init__(self, way, time):
        self.path = way
        self.whole_duration = time

    def __str__(self):
        for elem in self.path:
            print(elem)
        return f'Общая длительность: {self.whole_duration}'

def calculate_time_difference(datetime1, datetime2):
    if datetime1 > datetime2:
        time_difference = timedelta(minutes=0)
    else:
        time_difference = datetime2 - datetime1
    return time_difference


rout_segments = []
cities = {}
while True:
    if len(cities) == 0:
        city = input("Введите город: ")
        date = input("Введите дату отправления в формате ГГГГ-ММ-ДД: ")
        cities[city] = date
    else:
        city = input("Введите следующую точку машрута или пустую строку для выхода: ")  
        if city == '':
            break
        else:
            date = input("Введите дату отправления в формате ГГГГ-ММ-ДД: ")
            cities[city] = date 
            if date == '':
                break  

sources = []
city_list = list(cities.keys())

for i, city in enumerate(city_list):
    if i < len(cities) - 1:
        response = requests.get(f'http://localhost:8080/api/schedule?from_station={city}&to_station={city_list[i+1]}&date={cities[city]}&transfers=true')
        with open(f'{i}.json', 'w', encoding='utf-8') as json_file:
                json.dump(response.json(), json_file, ensure_ascii=False, indent=4)
        sources.append(f'{i}.json')
    
i = 1
route_list = {}
for source in sources:
    with open(source, 'r', encoding='utf-8') as file:
        data = json.load(file)

    route_list[f'seg{i}'] = []

    if "segments" in data.keys():
        routes = data["segments"]
        for route in routes:
            if "from" in route.keys():
                start_point = route["from"]["title"]
            else:
                start_point = route["departure_from"]["title"]
            if "to" in route.keys():
                finish_point = route["to"]["title"]
            else:
                finish_point = route["arrival_to"]["title"]
            start_time = route["departure"]
            finish_time = route["arrival"]
            transfers = []
            if "details" in route.keys():
                for transfer in route["details"]:
                    if "is_transfer" in transfer.keys():
                        transfers.append(transfer["transfer_point"]["title"])
            new_route = SimpleRoute(start_point, finish_point, start_time.split(sep='T'), finish_time.split(sep='T'), transfers)
            route_list[f'seg{i}'].append(new_route)

        i += 1


complex_routes = []
for item in route_list["seg1"]:
    complex_routes.append([item])

route_list.pop("seg1")

for route in complex_routes:
    for key in route_list.keys():
        for new_route in route_list[key]:
            if calculate_time_difference(route[-1].finish_time, new_route.start_time) > timedelta(minutes=15):
                complex_routes.append([*route, new_route])

complex_routes_res = []
for item in complex_routes:
    if len(item) == len(sources):
        complex_routes_res.append(item)

complex_routes_final = []
for route in complex_routes_res:
    route_duration = route[-1].finish_time - route[0].start_time 
    complex_routes_final.append(ComplexRoute(route, route_duration))

#Сортировка маршрутов по времени в пути. В итоговом варианте приложения будет одним из параметров. 
sorted_routes = sorted(complex_routes_final, key=lambda x: x.whole_duration)

n = 1
for item in sorted_routes:
    print(f"Маршрут {n}:")
    n += 1
    print(item)

for filename in sources:
    if os.path.exists(filename):
        os.remove(filename)