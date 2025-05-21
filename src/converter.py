import json


def load_json(filename):
    with open(filename, 'r', encoding='utf-8') as file:
        return json.load(file)


def find_code(data):
    station_codes = {}
    if data["countries"]:
        counties = data["countries"]
        for item in counties:
            regions = item["regions"]
            for elem in regions:
                settlements = elem["settlements"]
                for settle in settlements:
                    stations = settle["stations"]
                    for station in stations:
                        if station["title"] not in station_codes.keys():
                            station_codes[station["title"]] = []
                        station_codes[station["title"]].append(
                            station["codes"]["yandex_code"])
    return station_codes

def detect_code(json_filename):
    json_data = load_json(json_filename)
    station_codes = find_code(json_data)
    with open('codes.json', 'w', encoding='utf-8') as file:
        json.dump(station_codes, file, ensure_ascii=False, indent=4)


json_filename = 'rasp.json'
detect_code(json_filename)
