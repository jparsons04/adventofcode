reports = []

with open('day2-input.txt', 'r') as f:
    for line in f:
        reports.append(line.strip())

safe_reports = 0


def level_safety_check(lvl1, lvl2, increasing):
    if (lvl2 > lvl1 and increasing is True) \
            or (lvl2 < lvl1 and increasing is False):
        if abs(lvl2 - lvl1) >= 1 and abs(lvl2 - lvl1) <= 3:
            return True
        else:
            return False
    else:
        return False


for i in range(len(reports)):
    levels = [int(x) for x in reports[i].split()]
    safe = True
    increasing = False
    cur_level = None
    previous_level = None
    for j in range(len(levels)):
        cur_level = levels[j]
        if j == 0:
            pass
        elif j == 1:
            increasing = cur_level > previous_level
            safe = safe \
                and level_safety_check(previous_level, cur_level, increasing)
        else:
            cur_level = levels[j]
            safe = safe \
                and level_safety_check(previous_level, cur_level, increasing)
        previous_level = cur_level
    if safe is True:
        safe_reports += 1
    safe = True

print(safe_reports)
