#!/usr/bin/env python3
"""Build packs/mylands/pack.json deterministically (ML-D02, ML-D03, ML-D04).

Inputs : docs/reference/mylands/tables-dom.json, docs/reference/mylands/units-roster.csv
Output : packs/mylands/pack.json
Rules  : a blank or zero source cell stays absent; no value is invented.
         Crystal quantities follow docs/modernization/CRYSTAL_ECONOMY.md (PROPOSED).
Usage  : python3 scripts/gen_mylands_pack.py [--check]   (--check fails if the file is stale)
"""
import csv
import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
SRC = ROOT / "docs/reference/mylands/tables-dom.json"
ROSTER = ROOT / "docs/reference/mylands/units-roster.csv"
OUT = ROOT / "packs/mylands/pack.json"

RACES = [  # id, ru name, faction, qualitative description (docs/reference/mylands/ECONOMY_RACES_BUILDINGS.md 1.1)
    ("knights", "Рыцари", "light", "Золото, вместимость носильщиков и складов, повышенная добыча злаков."),
    ("elves", "Эльфы", "light", "Максимальная добыча дерева; особенность лекарей."),
    ("wizards", "Волшебники", "light", "Злаки и железо; половина затрат маны; сопротивление героев заклинаниям."),
    ("elementals", "Элементали", "light", "Ускоренная наука, защитные особенности; обитают на Священной земле."),
    ("demons", "Демоны", "dark", "Камень; Демонесса имеет наибольшую атаку среди наёмников."),
    ("dark_elves", "Тёмные Эльфы", "dark", "Дерево и железо; ускоренное обучение магов, противодействие башням."),
    ("undead", "Нежить", "dark", "Дерево и камень; оживление павших монстров."),
    ("chimeras", "Химеры", "dark", "Подземные владения, добыча приисков и шахт, отсутствие подземного штрафа движения."),
]
RACE_STEMS = [  # order matters: "Тёмных Эльфов" before "Эльф"
    ("Тёмных Эльфов", "dark_elves"), ("Тёмные Эльфы", "dark_elves"), ("Рыцар", "knights"),
    ("Волшебник", "wizards"), ("Элементал", "elementals"), ("Эльф", "elves"), ("Нежит", "undead"),
    ("Демон", "demons"), ("Химер", "chimeras"),
]
RACE_NAME_ID = {n: i for i, n, _, _ in RACES}
ECON = ["economic", "tactical"]  # "Экономический, тактический сервер"

RESOURCES = [  # id, ru, kind, capped, tradable, lootable
    ("gold", "Золото", "basic", True, True, True),
    ("wood", "Дерево", "basic", True, True, True),
    ("stone", "Камень", "basic", True, True, True),
    ("iron", "Железо", "basic", True, True, True),
    ("grain", "Зерно", "food", True, True, True),
    ("pearl", "Чёрный жемчуг", "premium", False, False, False),
    ("people_free", "Свободное население", "basic", True, False, False),
    ("people_workers", "Рабочие и учёные", "basic", True, False, False),
    ("people_builders", "Строители", "basic", True, False, False),
]
COST_COLS = {"[gold]": "gold", "[wood]": "wood", "[stone]": "stone", "[iron]": "iron", "[pearl]": "pearl"}

CRYSTAL_A = ["red", "green", "blue", "violet", "yellow"]  # tier 1, 5 per icon (PROPOSED)
CRYSTAL_B = ["orange", "white", "emerald", "amethyst", "red_heart"]  # tier 2, 10 per icon (PROPOSED)
CRYSTAL_RU = {"red": "Красный", "green": "Зелёный", "blue": "Синий", "violet": "Фиолетовый", "yellow": "Жёлтый",
              "orange": "Оранжевый", "white": "Белый", "emerald": "Изумрудный", "amethyst": "Аметистовый",
              "red_heart": "Красное сердце"}

# page -> (building id, ru name, category)
BUILDINGS = {
    "City_hall": ("city_hall", "Центральное здание", "core"),
    "Houses": ("houses", "Жилища", "core"),
    "Warehouse": ("warehouse", "Склад", "storage"),
    "Market": ("market", "Рынок", "economy"),
    "Shop": ("shop", "Торговая лавка", "economy"),
    "Windmill": ("windmill", "Мельница", "economy"),
    "Samwill": ("samwill", "Хижина дровосека", "economy"),
    "Mine_gold": ("mine_gold", "Прииск", "economy"),
    "Mine_iron": ("mine_iron", "Рудник", "economy"),
    "Mine_stone": ("mine_stone", "Каменоломня", "economy"),
    "Alchemist": ("alchemist", "Лаборатория алхимиков", "science"),
    "Recovery": ("recovery", "Чаша изобилия", "special"),
    "Prison": ("prison", "Темница", "special"),
    "Tavern": ("tavern", "Таверна", "special"),
    "Artefact_store": ("artefact_store", "Сокровищница", "storage"),
    "Temple": ("temple", "Храм", "special"),
    "Terraformer": ("terraformer", "Терраформер", "special"),
    "Resurrector": ("resurrector", "Воскреситель", "military"),
    "Colony": ("colony", "Центр колонизации", "special"),
    "Exped": ("exped", "Экспедиционный корпус", "special"),
    "Fortification": ("fortification", "Укрепление", "military"),
    "Watch_tower": ("watch_tower", "Башня провидцев", "special"),
    "Antiwatch_tower": ("antiwatch_tower", "Башня зеркал", "special"),
    "Spy_guild": ("spy_guild", "Гильдия шпионов", "special"),
    "Antispy_guild": ("antispy_guild", "Гильдия антишпионов", "special"),
    "Merch_guild": ("merch_guild", "Гильдия купцов", "special"),
    "Scavenger_build": ("scavenger_build", "Дом вассалов", "military"),
    "Melee_build": ("melee_build", "Казарма", "military"),
    "Cavalry_build": ("cavalry_build", "Конюшня", "military"),
    "Flyer_build": ("flyer_build", "Бастион пегасов", "military"),
    "Range_build": ("range_build", "Стрельбище", "military"),
    "Healer_build": ("healer_build", "Монастырь", "military"),
    "Miner_build": ("miner_build", "Орден наёмников", "military"),
    "Mage_build": ("mage_build", "Школа магии", "military"),
    "Z_merch_guild": ("merch_guild_kz", "Гильдия купцов (замок клана)", "special"),
}
KZ = "(замок клана)"
BUILDINGS_KZ = {  # kz page -> base page
    "Warehouse_kz": "Warehouse", "Houses_kz": "Houses", "Fortification_kz": "Fortification",
    "Resurrector_kz": "Resurrector", "Antiwatch_tower_kz": "Antiwatch_tower", "Antispy_guild_kz": "Antispy_guild",
    "Scavenger_build_kz": "Scavenger_build", "Melee_build_kz": "Melee_build", "Cavalry_build_kz": "Cavalry_build",
    "Flyer_build_kz": "Flyer_build", "Range_build_kz": "Range_build", "Healer_build_kz": "Healer_build",
    "Miner_build_kz": "Miner_build", "Mage_build_kz": "Mage_build",
}
# (page, table name) -> (building id, ru name, category) for pages with two tables
MULTI = {
    ("Magic_Tower", "простая"): ("tower", "Башня", "military"),
    ("Magic_Tower", "магическая"): ("magic_tower", "Магическая башня", "military"),
    ("Magic_Tower_kz", "простая"): ("tower_kz", f"Башня {KZ}", "military"),
    ("Magic_Tower_kz", "магическая"): ("magic_tower_kz", f"Магическая башня {KZ}", "military"),
    ("Clan_warehouse", "склад"): ("clan_warehouse", "Склад клана", "storage"),
    ("Clan_warehouse", "резервация"): ("clan_reservation", "Резервация клана", "storage"),
}
PROD_RES = {"City_hall": "wood", "Windmill": "grain", "Samwill": "wood", "Mine_gold": "gold",
            "Mine_iron": "iron", "Mine_stone": "stone", "Shop": "gold"}
ONLY_MAIN_TERRAIN = [  # id, ru, resource (docs Resources page: terrain table)
    ("mountains", "Горы", "stone", "#7a7a7a"), ("desert", "Пустыня", "gold", "#c9a85b"),
    ("forest", "Лес", "wood", "#3f6b3a"), ("steppe", "Степь", "grain", "#9db35a"),
    ("underground", "Подземелье", None, "#3a3040"),
]

UNKNOWN = [
    "Время постройки и улучшения зданий не указано в таблицах: timeSeconds отсутствует.",
    "Источники кристаллов в энциклопедии не названы; количества кристаллов - PROPOSED (CRYSTAL_ECONOMY.md).",
    "Атака и здоровье монстров не указаны в таблицах монстров: параметры attackMin, attackMax, health, upkeep отсутствуют.",
    "Рейтинг и имена монстров клановых замков не указаны (в ростере только монстры 1-3 уровня).",
    "Множители профилей не заданы источником: профили содержат только переопределения таблиц (economic, tactical, heroic) и описание различий.",
    "Центральное здание: таблица содержит уровни 1-5; уровни 6-9 отсутствуют.",
    "Центр колонизации: таблица содержит только уровень 1.",
    "Сокровищница: цены страниц за ЧЖ (20 страниц) не перенесены, схема пака не имеет для них поля.",
    "Нотация 'Раса:значение' не объяснена страницей; принята как переопределение по расе (решение D-14).",
    "Столбцы Чаши изобилия A, B, C, D приняты как золото, камень, железо, зерно по WORLD_DOMAINS_EXPLORATION W-25 (INFERRED).",
    "Столбцы 'Скорость тренировки, мин' и 'Очередь тренировки, юнитов' Школы магии приняты как trainSecondsPerUnit и orderLimit (INFERRED).",
    "Названия зданий Восстановление, Темница и других взяты из индекса Buildings; названия для _kz-вариантов составлены по шаблону.",
    "Веса и цвета террайнов заглушки (PROPOSED): источник их не задаёт.",
    "Бонусы рас заданы только текстом; процентные модификаторы не вводились.",
    "Население (свободное, рабочие, строители) заведено как ресурсы без лимитов в таблицах; поведение определяет движок.",
    "Мощность простой башни хранится как диапазон towerDamageMin/towerDamageMax; единица урона принята по D-15 (за раунд).",
    "Профили: полный базовый набор чисел и наследование параметров между восемью профилями не заданы страницей Serv (ML-D08); профили miner, combat, strategic, operational, faction_war не имеют переопределений.",
    "Границы paramDefs (0.1-10) и значение по умолчанию 1 - PROPOSED: источник не задаёт ни параметров бонусов, ни границ.",
    "Цены всех услуг ЧЖ не заданы: поле priceUnknown, цена в паке отсутствует (BlackPearl, ML-D09).",
    "Длительность премиума: BlackPearl называет семь дней, Premium не называет срок, часть защиты исключена для Войны фракций (CONFLICT); durationSeconds отсутствует.",
    "Длительность и разбиение на услуги для глобальных, локальных услуг и автоматизации героев не заданы: перечислены только эффекты; каждый эффект заведён отдельной услугой с неизвестными ценой и сроком.",
    "Группа vacation (отпуск) - PROPOSED: BlackPearl делит услуги на четыре группы, отпуск в них не назван; отдельная группа нужна, потому что премиум во время отпуска не замораживается.",
    "Какие мгновенные услуги запрещены под осадой, правила возврата списаний и соответствие 15/30/60 суток интервалам 1/6/12 месяцев по порядку (формулировка 'соответствующие') не уточнены.",
]

PROFILES = [  # id, ru name, differences confirmed by Serv (COMBAT_HEROES_SOCIAL section 7); no numbers
    ("miner", "Шахтёрский", "Иная плотность внешних владений и руин, защита и шкала развития относительно боевого."),
    ("combat", "Боевой", "База сравнения страницы Serv; ограничения званий уточняет Title."),
    ("economic", "Экономический", "Удвоенная добыча, отдельные условия героев и генерации копей."),
    ("tactical", "Тактический", "Нет жалования во владениях; отдельные добыча, храм, расходы миссий и лимит руин."),
    ("heroic", "Геройский", "Ранние легендарные герои, мгновенные услуги науки, ресурсов и воскрешения, иные покупки свитков."),
    ("strategic", "Стратегический", "Нет жалования во владениях; иные предметы и науки, ограничение храма."),
    ("operational", "Оперативный", "Изменены некоторые задания и награды; полный перечень не дан."),
    ("faction_war", "Война фракций", "Три участника атаки, запрет своей фракции, разрушение городов, конечная цель и завершение мира."),
]
DEFAULT_PROFILE = "combat"  # PROPOSED, TARGET_PROFILE A-01
PARAM_DEFS = ["production_multiplier", "upkeep_multiplier", "build_speed_multiplier", "train_speed_multiplier"]  # PROPOSED names and bounds

# Services from BlackPearl, Premium, Vacation. Prices and durations are UNKNOWN unless stated.
SERVICE_EFFECTS = {  # group -> [(id, ru)]
    "global": [("production_science", "Добыча и наука"), ("routes", "Маршруты"), ("experience", "Опыт"), ("loot", "Лут"),
               ("upkeep", "Жалование"), ("calculator", "Калькулятор"), ("favor", "Милость"), ("reputation", "Репутация")],
    "local": [("gates", "Ворота"), ("population", "Население"), ("queue_finish", "Завершение очередей"), ("protection", "Защита"),
              ("damage", "Повреждения"), ("architecture", "Архитектура"), ("building_layout", "Расположение построек")],
    "hero_automation": [("hero_automation", "Автоматизация героев")],
}
VACATION_DAYS = [(15, 1), (30, 6), (60, 12)]  # days, cooldown months (Vacation)
VACATION_LOCKOUT_SECONDS = 6 * 3600  # not purchasable 6 hours before an attack


def services():
    unknown = {"priceUnknown": True, "durationUnknown": True}
    out = [{"id": "premium", "group": "premium", "name": {"ru": "Премиум"}, "proposed": True,
            "renewal": "upgrade_replaces", "levels": [dict(unknown)]}]
    for group, items in SERVICE_EFFECTS.items():
        for sid, ru in items:
            sid = sid if group == "hero_automation" else f"{group}_{sid}"
            out.append({"id": sid, "group": group, "name": {"ru": ru}, "proposed": True,
                        "renewal": "upgrade_replaces", "levels": [dict(unknown)]})
    out.append({"id": "vacation", "group": "vacation", "name": {"ru": "Отпуск"}, "proposed": True,
                "params": {"attackLockoutSeconds": VACATION_LOCKOUT_SECONDS}, "renewal": "upgrade_replaces",
                "levels": [{"priceUnknown": True, "durationSeconds": d * 86400, "cooldownMonths": m} for d, m in VACATION_DAYS]})
    return out


def num(s):
    s = s.strip().replace(",", ".")
    f = float(s)
    return int(f) if f == int(f) else f


def dur_seconds(s):
    total = 0
    for n, u in re.findall(r"(\d+(?:,\d+)?)\s*(дн|день|час|мин|сек)", s):
        v = num(n)
        total += v * {"дн": 86400, "день": 86400, "час": 3600, "мин": 60, "сек": 1}[u]
    if total == 0:
        raise ValueError(f"cannot parse duration {s!r}")
    return int(total) if total == int(total) else total


def minutes(s):
    v = num(s) * 60
    return int(v) if v == int(v) else v


def races_in(h):
    out = []
    for stem, rid in RACE_STEMS:
        if stem in h:
            if rid not in out:
                out.append(rid)
            h = h.replace(stem, "")
    return out


def parse_crystals(cell, mult_a=5, mult_b=10):
    if not cell:
        return {}
    out = {}
    for tok in cell.strip("[]").split("+"):
        if tok in CRYSTAL_A:
            out[tok] = out.get(tok, 0) + mult_a
        elif tok in CRYSTAL_B:
            out[tok] = out.get(tok, 0) + mult_b
        else:
            raise ValueError(f"unknown crystal {tok}")
    return out


class Eff:
    """Accumulates effects with per-race and per-profile layers."""

    def __init__(self):
        self.base = {}  # profile or None -> name -> {"base": v, "races": {}}

    def add(self, name, value, races, profiles):
        for prof in (profiles or [None]):
            slot = self.base.setdefault(prof, {}).setdefault(name, {"races": {}})
            if races:
                for r in races:
                    slot["races"][r] = value
            else:
                slot["base"] = value

    @staticmethod
    def render(layer):
        out = {}
        for name, s in layer.items():
            if "base" not in s:
                raise ValueError(f"effect {name} has race values without a base")
            out[name] = {"base": s["base"], "races": s["races"]} if s["races"] else s["base"]
        return out

    def effects(self):
        return self.render(self.base.get(None, {}))

    def overrides(self):
        res = {}
        for prof, layer in self.base.items():
            if prof is not None:
                res[prof] = self.render(layer)
        return res


def one(name, parser):
    return lambda h, c: [(name, parser(c))]


def resurrect(name, disc):
    def f(h, c):
        m = re.fullmatch(r"(\d+)(?: \(-(\d+)% от стоимости воскрешения\))?", c)
        if not m:
            raise ValueError(c)
        out = [(name, int(m.group(1)))]
        if m.group(2):
            out.append((disc, int(m.group(2))))
        return out
    return f


def rng(lo, hi):
    def f(h, c):
        a, b = [num(x) for x in c.split("-")]
        return [(lo, a), (hi, b)]
    return f


def prod_rules(page):
    res = PROD_RES.get(page)
    rules = [(r"добыв|добыч", one(f"{res}PerHour", num))] if res else []
    return rules


RULES = [
    (r"^Количество науки|^У Элементалей, в час", one("sciencePerHour", num)),
    (r"^Жилых мест", one("housingCapacity", num)),
    (r"^Прирост населения", one("populationGrowth", num)),
    (r"вместимость, ресурсов", one("resourceCapacity", num)),
    (r"^Вместимость, населения", one("populationCapacity", num)),
    (r"^Вместимость, предметов", one("itemCapacity", num)),
    (r"^Вместимость$", one("marketCapacity", num)),
    (r"^Лимит милости богов", one("mercyLimit", num)),
    (r"^Прирост милости богов", one("mercyGainPerHour", num)),
    (r"^Необходимо строителей", one("buildersRequired", num)),
    (r"^Воскрешает,%", resurrect("resurrectPct", "resurrectCostDiscountPct")),
    (r"^Количество исследователей", one("explorerCount", num)),
    (r"^Время тренировки исследователя, мин", one("explorerTrainSeconds", minutes)),
    (r"^Убивает", one("kills", num)),
    (r"^Бонус к защите", one("garrisonDefenseBonus", lambda c: num(c.lstrip("+")))),
    (r"^Время тренировки поселенца", one("settlerTrainSeconds", dur_seconds)),
    (r"^Количество поселенцев", one("settlerCount", num)),
    (r"^Время терраформирования", one("terraformSeconds", dur_seconds)),
    (r"^Урон обычной башни", rng("towerDamageMin", "towerDamageMax")),
    (r"^Урон магической башни", one("magicTowerDamageFactor", lambda c: num(c.split("*")[0]))),
    (r"^Количество шпионов", one("spyCount", num)),
    (r"^Время тренировки шпиона", one("spyTrainSeconds", dur_seconds)),
    (r"^Количество антишпионов", one("antispyCount", num)),
    (r"^Время тренировки антишпиона, мин", one("antispyTrainSeconds", minutes)),
    (r"^Количество зеркал", one("mirrorCount", num)),
    (r"^Время тренировки зеркала", one("mirrorTrainSeconds", dur_seconds)),
    (r"^Количество провидцев", one("seerCount", num)),
    (r"^Время тренировки провидца, сек", one("seerTrainSeconds", num)),
    (r"^Количество купцов", one("merchantCount", num)),
    (r"^Время тренировки купца, мин", one("merchantTrainSeconds", minutes)),
    (r"^Время тренировки купца$", one("merchantTrainSeconds", dur_seconds)),
    (r"^Ожидание после открытия туннеля", one("tunnelWaitSeconds", dur_seconds)),
    (r"^Лимит заказа|^Очередь тренировки", one("orderLimit", num)),
    (r"^Время тренировки 1 юнита, мин|корость тренировки, мин", one("trainSecondsPerUnit", minutes)),
    (r"^Время тренировки 1 юнита$", one("trainSecondsPerUnit", dur_seconds)),
    (r"^Вмещает пленных героев", one("prisonerCapacity", num)),
    (r"^Коэффициент шанса пленения", one("captureChanceFactor", num)),
    (r"^Восстанавливает, в час \(A\)", one("goldRecoveryPerHour", num)),
    (r"^Восстанавливает, в час \(B\)", one("stoneRecoveryPerHour", num)),
    (r"^Восстанавливает, в час \(C\)", one("ironRecoveryPerHour", num)),
    (r"^Восстанавливает, в час \(D\)", one("grainRecoveryPerHour", num)),
    (r"^Увеличение сбора с населения", one("populationCollectionFactor", num)),
]


def classify(page, header):
    """Return (rule function, races, profiles) for one table column header."""
    h = header
    profiles = []
    if h.startswith("Экономический, тактический сервер"):
        profiles, h = list(ECON), h[len("Экономический, тактический сервер"):].strip()
    elif h.startswith("Героический сервер, "):
        profiles, h = ["heroic"], h[len("Героический сервер, "):].strip()
        h = h[0].upper() + h[1:]
    races = races_in(h)
    plain = h
    for stem, _ in RACE_STEMS:
        plain = plain.replace(stem, "")
    for fn_re, fn in prod_rules(page) + RULES:
        if re.search(fn_re, h, re.I) or re.search(fn_re, plain, re.I):
            return fn, races, profiles
    raise ValueError(f"{page}: unmapped header {header!r}")


NOTE_RULES = [  # numeric facts in the 'Дополнительный эффект' text
    (r"Бесплатно создаёт (\d+) юнит", "freeUnitsPerDay"),
    (r"([+-]\d+)% к цене выкупа героя", "ransomPricePct"),
    (r"Расхищение переполненных складов (-?\d+)%", "overflowPlunderPct"),
    (r"Добыча на любом террайне как в [^(]+\(\+(\d+)%\)", "anyTerrainBonusPct"),
    (r"порцию \S+ до (\d+)", "weeklyRewardMax"),
    (r"Прирост \+(\d+) в каждом другом", "growthPerOtherHouse"),
    (r"\+(\d+) жилых мест за каждый другой", "housingPerOtherHouse"),
]


def note_effects(text):
    out = []
    for rx, name in NOTE_RULES:
        m = re.search(rx, text)
        if m:
            out.append((name, num(m.group(1).lstrip("+"))))
    if "Не вырабатывает залежи" in text:
        out.append(("noCellDeposit", 1))
    return out


def building_table(page, bid, rows, strings):
    head, body = rows[0], rows[1:]
    cols = []
    for i, h in enumerate(head):
        if h == "Уровень" or h in COST_COLS or h in ("Кристаллы", "Дополнительный эффект"):
            continue
        cols.append((i, h, classify(page, h)))
    levels = []
    for n, row in enumerate(body, 1):
        if int(row[0]) != n:
            raise ValueError(f"{page}: level {row[0]} at row {n}")
        cost = {}
        crystals = {}
        eff = Eff()
        for i, h in enumerate(head):
            c = row[i]
            if h in COST_COLS:
                if c and num(c) != 0:
                    cost[COST_COLS[h]] = int(num(c))
            elif h == "Кристаллы":
                crystals = parse_crystals(c)
            elif h == "Дополнительный эффект" and c:
                strings[f"building.{bid}.{n}.note"] = {"ru": c}
                for name, v in note_effects(c):
                    eff.add(name, v, [], [])
        for i, h, (fn, races, profiles) in cols:
            if row[i] == "":
                continue
            for name, v in fn(h, row[i]):
                eff.add(name, v, races, profiles)
        lv = {"level": n}
        if cost:
            lv["cost"] = cost
        if crystals:
            lv["crystals"] = crystals
        e = eff.effects()
        if e:
            lv["effects"] = e
        o = eff.overrides()
        if o:
            lv["overrides"] = o
        levels.append(lv)
    return {"building": bid, "levels": levels}


def split_races(cell):
    """'50 (Рыцари:100)' -> ('50', {'knights': '100'})."""
    m = re.fullmatch(r"(.*?)\s*\((.+)\)", cell)
    if not m:
        return cell.strip(), {}
    rs = {}
    for part in m.group(2).split(";"):
        name, val = part.split(":")
        rs[RACE_NAME_ID[name.strip()]] = val.strip()
    return m.group(1).strip(), rs


def mk(base, races):
    return {"base": base, "races": races} if races else base


def attack_pair(cell):
    base, rs = split_races(cell)
    def pair(s):
        parts = [num(x) for x in s.split("-")]
        return (parts[0], parts[-1])
    lo, hi = pair(base)
    rlo = {r: pair(v)[0] for r, v in rs.items()}
    rhi = {r: pair(v)[1] for r, v in rs.items()}
    return mk(lo, rlo), mk(hi, rhi)


def unit_defs(tables, roster):
    names = {}  # (cat, group, level) -> {race: ru}
    for r in roster:
        group = "monster" if r["race"] == "Монстры" else "player"
        key = RACE_NAME_ID.get(r["race"], "monsters")
        names.setdefault((r["category"], group, int(r["level"])), {})[key] = {"ru": r["name_ru"]}
    groups = {"игроки": ("", "player"), "монстры": ("_monster", "monster"), "замок": ("_castle", "castle")}
    out = []
    for tname, rows in tables.items():
        prefix, cat_ru = tname.split("-")
        suffix, group = groups[prefix]
        head = rows[0]
        cat = head[0].strip("[]")
        for n, row in enumerate(rows[1:], 1):
            if int(row[0]) != n:
                raise ValueError(f"{tname}: level {row[0]} at row {n}")
            params, cost, econ = {}, {}, {}
            for h, c in list(zip(head, row))[1:]:  # column 0 is the level
                if c == "":
                    continue
                target = params
                if h.startswith("Экономический сервер"):
                    h, target = h[len("Экономический сервер"):], econ
                if h == "[attack]":
                    lo, hi = attack_pair(c)
                    target["attackMin"], target["attackMax"] = lo, hi
                elif h == "[health]":
                    target["health"] = num(c)
                elif h == "Жалование":
                    params["upkeep"] = num(c)
                elif h in ("[gold]", "[wood]", "[grain]", "[iron]"):
                    if num(c) != 0:
                        cost[h[1:-1]] = int(num(c))
                elif h == "[people_all]":
                    target["population"] = num(c)
                elif h == "Рейтинг":
                    params["rating"] = num(c)
                elif h in ("Вместимость", "Воскрешает") or h.startswith("Уменьшает урон"):
                    pname = {"Вместимость": "capacity", "Воскрешает": "resurrectCount"}.get(h, "towerReduction")
                    base, rs = split_races(c)
                    params[pname] = mk(num(base), {r: num(v) for r, v in rs.items()})
                else:
                    raise ValueError(f"{tname}: unmapped header {h!r}")
            u = {"id": f"{cat}{suffix}_{n}", "category": cat, "level": n, "params": params}
            nm = names.get((cat, group, n))
            if nm:
                u["names"] = nm
            if cost:
                u["cost"] = cost
            if econ:
                u["overrides"] = {"economic": {"params": econ}}
            out.append(u)
    return out


def build():
    data = json.loads(SRC.read_text(encoding="utf-8"))["pages"]
    roster = list(csv.DictReader(ROSTER.open(encoding="utf-8")))
    strings = {}
    buildings, tables = [], []

    def add(page, tname, bid, ru, cat):
        t = building_table(page, bid, data[page][tname], strings)
        tables.append(t)
        buildings.append({
            "id": bid, "name": {"ru": ru}, "category": cat, "maxLevel": len(t["levels"]),
            "desc": {"ru": f"Источник: ru.mlgame.org/wiki-ru, страница {page} (DOM_EXTRACT, 2026-11-01)."},
            "cost": {"base": {}, "growth": 0}, "time": {"baseSeconds": 0, "growth": 0}, "score": 0,
        })

    for page, (bid, ru, cat) in BUILDINGS.items():
        add(page, "main", bid, ru, cat)
    for page, base in BUILDINGS_KZ.items():
        bid, ru, cat = BUILDINGS[base]
        add(page, "main", bid + "_kz", f"{ru} {KZ}", cat)
    for (page, tname), (bid, ru, cat) in MULTI.items():
        add(page, tname, bid, ru, cat)
    order = {b["id"]: i for i, b in enumerate(buildings)}  # keep insertion order (deterministic)
    assert len(order) == len(buildings)

    for cid in CRYSTAL_A + CRYSTAL_B:
        strings[f"crystal.{cid}"] = {"ru": CRYSTAL_RU[cid]}
    for i, text in enumerate(UNKNOWN, 1):
        strings[f"unknown.{i:02d}"] = {"ru": f"UNKNOWN: {text}"}
    strings["proposed.crystals"] = {"ru": "PROPOSED: 5 единиц на значок ступени A, 10 на значок ступени B (CRYSTAL_ECONOMY.md)."}
    strings["proposed.geometry"] = {"ru": "PROPOSED: империя 90 x 30 королевств, королевство 9 x 9 полей."}
    strings["factions.light"] = {"ru": "Светлая фракция"}
    strings["factions.dark"] = {"ru": "Тёмная фракция"}

    unit_tables = {k: v for k, v in data["UnitParametrs"].items()}
    units = unit_defs(unit_tables, roster)

    return {
        "formatVersion": 1,
        "id": "mylands",
        "version": "0.1.0",
        "meta": {
            "title": {"ru": "My Lands"},
            "description": {"ru": "Игровые данные My Lands, перенесённые из таблиц энциклопедии ru.mlgame.org/wiki-ru. Пак содержит только данные."},
            "studio": "Realmforge",
            "license": "Числа перенесены из публичной энциклопедии для исследования; без чужой графики.",
        },
        "locales": ["ru"],
        "defaultLocale": "ru",
        "theme": {
            "palette": {"bg": "#14110f", "surface": "#221d19", "text": "#f1e8d8", "muted": "#9a8f80",
                        "accent": "#d9a441", "danger": "#c0392b", "ok": "#4f9d5d"},
            "assetsLicense": "none: пак содержит только данные, графики нет",
        },
        "params": {"foodResource": "grain", "premiumResource": "pearl", "mapWidth": 90, "mapHeight": 30},
        "resources": [
            {"id": i, "name": {"ru": n}, "kind": k, "capped": c, "tradable": t, "lootable": l}
            for i, n, k, c, t, l in RESOURCES
        ],
        "races": [
            {"id": i, "name": {"ru": n}, "desc": {"ru": d}, "modifiers": [], "faction": f}
            for i, n, f, d in RACES
        ],
        "buildings": buildings,
        "terrains": [
            {"id": i, "name": {"ru": n}, "weight": 1, "passable": True, "color": c, **({"resource": r} if r else {})}
            for i, n, r, c in ONLY_MAIN_TERRAIN
        ],
        "modes": ["standard"],
        "ui": {"screens": ["city", "army", "heroes", "map", "missions", "reports", "market", "alliance", "quests", "mail", "profile", "settings"]},
        "crystals": [{"id": c, "tier": 1} for c in CRYSTAL_A] + [{"id": c, "tier": 2} for c in CRYSTAL_B],
        "buildingTables": tables,
        "paramDefs": [{"id": p, "min": 0.1, "max": 10, "default": 1, "proposed": True} for p in PARAM_DEFS],
        "serverProfiles": [
            {"id": i, "name": {"ru": n}, "desc": {"ru": d}, "multipliers": {},
             **({"default": True, "proposed": True} if i == DEFAULT_PROFILE else {})}
            for i, n, d in PROFILES
        ],
        "services": services(),
        "unitDefs": units,
        "geometry": {"empireWidth": 90, "empireHeight": 30, "kingdomWidth": 9, "kingdomHeight": 9, "proposed": True},
        "strings": strings,
    }


def main():
    text = json.dumps(build(), ensure_ascii=False, indent=1) + "\n"
    if "--check" in sys.argv:
        if not OUT.exists() or OUT.read_text(encoding="utf-8") != text:
            sys.exit(f"{OUT} is stale; run scripts/gen_mylands_pack.py")
        return
    OUT.parent.mkdir(parents=True, exist_ok=True)
    OUT.write_text(text, encoding="utf-8")
    print(f"wrote {OUT}")


if __name__ == "__main__":
    main()
